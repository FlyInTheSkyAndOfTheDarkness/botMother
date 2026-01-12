package serp

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// SerpService provides internet search capabilities via SerpAPI
type SerpService struct {
	apiKey string
	client *http.Client
}

// NewService creates a new SerpAPI service
func NewService(apiKey string) *SerpService {
	if apiKey == "" {
		return nil
	}
	return &SerpService{
		apiKey: apiKey,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

// SearchResult represents a search result from SerpAPI
type SearchResult struct {
	Title       string `json:"title"`
	Link        string `json:"link"`
	Snippet     string `json:"snippet"`
	Source      string `json:"source,omitempty"`
	Date        string `json:"date,omitempty"`
}

// SearchResponse represents the full response from SerpAPI
type SearchResponse struct {
	OrganicResults []SearchResult `json:"organic_results"`
	AnswerBox      struct {
		Answer string `json:"answer"`
	} `json:"answer_box,omitempty"`
	KnowledgeGraph struct {
		Description string `json:"description"`
	} `json:"knowledge_graph,omitempty"`
}

// Search performs a Google search via SerpAPI
func (s *SerpService) Search(query string) (string, error) {
	if s == nil || s.apiKey == "" {
		return "", fmt.Errorf("SerpAPI service not configured")
	}

	apiURL := fmt.Sprintf("https://serpapi.com/search.json?engine=google&q=%s&api_key=%s", 
		url.QueryEscape(query), s.apiKey)

	resp, err := s.client.Get(apiURL)
	if err != nil {
		return "", fmt.Errorf("failed to fetch search results: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("SerpAPI error (status %d): %s", resp.StatusCode, string(body))
	}

	var searchResp SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return "", fmt.Errorf("failed to parse search response: %w", err)
	}

	// Build formatted result
	var result strings.Builder
	
	// Check for direct answer
	if searchResp.AnswerBox.Answer != "" {
		result.WriteString(fmt.Sprintf("Direct Answer: %s\n\n", searchResp.AnswerBox.Answer))
	}
	
	// Check for knowledge graph
	if searchResp.KnowledgeGraph.Description != "" {
		result.WriteString(fmt.Sprintf("Knowledge Graph: %s\n\n", searchResp.KnowledgeGraph.Description))
	}
	
	// Add top organic results
	if len(searchResp.OrganicResults) > 0 {
		result.WriteString("Search Results:\n")
		for i, res := range searchResp.OrganicResults {
			if i >= 5 { // Limit to top 5 results
				break
			}
			result.WriteString(fmt.Sprintf("%d. %s\n   %s\n   Source: %s\n\n", 
				i+1, res.Title, res.Snippet, res.Link))
		}
	} else {
		result.WriteString("No search results found.")
	}

	return result.String(), nil
}

