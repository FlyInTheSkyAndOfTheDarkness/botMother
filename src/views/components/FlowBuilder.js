// Flow Builder Component
const FlowBuilder = {
    props: ['agentId', 'flowId'],
    emits: ['close', 'save'],
    
    template: `
    <div class="fixed inset-0 z-50 bg-dark-bg">
        <!-- Header -->
        <div class="h-14 bg-dark-card border-b border-dark-border flex items-center justify-between px-4">
            <div class="flex items-center gap-4">
                <button @click="$emit('close')" class="p-2 hover:bg-dark-border rounded-lg transition-colors">
                    <svg class="w-5 h-5 text-dark-muted" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                    </svg>
                </button>
                <input v-model="flowName" type="text" 
                       class="bg-transparent text-white text-lg font-semibold focus:outline-none border-b border-transparent hover:border-dark-border focus:border-primary-500"
                       placeholder="Flow name">
            </div>
            <div class="flex items-center gap-3">
                <!-- Zoom Controls -->
                <div class="flex items-center gap-1 bg-dark-bg rounded-lg p-1">
                    <button @click="zoomOut" class="p-1.5 hover:bg-dark-border rounded text-dark-muted hover:text-white">
                        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20 12H4" />
                        </svg>
                    </button>
                    <span class="text-xs text-dark-muted px-2">{{ Math.round(zoom * 100) }}%</span>
                    <button @click="zoomIn" class="p-1.5 hover:bg-dark-border rounded text-dark-muted hover:text-white">
                        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
                        </svg>
                    </button>
                    <button @click="resetZoom" class="p-1.5 hover:bg-dark-border rounded text-dark-muted hover:text-white text-xs">
                        Reset
                    </button>
                </div>
                <label class="flex items-center gap-2 text-sm text-dark-muted">
                    <input type="checkbox" v-model="isActive" class="rounded">
                    Active
                </label>
                <button @click="saveFlow" :disabled="saving" 
                        class="px-4 py-2 bg-primary-600 hover:bg-primary-500 text-white rounded-lg transition-colors disabled:opacity-50">
                    {{ saving ? 'Saving...' : 'Save' }}
                </button>
            </div>
        </div>

        <div class="flex h-[calc(100vh-3.5rem)]">
            <!-- Sidebar - Node Palette -->
            <div class="w-56 bg-dark-card border-r border-dark-border overflow-y-auto">
                <div class="p-3">
                    <h3 class="text-xs font-semibold text-dark-muted uppercase tracking-wider mb-2">Triggers</h3>
                    <div class="space-y-1.5 mb-4">
                        <div v-for="node in triggerNodes" :key="node.type"
                             draggable="true" @dragstart="onDragStart($event, node)"
                             class="flex items-center gap-2 p-2 bg-dark-bg rounded-lg cursor-grab hover:bg-dark-border transition-colors text-sm">
                            <span>{{ node.icon }}</span>
                            <span class="text-white">{{ node.label }}</span>
                        </div>
                    </div>

                    <h3 class="text-xs font-semibold text-dark-muted uppercase tracking-wider mb-2">AI & Logic</h3>
                    <div class="space-y-1.5 mb-4">
                        <div v-for="node in aiNodes" :key="node.type"
                             draggable="true" @dragstart="onDragStart($event, node)"
                             class="flex items-center gap-2 p-2 bg-dark-bg rounded-lg cursor-grab hover:bg-dark-border transition-colors text-sm">
                            <span>{{ node.icon }}</span>
                            <span class="text-white">{{ node.label }}</span>
                        </div>
                    </div>

                    <h3 class="text-xs font-semibold text-dark-muted uppercase tracking-wider mb-2">Integrations</h3>
                    <div class="space-y-1.5 mb-4">
                        <div v-for="node in integrationNodes" :key="node.type"
                             draggable="true" @dragstart="onDragStart($event, node)"
                             class="flex items-center gap-2 p-2 bg-dark-bg rounded-lg cursor-grab hover:bg-dark-border transition-colors text-sm">
                            <span>{{ node.icon }}</span>
                            <span class="text-white">{{ node.label }}</span>
                        </div>
                    </div>

                    <h3 class="text-xs font-semibold text-dark-muted uppercase tracking-wider mb-2">Actions</h3>
                    <div class="space-y-1.5">
                        <div v-for="node in actionNodes" :key="node.type"
                             draggable="true" @dragstart="onDragStart($event, node)"
                             class="flex items-center gap-2 p-2 bg-dark-bg rounded-lg cursor-grab hover:bg-dark-border transition-colors text-sm">
                            <span>{{ node.icon }}</span>
                            <span class="text-white">{{ node.label }}</span>
                        </div>
                    </div>
                </div>
            </div>

            <!-- Canvas -->
            <div class="flex-1 relative overflow-hidden bg-dark-bg" 
                 ref="canvas"
                 @drop="onDrop" 
                 @dragover.prevent
                 @mousedown="onCanvasMouseDown"
                 @wheel="onWheel">
                
                <!-- Grid Background -->
                <div class="absolute inset-0 pointer-events-none" 
                     :style="{ 
                         backgroundImage: 'radial-gradient(circle, #334155 1px, transparent 1px)', 
                         backgroundSize: (20 * zoom) + 'px ' + (20 * zoom) + 'px',
                         transform: 'translate(' + pan.x + 'px, ' + pan.y + 'px)'
                     }"></div>

                <!-- Zoomable/Pannable container -->
                <div class="absolute inset-0" 
                     :style="{ 
                         transform: 'translate(' + pan.x + 'px, ' + pan.y + 'px) scale(' + zoom + ')',
                         transformOrigin: '0 0'
                     }">
                    
                    <!-- SVG for connections -->
                    <svg class="absolute inset-0 w-full h-full overflow-visible" style="z-index: 1; pointer-events: none;">
                        <defs>
                            <marker id="arrowhead" markerWidth="10" markerHeight="7" refX="9" refY="3.5" orient="auto">
                                <polygon points="0 0, 10 3.5, 0 7" fill="#64748b" />
                            </marker>
                            <marker id="arrowhead-active" markerWidth="10" markerHeight="7" refX="9" refY="3.5" orient="auto">
                                <polygon points="0 0, 10 3.5, 0 7" fill="#0ea5e9" />
                            </marker>
                        </defs>
                        <!-- Existing edges -->
                        <g v-for="edge in edges" :key="edge.id">
                            <path :d="getEdgePath(edge)"
                                  fill="none" stroke="#64748b" stroke-width="2" marker-end="url(#arrowhead)" />
                        </g>
                        <!-- Drawing edge -->
                        <path v-if="drawingEdge"
                              :d="getDrawingEdgePath()"
                              fill="none" stroke="#0ea5e9" stroke-width="2" stroke-dasharray="5,5" marker-end="url(#arrowhead-active)" />
                    </svg>

                    <!-- Nodes -->
                    <div v-for="node in nodes" :key="node.id"
                         :style="{ left: node.position.x + 'px', top: node.position.y + 'px' }"
                         class="absolute"
                         style="z-index: 10;"
                         @mousedown.stop="onNodeMouseDown($event, node)">
                        
                        <div :class="['w-56 rounded-xl border-2 transition-all cursor-move select-none', 
                                      selectedNode?.id === node.id ? 'border-primary-500 shadow-lg shadow-primary-500/30' : 'border-dark-border',
                                      getNodeBgClass(node.type)]">
                            <!-- Node Header -->
                            <div class="flex items-center gap-2 p-2.5 border-b border-dark-border/50">
                                <span class="text-base">{{ getNodeIcon(node.type) }}</span>
                                <span class="text-sm font-medium text-white flex-1 truncate">{{ node.label }}</span>
                                <button @click.stop="deleteNode(node.id)" class="p-1 hover:bg-red-500/20 rounded text-dark-muted hover:text-red-400">
                                    <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                                    </svg>
                                </button>
                            </div>
                            
                            <!-- Node Content Preview -->
                            <div class="p-2.5 text-xs text-dark-muted truncate">
                                {{ getNodePreview(node) }}
                            </div>

                            <!-- Input Connection Point (left) -->
                            <div v-if="!node.type.startsWith('trigger_')" 
                                 class="absolute -left-3 top-1/2 -translate-y-1/2 w-5 h-5 bg-dark-border rounded-full border-2 border-dark-bg cursor-pointer hover:bg-primary-500 hover:border-primary-400 transition-colors flex items-center justify-center"
                                 @mouseup.stop="onConnectEnd(node)">
                                <div class="w-2 h-2 bg-dark-bg rounded-full"></div>
                            </div>
                            
                            <!-- Output Connection Point (right) -->
                            <div v-if="node.type !== 'condition'"
                                 class="absolute -right-3 top-1/2 -translate-y-1/2 w-5 h-5 bg-dark-border rounded-full border-2 border-dark-bg cursor-pointer hover:bg-primary-500 hover:border-primary-400 transition-colors flex items-center justify-center"
                                 @mousedown.stop="onConnectStart($event, node)">
                                <div class="w-2 h-2 bg-dark-bg rounded-full"></div>
                            </div>
                            
                            <!-- Condition outputs (true/false) -->
                            <template v-if="node.type === 'condition'">
                                <div class="absolute -right-3 top-1/3 -translate-y-1/2 w-5 h-5 bg-green-500 rounded-full border-2 border-dark-bg cursor-pointer hover:bg-green-400 transition-colors flex items-center justify-center text-[8px] text-white font-bold"
                                     @mousedown.stop="onConnectStart($event, node, 'true')">✓</div>
                                <div class="absolute -right-3 top-2/3 -translate-y-1/2 w-5 h-5 bg-red-500 rounded-full border-2 border-dark-bg cursor-pointer hover:bg-red-400 transition-colors flex items-center justify-center text-[8px] text-white font-bold"
                                     @mousedown.stop="onConnectStart($event, node, 'false')">✗</div>
                            </template>
                        </div>
                    </div>
                </div>
            </div>

            <!-- Properties Panel -->
            <div v-if="selectedNode" class="w-72 bg-dark-card border-l border-dark-border overflow-y-auto">
                <div class="p-4">
                    <h3 class="text-lg font-semibold text-white mb-4">{{ selectedNode.label }}</h3>
                    
                    <div class="space-y-4">
                        <!-- Common: Label -->
                        <div>
                            <label class="block text-sm text-dark-muted mb-1">Label</label>
                            <input v-model="selectedNode.label" type="text"
                                   class="w-full px-3 py-2 bg-dark-bg border border-dark-border rounded-lg text-white text-sm focus:border-primary-500 focus:outline-none">
                        </div>

                        <!-- AI Agent Properties -->
                        <template v-if="selectedNode.type === 'ai_agent'">
                            <div>
                                <label class="block text-sm text-dark-muted mb-1">Model</label>
                                <select v-model="selectedNode.data.model"
                                        class="w-full px-3 py-2 bg-dark-bg border border-dark-border rounded-lg text-white text-sm focus:border-primary-500 focus:outline-none">
                                    <option value="gpt-4o-mini">GPT-4o Mini</option>
                                    <option value="gpt-4o">GPT-4o</option>
                                    <option value="gpt-4-turbo">GPT-4 Turbo</option>
                                </select>
                            </div>
                            <div>
                                <label class="block text-sm text-dark-muted mb-1">System Prompt</label>
                                <textarea v-model="selectedNode.data.system_prompt" rows="4"
                                          class="w-full px-3 py-2 bg-dark-bg border border-dark-border rounded-lg text-white text-sm focus:border-primary-500 focus:outline-none resize-none"
                                          placeholder="You are a helpful assistant..."></textarea>
                            </div>
                            <div>
                                <label class="block text-sm text-dark-muted mb-1">OpenAI API Key</label>
                                <input v-model="selectedNode.data.api_key" type="password"
                                       class="w-full px-3 py-2 bg-dark-bg border border-dark-border rounded-lg text-white text-sm focus:border-primary-500 focus:outline-none"
                                       placeholder="sk-...">
                                <p class="mt-1 text-xs text-dark-muted">Or select credential below</p>
                            </div>
                            <div>
                                <label class="block text-sm text-dark-muted mb-1">Credential</label>
                                <select v-model="selectedNode.data.credential_id"
                                        class="w-full px-3 py-2 bg-dark-bg border border-dark-border rounded-lg text-white text-sm focus:border-primary-500 focus:outline-none">
                                    <option value="">Select credential...</option>
                                    <option v-for="cred in credentials.filter(c => c.type === 'openai')" :key="cred.id" :value="cred.id">
                                        {{ cred.name }}
                                    </option>
                                </select>
                                <button @click="showOpenAICredentialModal = true" class="mt-2 text-xs text-primary-400 hover:text-primary-300">
                                    + Add OpenAI credential
                                </button>
                            </div>
                        </template>

                        <!-- HTTP Request Properties -->
                        <template v-if="selectedNode.type === 'http_request'">
                            <div>
                                <label class="block text-sm text-dark-muted mb-1">Method</label>
                                <select v-model="selectedNode.data.method"
                                        class="w-full px-3 py-2 bg-dark-bg border border-dark-border rounded-lg text-white text-sm focus:border-primary-500 focus:outline-none">
                                    <option value="GET">GET</option>
                                    <option value="POST">POST</option>
                                    <option value="PUT">PUT</option>
                                    <option value="DELETE">DELETE</option>
                                    <option value="PATCH">PATCH</option>
                                </select>
                            </div>
                            <div>
                                <label class="block text-sm text-dark-muted mb-1">URL</label>
                                <input v-model="selectedNode.data.url" type="text"
                                       class="w-full px-3 py-2 bg-dark-bg border border-dark-border rounded-lg text-white text-sm focus:border-primary-500 focus:outline-none"
                                       placeholder="https://api.example.com/...">
                            </div>
                            <div>
                                <label class="block text-sm text-dark-muted mb-1">Headers (JSON)</label>
                                <textarea v-model="selectedNode.data.headers_json" rows="3"
                                          class="w-full px-3 py-2 bg-dark-bg border border-dark-border rounded-lg text-white text-sm font-mono focus:border-primary-500 focus:outline-none resize-none"
                                          placeholder='{"Authorization": "Bearer ..."}'></textarea>
                            </div>
                            <div>
                                <label class="block text-sm text-dark-muted mb-1">Body</label>
                                <textarea v-model="selectedNode.data.body" rows="4"
                                          class="w-full px-3 py-2 bg-dark-bg border border-dark-border rounded-lg text-white text-sm font-mono focus:border-primary-500 focus:outline-none resize-none"
                                          placeholder='{"key": "value"}'></textarea>
                            </div>
                        </template>

                        <!-- Database Properties -->
                        <template v-if="selectedNode.type === 'database'">
                            <div>
                                <label class="block text-sm text-dark-muted mb-1">Credential</label>
                                <select v-model="selectedNode.data.credential_id"
                                        class="w-full px-3 py-2 bg-dark-bg border border-dark-border rounded-lg text-white text-sm focus:border-primary-500 focus:outline-none">
                                    <option value="">Select database...</option>
                                    <option v-for="cred in credentials.filter(c => c.type === 'database')" :key="cred.id" :value="cred.id">
                                        {{ cred.name }}
                                    </option>
                                </select>
                                <button @click="showCredentialModal = true" class="mt-2 text-xs text-primary-400 hover:text-primary-300">
                                    + Add new database
                                </button>
                            </div>
                            <div>
                                <label class="block text-sm text-dark-muted mb-1">Operation</label>
                                <select v-model="selectedNode.data.operation"
                                        class="w-full px-3 py-2 bg-dark-bg border border-dark-border rounded-lg text-white text-sm focus:border-primary-500 focus:outline-none">
                                    <option value="select">SELECT</option>
                                    <option value="insert">INSERT</option>
                                    <option value="update">UPDATE</option>
                                    <option value="delete">DELETE</option>
                                    <option value="raw">Raw SQL</option>
                                </select>
                            </div>
                            <div v-if="selectedNode.data.operation !== 'raw'">
                                <label class="block text-sm text-dark-muted mb-1">Table</label>
                                <input v-model="selectedNode.data.table" type="text"
                                       class="w-full px-3 py-2 bg-dark-bg border border-dark-border rounded-lg text-white text-sm focus:border-primary-500 focus:outline-none"
                                       placeholder="users">
                            </div>
                            <div v-if="selectedNode.data.operation === 'raw'">
                                <label class="block text-sm text-dark-muted mb-1">SQL Query</label>
                                <textarea v-model="selectedNode.data.query" rows="4"
                                          class="w-full px-3 py-2 bg-dark-bg border border-dark-border rounded-lg text-white text-sm font-mono focus:border-primary-500 focus:outline-none resize-none"
                                          placeholder="SELECT * FROM users WHERE id = $1"></textarea>
                            </div>
                            <div v-if="selectedNode.data.operation === 'select' || selectedNode.data.operation === 'update' || selectedNode.data.operation === 'delete'">
                                <label class="block text-sm text-dark-muted mb-1">WHERE</label>
                                <input v-model="selectedNode.data.where" type="text"
                                       class="w-full px-3 py-2 bg-dark-bg border border-dark-border rounded-lg text-white text-sm font-mono focus:border-primary-500 focus:outline-none"
                                       placeholder="id = 1">
                            </div>
                        </template>

                        <!-- Condition Properties -->
                        <template v-if="selectedNode.type === 'condition'">
                            <div>
                                <label class="block text-sm text-dark-muted mb-1">Field</label>
                                <input v-model="selectedNode.data.field" type="text"
                                       class="w-full px-3 py-2 bg-dark-bg border border-dark-border rounded-lg text-white text-sm focus:border-primary-500 focus:outline-none"
                                       placeholder="message">
                            </div>
                            <div>
                                <label class="block text-sm text-dark-muted mb-1">Operator</label>
                                <select v-model="selectedNode.data.operator"
                                        class="w-full px-3 py-2 bg-dark-bg border border-dark-border rounded-lg text-white text-sm focus:border-primary-500 focus:outline-none">
                                    <option value="eq">Equals (==)</option>
                                    <option value="ne">Not Equals (!=)</option>
                                    <option value="contains">Contains</option>
                                    <option value="starts_with">Starts With</option>
                                    <option value="ends_with">Ends With</option>
                                    <option value="gt">Greater Than (>)</option>
                                    <option value="lt">Less Than (<)</option>
                                    <option value="empty">Is Empty</option>
                                    <option value="not_empty">Is Not Empty</option>
                                </select>
                            </div>
                            <div>
                                <label class="block text-sm text-dark-muted mb-1">Value</label>
                                <input v-model="selectedNode.data.value" type="text"
                                       class="w-full px-3 py-2 bg-dark-bg border border-dark-border rounded-lg text-white text-sm focus:border-primary-500 focus:outline-none"
                                       placeholder="hello">
                            </div>
                        </template>

                        <!-- Send Message Properties -->
                        <template v-if="selectedNode.type === 'send_message'">
                            <div>
                                <label class="block text-sm text-dark-muted mb-1">Message</label>
                                <textarea v-model="selectedNode.data.message" rows="4"
                                          class="w-full px-3 py-2 bg-dark-bg border border-dark-border rounded-lg text-white text-sm focus:border-primary-500 focus:outline-none resize-none"
                                          placeholder="Hello! {{ai_response}}"></textarea>
                                <p class="mt-1 text-xs text-dark-muted">Use {{variable}} for dynamic content</p>
                            </div>
                            <div class="flex items-center gap-2">
                                <input type="checkbox" v-model="selectedNode.data.reply_to_trigger" id="replyToTrigger" class="rounded">
                                <label for="replyToTrigger" class="text-sm text-dark-muted">Reply to trigger message</label>
                            </div>
                        </template>

                        <!-- Delay Properties -->
                        <template v-if="selectedNode.type === 'delay'">
                            <div class="flex gap-2">
                                <div class="flex-1">
                                    <label class="block text-sm text-dark-muted mb-1">Duration</label>
                                    <input v-model.number="selectedNode.data.duration" type="number" min="1"
                                           class="w-full px-3 py-2 bg-dark-bg border border-dark-border rounded-lg text-white text-sm focus:border-primary-500 focus:outline-none">
                                </div>
                                <div class="flex-1">
                                    <label class="block text-sm text-dark-muted mb-1">Unit</label>
                                    <select v-model="selectedNode.data.unit"
                                            class="w-full px-3 py-2 bg-dark-bg border border-dark-border rounded-lg text-white text-sm focus:border-primary-500 focus:outline-none">
                                        <option value="seconds">Seconds</option>
                                        <option value="minutes">Minutes</option>
                                    </select>
                                </div>
                            </div>
                        </template>
                    </div>
                </div>
            </div>
        </div>

        <!-- Database Credential Modal -->
        <div v-if="showCredentialModal" class="fixed inset-0 z-[60] flex items-center justify-center bg-black/70 backdrop-blur-sm">
            <div class="bg-dark-card rounded-2xl p-6 max-w-md w-full mx-4 border border-dark-border">
                <h3 class="text-xl font-bold text-white mb-4">Add Database Connection</h3>
                
                <div class="space-y-4">
                    <div>
                        <label class="block text-sm text-dark-muted mb-1">Name</label>
                        <input v-model="newCredential.name" type="text"
                               class="w-full px-3 py-2 bg-dark-bg border border-dark-border rounded-lg text-white text-sm"
                               placeholder="My Supabase DB">
                    </div>
                    <div>
                        <label class="block text-sm text-dark-muted mb-1">Host</label>
                        <input v-model="newCredential.host" type="text"
                               class="w-full px-3 py-2 bg-dark-bg border border-dark-border rounded-lg text-white text-sm"
                               placeholder="db.xxxxx.supabase.co">
                    </div>
                    <div class="flex gap-2">
                        <div class="flex-1">
                            <label class="block text-sm text-dark-muted mb-1">Port</label>
                            <input v-model.number="newCredential.port" type="number"
                                   class="w-full px-3 py-2 bg-dark-bg border border-dark-border rounded-lg text-white text-sm"
                                   placeholder="5432">
                        </div>
                        <div class="flex-1">
                            <label class="block text-sm text-dark-muted mb-1">Database</label>
                            <input v-model="newCredential.database" type="text"
                                   class="w-full px-3 py-2 bg-dark-bg border border-dark-border rounded-lg text-white text-sm"
                                   placeholder="postgres">
                        </div>
                    </div>
                    <div>
                        <label class="block text-sm text-dark-muted mb-1">User</label>
                        <input v-model="newCredential.user" type="text"
                               class="w-full px-3 py-2 bg-dark-bg border border-dark-border rounded-lg text-white text-sm"
                               placeholder="postgres">
                    </div>
                    <div>
                        <label class="block text-sm text-dark-muted mb-1">Password</label>
                        <input v-model="newCredential.password" type="password"
                               class="w-full px-3 py-2 bg-dark-bg border border-dark-border rounded-lg text-white text-sm">
                    </div>
                    <div>
                        <label class="block text-sm text-dark-muted mb-1">SSL Mode</label>
                        <select v-model="newCredential.ssl_mode"
                                class="w-full px-3 py-2 bg-dark-bg border border-dark-border rounded-lg text-white text-sm">
                            <option value="require">Require (Supabase)</option>
                            <option value="disable">Disable</option>
                            <option value="verify-ca">Verify CA</option>
                            <option value="verify-full">Verify Full</option>
                        </select>
                    </div>
                </div>

                <div class="flex gap-3 mt-6">
                    <button @click="showCredentialModal = false" 
                            class="flex-1 py-2 bg-dark-border hover:bg-dark-muted/20 text-white rounded-lg">
                        Cancel
                    </button>
                    <button @click="saveCredential('database')" :disabled="savingCredential"
                            class="flex-1 py-2 bg-primary-600 hover:bg-primary-500 text-white rounded-lg">
                        {{ savingCredential ? 'Saving...' : 'Save' }}
                    </button>
                </div>
            </div>
        </div>

        <!-- OpenAI Credential Modal -->
        <div v-if="showOpenAICredentialModal" class="fixed inset-0 z-[60] flex items-center justify-center bg-black/70 backdrop-blur-sm">
            <div class="bg-dark-card rounded-2xl p-6 max-w-md w-full mx-4 border border-dark-border">
                <h3 class="text-xl font-bold text-white mb-4">Add OpenAI Credential</h3>
                
                <div class="space-y-4">
                    <div>
                        <label class="block text-sm text-dark-muted mb-1">Name</label>
                        <input v-model="newOpenAICredential.name" type="text"
                               class="w-full px-3 py-2 bg-dark-bg border border-dark-border rounded-lg text-white text-sm"
                               placeholder="My OpenAI Key">
                    </div>
                    <div>
                        <label class="block text-sm text-dark-muted mb-1">API Key</label>
                        <input v-model="newOpenAICredential.api_key" type="password"
                               class="w-full px-3 py-2 bg-dark-bg border border-dark-border rounded-lg text-white text-sm"
                               placeholder="sk-...">
                    </div>
                </div>

                <div class="flex gap-3 mt-6">
                    <button @click="showOpenAICredentialModal = false" 
                            class="flex-1 py-2 bg-dark-border hover:bg-dark-muted/20 text-white rounded-lg">
                        Cancel
                    </button>
                    <button @click="saveOpenAICredential" :disabled="savingCredential"
                            class="flex-1 py-2 bg-primary-600 hover:bg-primary-500 text-white rounded-lg">
                        {{ savingCredential ? 'Saving...' : 'Save' }}
                    </button>
                </div>
            </div>
        </div>
    </div>
    `,

    setup(props, { emit }) {
        const { ref, reactive, onMounted, onUnmounted } = Vue;

        // State
        const flowName = ref('New Flow');
        const isActive = ref(true);
        const saving = ref(false);
        const nodes = ref([]);
        const edges = ref([]);
        const selectedNode = ref(null);
        const credentials = ref([]);
        const canvas = ref(null);

        // Zoom and pan
        const zoom = ref(1);
        const pan = reactive({ x: 0, y: 0 });
        const isPanning = ref(false);
        const panStart = reactive({ x: 0, y: 0 });

        // Dragging state
        const draggingNode = ref(null);
        const dragOffset = reactive({ x: 0, y: 0 });

        // Connection drawing state
        const drawingEdge = ref(false);
        const edgeStart = reactive({ node: null, handle: '', x: 0, y: 0 });
        const edgeEnd = reactive({ x: 0, y: 0 });

        // Credential modals
        const showCredentialModal = ref(false);
        const showOpenAICredentialModal = ref(false);
        const savingCredential = ref(false);
        const newCredential = reactive({
            name: '',
            host: '',
            port: 5432,
            database: 'postgres',
            user: 'postgres',
            password: '',
            ssl_mode: 'require'
        });
        const newOpenAICredential = reactive({
            name: '',
            api_key: ''
        });

        // Node definitions
        const triggerNodes = [
            { type: 'trigger_whatsapp', label: 'WhatsApp', icon: '📱' },
            { type: 'trigger_telegram', label: 'Telegram', icon: '✈️' },
            { type: 'trigger_instagram', label: 'Instagram', icon: '📷' },
            { type: 'trigger_webhook', label: 'Webhook', icon: '🪝' },
        ];

        const aiNodes = [
            { type: 'ai_agent', label: 'AI Agent', icon: '🤖' },
            { type: 'condition', label: 'Condition', icon: '⚡' },
            { type: 'delay', label: 'Delay', icon: '⏱️' },
        ];

        const integrationNodes = [
            { type: 'http_request', label: 'HTTP Request', icon: '🌐' },
            { type: 'database', label: 'Database', icon: '🗄️' },
        ];

        const actionNodes = [
            { type: 'send_message', label: 'Send Message', icon: '💬' },
        ];

        // Methods
        const generateId = () => 'node_' + Math.random().toString(36).substr(2, 9);

        const onDragStart = (event, nodeType) => {
            event.dataTransfer.setData('nodeType', JSON.stringify(nodeType));
        };

        const onDrop = (event) => {
            const data = event.dataTransfer.getData('nodeType');
            if (!data) return;

            const nodeType = JSON.parse(data);
            const rect = canvas.value.getBoundingClientRect();
            const x = (event.clientX - rect.left - pan.x) / zoom.value - 112;
            const y = (event.clientY - rect.top - pan.y) / zoom.value - 30;

            const newNode = {
                id: generateId(),
                type: nodeType.type,
                label: nodeType.label,
                position: { x, y },
                data: getDefaultNodeData(nodeType.type)
            };

            nodes.value.push(newNode);
            selectedNode.value = newNode;
        };

        const getDefaultNodeData = (type) => {
            switch (type) {
                case 'ai_agent':
                    return { model: 'gpt-4o-mini', system_prompt: '', api_key: '', credential_id: '' };
                case 'http_request':
                    return { method: 'GET', url: '', headers_json: '', body: '' };
                case 'database':
                    return { credential_id: '', operation: 'select', table: '', query: '', where: '' };
                case 'condition':
                    return { field: '', operator: 'eq', value: '' };
                case 'send_message':
                    return { message: '{{ai_response}}', reply_to_trigger: true };
                case 'delay':
                    return { duration: 1, unit: 'seconds' };
                default:
                    return {};
            }
        };

        const onNodeMouseDown = (event, node) => {
            if (event.target.closest('button')) return;
            
            draggingNode.value = node;
            const rect = canvas.value.getBoundingClientRect();
            dragOffset.x = (event.clientX - rect.left - pan.x) / zoom.value - node.position.x;
            dragOffset.y = (event.clientY - rect.top - pan.y) / zoom.value - node.position.y;
            selectedNode.value = node;
            event.preventDefault();
        };

        const onCanvasMouseDown = (event) => {
            if (event.target === canvas.value || event.target.classList.contains('pointer-events-none')) {
                selectedNode.value = null;
                // Start panning with middle mouse or shift+click
                if (event.button === 1 || (event.button === 0 && event.shiftKey)) {
                    isPanning.value = true;
                    panStart.x = event.clientX - pan.x;
                    panStart.y = event.clientY - pan.y;
                    event.preventDefault();
                }
            }
        };

        const onMouseMove = (event) => {
            if (draggingNode.value) {
                const rect = canvas.value.getBoundingClientRect();
                draggingNode.value.position.x = (event.clientX - rect.left - pan.x) / zoom.value - dragOffset.x;
                draggingNode.value.position.y = (event.clientY - rect.top - pan.y) / zoom.value - dragOffset.y;
            }
            
            if (isPanning.value) {
                pan.x = event.clientX - panStart.x;
                pan.y = event.clientY - panStart.y;
            }
            
            if (drawingEdge.value) {
                const rect = canvas.value.getBoundingClientRect();
                edgeEnd.x = (event.clientX - rect.left - pan.x) / zoom.value;
                edgeEnd.y = (event.clientY - rect.top - pan.y) / zoom.value;
            }
        };

        const onMouseUp = () => {
            draggingNode.value = null;
            isPanning.value = false;
            if (drawingEdge.value && !edgeStart.node) {
                drawingEdge.value = false;
            }
        };

        const onConnectStart = (event, node, handle = '') => {
            drawingEdge.value = true;
            edgeStart.node = node;
            edgeStart.handle = handle;
            edgeStart.x = node.position.x + 224;
            edgeStart.y = node.position.y + 40;
            if (handle === 'true') edgeStart.y = node.position.y + 27;
            if (handle === 'false') edgeStart.y = node.position.y + 53;
            
            const rect = canvas.value.getBoundingClientRect();
            edgeEnd.x = (event.clientX - rect.left - pan.x) / zoom.value;
            edgeEnd.y = (event.clientY - rect.top - pan.y) / zoom.value;
            event.preventDefault();
        };

        const onConnectEnd = (targetNode) => {
            if (drawingEdge.value && edgeStart.node && edgeStart.node.id !== targetNode.id) {
                const exists = edges.value.some(e => 
                    e.source === edgeStart.node.id && e.target === targetNode.id && e.source_handle === edgeStart.handle
                );
                
                if (!exists) {
                    edges.value.push({
                        id: 'edge_' + Math.random().toString(36).substr(2, 9),
                        source: edgeStart.node.id,
                        target: targetNode.id,
                        source_handle: edgeStart.handle
                    });
                }
            }
            drawingEdge.value = false;
            edgeStart.node = null;
        };

        const getEdgePath = (edge) => {
            const sourceNode = nodes.value.find(n => n.id === edge.source);
            const targetNode = nodes.value.find(n => n.id === edge.target);
            if (!sourceNode || !targetNode) return '';

            const sx = sourceNode.position.x + 224;
            let sy = sourceNode.position.y + 40;
            
            if (edge.source_handle === 'true') sy = sourceNode.position.y + 27;
            if (edge.source_handle === 'false') sy = sourceNode.position.y + 53;

            const tx = targetNode.position.x;
            const ty = targetNode.position.y + 40;

            const dx = tx - sx;
            const mx = sx + dx * 0.5;

            return `M ${sx} ${sy} C ${mx} ${sy}, ${mx} ${ty}, ${tx} ${ty}`;
        };

        const getDrawingEdgePath = () => {
            const sx = edgeStart.x;
            const sy = edgeStart.y;
            const tx = edgeEnd.x;
            const ty = edgeEnd.y;
            const dx = tx - sx;
            const mx = sx + dx * 0.5;
            return `M ${sx} ${sy} C ${mx} ${sy}, ${mx} ${ty}, ${tx} ${ty}`;
        };

        const deleteNode = (nodeId) => {
            nodes.value = nodes.value.filter(n => n.id !== nodeId);
            edges.value = edges.value.filter(e => e.source !== nodeId && e.target !== nodeId);
            if (selectedNode.value?.id === nodeId) {
                selectedNode.value = null;
            }
        };

        const getNodeIcon = (type) => {
            const allNodes = [...triggerNodes, ...aiNodes, ...integrationNodes, ...actionNodes];
            return allNodes.find(n => n.type === type)?.icon || '📦';
        };

        const getNodeBgClass = (type) => {
            if (type.startsWith('trigger_')) return 'bg-emerald-900/50';
            if (type === 'ai_agent') return 'bg-purple-900/50';
            if (type === 'condition') return 'bg-yellow-900/50';
            if (type === 'http_request' || type === 'database') return 'bg-blue-900/50';
            if (type === 'send_message') return 'bg-pink-900/50';
            return 'bg-dark-card';
        };

        const getNodePreview = (node) => {
            switch (node.type) {
                case 'ai_agent':
                    return node.data.model || 'Configure AI model';
                case 'http_request':
                    return `${node.data.method || 'GET'} ${node.data.url || 'Set URL'}`;
                case 'database':
                    return `${node.data.operation?.toUpperCase() || 'SELECT'} ${node.data.table || '...'}`;
                case 'condition':
                    return `${node.data.field || '...'} ${node.data.operator || '=='} ${node.data.value || '...'}`;
                case 'send_message':
                    return (node.data.message?.substring(0, 25) || 'Set message') + '...';
                case 'delay':
                    return `Wait ${node.data.duration || 0} ${node.data.unit || 'seconds'}`;
                default:
                    return 'Trigger';
            }
        };

        // Zoom functions
        const zoomIn = () => {
            zoom.value = Math.min(zoom.value * 1.2, 2);
        };

        const zoomOut = () => {
            zoom.value = Math.max(zoom.value / 1.2, 0.25);
        };

        const resetZoom = () => {
            zoom.value = 1;
            pan.x = 0;
            pan.y = 0;
        };

        const onWheel = (event) => {
            if (event.ctrlKey || event.metaKey) {
                event.preventDefault();
                const delta = event.deltaY > 0 ? 0.9 : 1.1;
                zoom.value = Math.min(Math.max(zoom.value * delta, 0.25), 2);
            }
        };

        const saveCredential = async (type) => {
            savingCredential.value = true;
            try {
                const config = JSON.stringify({
                    host: newCredential.host,
                    port: newCredential.port,
                    database: newCredential.database,
                    user: newCredential.user,
                    password: newCredential.password,
                    ssl_mode: newCredential.ssl_mode
                });

                const response = await axios.post('/api/credentials', {
                    agent_id: props.agentId,
                    name: newCredential.name,
                    type: type,
                    config: config
                });

                credentials.value.push(response.data.results);
                showCredentialModal.value = false;
                
                Object.assign(newCredential, {
                    name: '', host: '', port: 5432, database: 'postgres',
                    user: 'postgres', password: '', ssl_mode: 'require'
                });
            } catch (error) {
                console.error('Failed to save credential:', error);
            } finally {
                savingCredential.value = false;
            }
        };

        const saveOpenAICredential = async () => {
            savingCredential.value = true;
            try {
                const config = JSON.stringify({
                    api_key: newOpenAICredential.api_key
                });

                const response = await axios.post('/api/credentials', {
                    agent_id: props.agentId,
                    name: newOpenAICredential.name,
                    type: 'openai',
                    config: config
                });

                credentials.value.push(response.data.results);
                showOpenAICredentialModal.value = false;
                
                newOpenAICredential.name = '';
                newOpenAICredential.api_key = '';
            } catch (error) {
                console.error('Failed to save credential:', error);
            } finally {
                savingCredential.value = false;
            }
        };

        const loadFlow = async () => {
            if (!props.flowId) return;
            try {
                const response = await axios.get(`/api/flows/${props.flowId}`);
                const flow = response.data.results;
                flowName.value = flow.name;
                isActive.value = flow.is_active;
                nodes.value = flow.nodes || [];
                edges.value = flow.edges || [];
            } catch (error) {
                console.error('Failed to load flow:', error);
            }
        };

        const loadCredentials = async () => {
            try {
                const response = await axios.get(`/api/credentials?agent_id=${props.agentId}`);
                credentials.value = response.data.results || [];
            } catch (error) {
                console.error('Failed to load credentials:', error);
                credentials.value = [];
            }
        };

        const saveFlow = async () => {
            saving.value = true;
            try {
                const flowData = {
                    agent_id: props.agentId,
                    name: flowName.value,
                    is_active: isActive.value,
                    nodes: nodes.value,
                    edges: edges.value
                };

                if (props.flowId) {
                    await axios.put(`/api/flows/${props.flowId}`, flowData);
                } else {
                    await axios.post('/api/flows', flowData);
                }
                
                emit('save');
                emit('close');
            } catch (error) {
                console.error('Failed to save flow:', error);
            } finally {
                saving.value = false;
            }
        };

        onMounted(() => {
            loadFlow();
            loadCredentials();
            
            // Global mouse events for smooth dragging
            window.addEventListener('mousemove', onMouseMove);
            window.addEventListener('mouseup', onMouseUp);
        });

        onUnmounted(() => {
            window.removeEventListener('mousemove', onMouseMove);
            window.removeEventListener('mouseup', onMouseUp);
        });

        return {
            flowName, isActive, saving, nodes, edges, selectedNode, credentials, canvas,
            zoom, pan,
            draggingNode, drawingEdge,
            showCredentialModal, showOpenAICredentialModal, savingCredential, newCredential, newOpenAICredential,
            triggerNodes, aiNodes, integrationNodes, actionNodes,
            onDragStart, onDrop, onNodeMouseDown, onCanvasMouseDown, onConnectStart, onConnectEnd,
            getEdgePath, getDrawingEdgePath, deleteNode, getNodeIcon, getNodeBgClass, getNodePreview,
            zoomIn, zoomOut, resetZoom, onWheel,
            saveCredential, saveOpenAICredential, loadCredentials, saveFlow
        };
    }
};
