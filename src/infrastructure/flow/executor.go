package flow

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/domains/flow"
	aiService "github.com/aldinokemal/go-whatsapp-web-multidevice/infrastructure/ai"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

// FlowExecutor executes flows
type FlowExecutor struct {
	flowRepo *SQLiteRepository
}

// NewFlowExecutor creates a new flow executor
func NewFlowExecutor(repo *SQLiteRepository) *FlowExecutor {
	return &FlowExecutor{flowRepo: repo}
}

// ExecutionContext holds the state during flow execution
type ExecutionContext struct {
	Variables   map[string]interface{}
	Input       map[string]interface{}
	Output      map[string]interface{}
	Credentials map[string]*flow.Credential
	CurrentNode string
	Flow        *flow.Flow
}

// Execute runs a flow with given input
func (e *FlowExecutor) Execute(ctx context.Context, f *flow.Flow, input map[string]interface{}) (map[string]interface{}, error) {
	if f == nil || len(f.Nodes) == 0 {
		return nil, fmt.Errorf("flow is empty")
	}

	// Build execution context
	execCtx := &ExecutionContext{
		Variables:   make(map[string]interface{}),
		Input:       input,
		Output:      make(map[string]interface{}),
		Credentials: make(map[string]*flow.Credential),
		Flow:        f,
	}

	// Load flow variables
	for _, v := range f.Variables {
		execCtx.Variables[v.Name] = v.Value
	}

	// Copy input to variables
	for k, v := range input {
		execCtx.Variables[k] = v
	}

	// Find trigger node (entry point)
	triggerNode := e.findTriggerNode(f.Nodes)
	if triggerNode == nil {
		return nil, fmt.Errorf("no trigger node found")
	}

	// Execute from trigger
	if err := e.executeFromNode(ctx, execCtx, triggerNode.ID); err != nil {
		return nil, err
	}

	return execCtx.Output, nil
}

func (e *FlowExecutor) findTriggerNode(nodes []flow.Node) *flow.Node {
	for i := range nodes {
		if strings.HasPrefix(nodes[i].Type, "trigger_") {
			return &nodes[i]
		}
	}
	return nil
}

func (e *FlowExecutor) executeFromNode(ctx context.Context, execCtx *ExecutionContext, nodeID string) error {
	node := e.findNode(execCtx.Flow.Nodes, nodeID)
	if node == nil {
		return fmt.Errorf("node %s not found", nodeID)
	}

	execCtx.CurrentNode = nodeID
	logrus.Debugf("Executing node: %s (%s)", node.Label, node.Type)

	// Execute current node
	output, err := e.executeNode(ctx, execCtx, node)
	if err != nil {
		return fmt.Errorf("node %s failed: %w", node.Label, err)
	}

	// Store output in variables
	if output != nil {
		for k, v := range output {
			execCtx.Variables[k] = v
		}
		execCtx.Output = output
	}

	// Find next nodes
	nextNodeIDs := e.findNextNodes(execCtx.Flow.Edges, nodeID, output)
	
	// Execute next nodes
	for _, nextID := range nextNodeIDs {
		if err := e.executeFromNode(ctx, execCtx, nextID); err != nil {
			return err
		}
	}

	return nil
}

func (e *FlowExecutor) findNode(nodes []flow.Node, id string) *flow.Node {
	for i := range nodes {
		if nodes[i].ID == id {
			return &nodes[i]
		}
	}
	return nil
}

func (e *FlowExecutor) findNextNodes(edges []flow.Edge, sourceID string, output map[string]interface{}) []string {
	var nextIDs []string
	for _, edge := range edges {
		if edge.Source == sourceID {
			// Check if edge has condition (for condition nodes)
			if edge.SourceHandle != "" && output != nil {
				// Only follow edges that match the output handle
				if result, ok := output["_handle"]; ok && result == edge.SourceHandle {
					nextIDs = append(nextIDs, edge.Target)
				}
			} else {
				nextIDs = append(nextIDs, edge.Target)
			}
		}
	}
	return nextIDs
}

// executeNode executes a single node
func (e *FlowExecutor) executeNode(ctx context.Context, execCtx *ExecutionContext, node *flow.Node) (map[string]interface{}, error) {
	switch node.Type {
	// Triggers just pass input through
	case flow.NodeTypeTriggerWhatsApp, flow.NodeTypeTriggerTelegram, flow.NodeTypeTriggerInstagram, flow.NodeTypeTriggerWebhook:
		return execCtx.Input, nil

	case flow.NodeTypeAIAgent:
		return e.executeAIAgent(ctx, execCtx, node)

	case flow.NodeTypeHTTPRequest:
		return e.executeHTTPRequest(ctx, execCtx, node)

	case flow.NodeTypeDatabase:
		return e.executeDatabase(ctx, execCtx, node)

	case flow.NodeTypeCondition:
		return e.executeCondition(ctx, execCtx, node)

	case flow.NodeTypeDelay:
		return e.executeDelay(ctx, execCtx, node)

	case flow.NodeTypeSendMessage:
		return e.executeSendMessage(ctx, execCtx, node)

	case flow.NodeTypeSetVariable:
		return e.executeSetVariable(ctx, execCtx, node)

	default:
		return execCtx.Variables, nil
	}
}

// === Node Executors ===

func (e *FlowExecutor) executeAIAgent(ctx context.Context, execCtx *ExecutionContext, node *flow.Node) (map[string]interface{}, error) {
	data := node.Data
	
	credentialID, _ := data["credential_id"].(string)
	model, _ := data["model"].(string)
	systemPrompt, _ := data["system_prompt"].(string)
	
	if model == "" {
		model = "gpt-4o-mini"
	}

	// Get API key from credential or data
	apiKey := ""
	if credentialID != "" {
		cred, err := e.flowRepo.GetCredentialByID(ctx, credentialID)
		if err == nil {
			var config flow.OpenAICredential
			json.Unmarshal([]byte(cred.Config), &config)
			apiKey = config.APIKey
		}
	}
	if apiKey == "" {
		apiKey, _ = data["api_key"].(string)
	}

	if apiKey == "" {
		return nil, fmt.Errorf("AI agent requires API key")
	}

	// Get user message from input
	userMessage, _ := execCtx.Variables["message"].(string)
	if userMessage == "" {
		userMessage, _ = execCtx.Variables["text"].(string)
	}

	if userMessage == "" {
		return nil, fmt.Errorf("no message to process")
	}

	// Interpolate variables in system prompt
	systemPrompt = e.interpolateVariables(systemPrompt, execCtx.Variables)

	// Call AI (Flow executor doesn't use SerpAPI, so pass empty string)
	ai := aiService.NewService(apiKey, "")

	// Use sensible defaults for flows; they are independent from per-agent settings
	const defaultMaxTokens = 500
	const defaultTemperature = 0.7

	response, err := ai.GenerateResponse(ctx, userMessage, systemPrompt, model, defaultMaxTokens, defaultTemperature)
	if err != nil {
		return nil, fmt.Errorf("AI generation failed: %w", err)
	}

	return map[string]interface{}{
		"response":    response,
		"ai_response": response,
		"message":     response,
	}, nil
}

func (e *FlowExecutor) executeHTTPRequest(ctx context.Context, execCtx *ExecutionContext, node *flow.Node) (map[string]interface{}, error) {
	data := node.Data

	method, _ := data["method"].(string)
	url, _ := data["url"].(string)
	bodyStr, _ := data["body"].(string)
	headersMap, _ := data["headers"].(map[string]interface{})
	timeoutSec, _ := data["timeout"].(float64)

	if method == "" {
		method = "GET"
	}
	if url == "" {
		return nil, fmt.Errorf("URL is required for HTTP request")
	}
	if timeoutSec == 0 {
		timeoutSec = 30
	}

	// Interpolate variables
	url = e.interpolateVariables(url, execCtx.Variables)
	bodyStr = e.interpolateVariables(bodyStr, execCtx.Variables)

	// Build request
	var body io.Reader
	if bodyStr != "" {
		body = bytes.NewBufferString(bodyStr)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	for k, v := range headersMap {
		if str, ok := v.(string); ok {
			req.Header.Set(k, e.interpolateVariables(str, execCtx.Variables))
		}
	}

	if bodyStr != "" && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	// Execute request
	client := &http.Client{Timeout: time.Duration(timeoutSec) * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Try to parse as JSON
	var jsonResp interface{}
	if err := json.Unmarshal(respBody, &jsonResp); err == nil {
		return map[string]interface{}{
			"status_code": resp.StatusCode,
			"body":        jsonResp,
			"headers":     resp.Header,
		}, nil
	}

	return map[string]interface{}{
		"status_code": resp.StatusCode,
		"body":        string(respBody),
		"headers":     resp.Header,
	}, nil
}

func (e *FlowExecutor) executeDatabase(ctx context.Context, execCtx *ExecutionContext, node *flow.Node) (map[string]interface{}, error) {
	data := node.Data

	credentialID, _ := data["credential_id"].(string)
	operation, _ := data["operation"].(string)
	table, _ := data["table"].(string)
	query, _ := data["query"].(string)

	if credentialID == "" {
		return nil, fmt.Errorf("database credential is required")
	}

	// Get database config
	cred, err := e.flowRepo.GetCredentialByID(ctx, credentialID)
	if err != nil {
		return nil, fmt.Errorf("credential not found: %w", err)
	}

	var dbConfig flow.DatabaseCredential
	if err := json.Unmarshal([]byte(cred.Config), &dbConfig); err != nil {
		return nil, fmt.Errorf("invalid database config: %w", err)
	}

	// Connect to database
	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		dbConfig.Host, dbConfig.Port, dbConfig.User, dbConfig.Password, dbConfig.Database, dbConfig.SSLMode,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}
	defer db.Close()

	// Execute based on operation
	switch operation {
	case "raw":
		query = e.interpolateVariables(query, execCtx.Variables)
		return e.executeRawSQL(ctx, db, query)
	case "select":
		return e.executeSelect(ctx, db, table, data, execCtx.Variables)
	case "insert":
		return e.executeInsert(ctx, db, table, data, execCtx.Variables)
	case "update":
		return e.executeUpdate(ctx, db, table, data, execCtx.Variables)
	case "delete":
		return e.executeDelete(ctx, db, table, data, execCtx.Variables)
	default:
		return nil, fmt.Errorf("unknown operation: %s", operation)
	}
}

func (e *FlowExecutor) executeRawSQL(ctx context.Context, db *sql.DB, query string) (map[string]interface{}, error) {
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, _ := rows.Columns()
	var results []map[string]interface{}

	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}

		row := make(map[string]interface{})
		for i, col := range columns {
			row[col] = values[i]
		}
		results = append(results, row)
	}

	return map[string]interface{}{
		"rows":  results,
		"count": len(results),
	}, nil
}

func (e *FlowExecutor) executeSelect(ctx context.Context, db *sql.DB, table string, data map[string]interface{}, vars map[string]interface{}) (map[string]interface{}, error) {
	columns := "*"
	if cols, ok := data["columns"].([]interface{}); ok && len(cols) > 0 {
		colStrs := make([]string, len(cols))
		for i, c := range cols {
			colStrs[i] = c.(string)
		}
		columns = strings.Join(colStrs, ", ")
	}

	query := fmt.Sprintf("SELECT %s FROM %s", columns, table)

	if where, ok := data["where"].(string); ok && where != "" {
		query += " WHERE " + e.interpolateVariables(where, vars)
	}

	if limit, ok := data["limit"].(float64); ok && limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", int(limit))
	}

	return e.executeRawSQL(ctx, db, query)
}

func (e *FlowExecutor) executeInsert(ctx context.Context, db *sql.DB, table string, data map[string]interface{}, vars map[string]interface{}) (map[string]interface{}, error) {
	values, ok := data["values"].(map[string]interface{})
	if !ok || len(values) == 0 {
		return nil, fmt.Errorf("values required for insert")
	}

	var cols []string
	var placeholders []string
	var args []interface{}
	i := 1

	for col, val := range values {
		cols = append(cols, col)
		placeholders = append(placeholders, fmt.Sprintf("$%d", i))
		// Interpolate string values
		if str, ok := val.(string); ok {
			args = append(args, e.interpolateVariables(str, vars))
		} else {
			args = append(args, val)
		}
		i++
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) RETURNING *",
		table, strings.Join(cols, ", "), strings.Join(placeholders, ", "))

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		columns, _ := rows.Columns()
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}
		rows.Scan(valuePtrs...)

		row := make(map[string]interface{})
		for i, col := range columns {
			row[col] = values[i]
		}
		return map[string]interface{}{"inserted": row}, nil
	}

	return map[string]interface{}{"inserted": true}, nil
}

func (e *FlowExecutor) executeUpdate(ctx context.Context, db *sql.DB, table string, data map[string]interface{}, vars map[string]interface{}) (map[string]interface{}, error) {
	values, ok := data["values"].(map[string]interface{})
	if !ok || len(values) == 0 {
		return nil, fmt.Errorf("values required for update")
	}

	where, _ := data["where"].(string)
	if where == "" {
		return nil, fmt.Errorf("where clause required for update")
	}

	var sets []string
	var args []interface{}
	i := 1

	for col, val := range values {
		sets = append(sets, fmt.Sprintf("%s = $%d", col, i))
		if str, ok := val.(string); ok {
			args = append(args, e.interpolateVariables(str, vars))
		} else {
			args = append(args, val)
		}
		i++
	}

	query := fmt.Sprintf("UPDATE %s SET %s WHERE %s",
		table, strings.Join(sets, ", "), e.interpolateVariables(where, vars))

	result, err := db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	affected, _ := result.RowsAffected()
	return map[string]interface{}{"rows_affected": affected}, nil
}

func (e *FlowExecutor) executeDelete(ctx context.Context, db *sql.DB, table string, data map[string]interface{}, vars map[string]interface{}) (map[string]interface{}, error) {
	where, _ := data["where"].(string)
	if where == "" {
		return nil, fmt.Errorf("where clause required for delete")
	}

	query := fmt.Sprintf("DELETE FROM %s WHERE %s", table, e.interpolateVariables(where, vars))

	result, err := db.ExecContext(ctx, query)
	if err != nil {
		return nil, err
	}

	affected, _ := result.RowsAffected()
	return map[string]interface{}{"rows_affected": affected}, nil
}

func (e *FlowExecutor) executeCondition(ctx context.Context, execCtx *ExecutionContext, node *flow.Node) (map[string]interface{}, error) {
	data := node.Data

	conditionsRaw, _ := data["conditions"].([]interface{})
	combineWith, _ := data["combine_with"].(string)
	if combineWith == "" {
		combineWith = "and"
	}

	results := make([]bool, 0)

	for _, condRaw := range conditionsRaw {
		cond, ok := condRaw.(map[string]interface{})
		if !ok {
			continue
		}

		field, _ := cond["field"].(string)
		operator, _ := cond["operator"].(string)
		value := cond["value"]

		// Get actual value from variables
		actualValue := e.getNestedValue(execCtx.Variables, field)

		// Evaluate condition
		result := e.evaluateCondition(actualValue, operator, value)
		results = append(results, result)
	}

	// Combine results
	finalResult := true
	if combineWith == "or" {
		finalResult = false
		for _, r := range results {
			if r {
				finalResult = true
				break
			}
		}
	} else {
		for _, r := range results {
			if !r {
				finalResult = false
				break
			}
		}
	}

	handle := "false"
	if finalResult {
		handle = "true"
	}

	return map[string]interface{}{
		"result":  finalResult,
		"_handle": handle,
	}, nil
}

func (e *FlowExecutor) evaluateCondition(actual interface{}, operator string, expected interface{}) bool {
	switch operator {
	case "eq", "equals", "==":
		return fmt.Sprintf("%v", actual) == fmt.Sprintf("%v", expected)
	case "ne", "not_equals", "!=":
		return fmt.Sprintf("%v", actual) != fmt.Sprintf("%v", expected)
	case "contains":
		return strings.Contains(fmt.Sprintf("%v", actual), fmt.Sprintf("%v", expected))
	case "starts_with":
		return strings.HasPrefix(fmt.Sprintf("%v", actual), fmt.Sprintf("%v", expected))
	case "ends_with":
		return strings.HasSuffix(fmt.Sprintf("%v", actual), fmt.Sprintf("%v", expected))
	case "gt", ">":
		return toFloat(actual) > toFloat(expected)
	case "lt", "<":
		return toFloat(actual) < toFloat(expected)
	case "gte", ">=":
		return toFloat(actual) >= toFloat(expected)
	case "lte", "<=":
		return toFloat(actual) <= toFloat(expected)
	case "empty":
		return fmt.Sprintf("%v", actual) == ""
	case "not_empty":
		return fmt.Sprintf("%v", actual) != ""
	default:
		return false
	}
}

func (e *FlowExecutor) executeDelay(ctx context.Context, execCtx *ExecutionContext, node *flow.Node) (map[string]interface{}, error) {
	data := node.Data

	duration, _ := data["duration"].(float64)
	unit, _ := data["unit"].(string)

	if duration <= 0 {
		return execCtx.Variables, nil
	}

	var sleepDuration time.Duration
	switch unit {
	case "minutes":
		sleepDuration = time.Duration(duration) * time.Minute
	case "hours":
		sleepDuration = time.Duration(duration) * time.Hour
	default:
		sleepDuration = time.Duration(duration) * time.Second
	}

	// Limit max delay to 5 minutes for safety
	if sleepDuration > 5*time.Minute {
		sleepDuration = 5 * time.Minute
	}

	select {
	case <-time.After(sleepDuration):
		return execCtx.Variables, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (e *FlowExecutor) executeSendMessage(ctx context.Context, execCtx *ExecutionContext, node *flow.Node) (map[string]interface{}, error) {
	data := node.Data

	message, _ := data["message"].(string)
	message = e.interpolateVariables(message, execCtx.Variables)

	// Store the message to be sent
	return map[string]interface{}{
		"message":         message,
		"response":        message,
		"reply_to_trigger": data["reply_to_trigger"],
	}, nil
}

func (e *FlowExecutor) executeSetVariable(ctx context.Context, execCtx *ExecutionContext, node *flow.Node) (map[string]interface{}, error) {
	data := node.Data

	name, _ := data["name"].(string)
	value := data["value"]

	if name == "" {
		return execCtx.Variables, nil
	}

	// Interpolate if string
	if str, ok := value.(string); ok {
		value = e.interpolateVariables(str, execCtx.Variables)
	}

	execCtx.Variables[name] = value

	return map[string]interface{}{
		name: value,
	}, nil
}

// === Helpers ===

func (e *FlowExecutor) interpolateVariables(template string, vars map[string]interface{}) string {
	re := regexp.MustCompile(`\{\{([^}]+)\}\}`)
	return re.ReplaceAllStringFunc(template, func(match string) string {
		varName := strings.TrimSpace(match[2 : len(match)-2])
		if val := e.getNestedValue(vars, varName); val != nil {
			return fmt.Sprintf("%v", val)
		}
		return match
	})
}

func (e *FlowExecutor) getNestedValue(data map[string]interface{}, path string) interface{} {
	parts := strings.Split(path, ".")
	current := interface{}(data)

	for _, part := range parts {
		switch v := current.(type) {
		case map[string]interface{}:
			current = v[part]
		default:
			return nil
		}
	}

	return current
}

func toFloat(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case float32:
		return float64(val)
	case int:
		return float64(val)
	case int64:
		return float64(val)
	case string:
		var f float64
		fmt.Sscanf(val, "%f", &f)
		return f
	default:
		return 0
	}
}

