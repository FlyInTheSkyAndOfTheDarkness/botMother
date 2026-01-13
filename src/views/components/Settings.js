// Agent Settings Component
const AgentSettings = {
    props: ['agentId'],
    emits: ['close'],
    
    template: `
    <div class="space-y-6">
        <div class="flex items-center justify-between">
            <h3 class="text-xl font-bold text-white">Agent Settings</h3>
            <button @click="$emit('close')" class="text-dark-muted hover:text-white">
                <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                </svg>
            </button>
        </div>

        <!-- Tabs -->
        <div class="flex gap-1 p-1 bg-dark-bg rounded-xl">
            <button v-for="tab in tabs" :key="tab.id"
                    @click="activeTab = tab.id"
                    :class="['flex-1 px-4 py-2 rounded-lg text-sm font-medium transition-colors',
                             activeTab === tab.id ? 'bg-primary-500 text-white' : 'text-dark-muted hover:text-white']">
                {{ tab.icon }} {{ tab.label }}
            </button>
        </div>

        <!-- Working Hours -->
        <div v-if="activeTab === 'hours'" class="space-y-4">
            <div class="flex items-center justify-between p-4 bg-dark-bg rounded-xl">
                <div>
                    <p class="text-white font-medium">Enable Working Hours</p>
                    <p class="text-sm text-dark-muted">Only respond during specified hours</p>
                </div>
                <label class="relative inline-flex items-center cursor-pointer">
                    <input type="checkbox" v-model="settings.working_hours.enabled" class="sr-only peer">
                    <div class="w-11 h-6 bg-dark-border rounded-full peer peer-checked:bg-primary-500 
                                after:content-[''] after:absolute after:top-[2px] after:left-[2px] 
                                after:bg-white after:rounded-full after:h-5 after:w-5 after:transition-all 
                                peer-checked:after:translate-x-full"></div>
                </label>
            </div>

            <div v-if="settings.working_hours.enabled" class="space-y-4">
                <div>
                    <label class="block text-sm font-medium text-dark-text mb-2">Timezone</label>
                    <select v-model="settings.working_hours.timezone"
                            class="w-full px-4 py-2 bg-dark-bg border border-dark-border rounded-lg text-white focus:border-primary-500 focus:outline-none">
                        <option value="UTC">UTC</option>
                        <option value="Europe/Moscow">Moscow (GMT+3)</option>
                        <option value="Europe/London">London (GMT+0)</option>
                        <option value="America/New_York">New York (GMT-5)</option>
                        <option value="America/Los_Angeles">Los Angeles (GMT-8)</option>
                        <option value="Asia/Tokyo">Tokyo (GMT+9)</option>
                    </select>
                </div>

                <div>
                    <label class="block text-sm font-medium text-dark-text mb-2">Away Message</label>
                    <textarea v-model="settings.working_hours.away_message" rows="2"
                              class="w-full px-4 py-2 bg-dark-bg border border-dark-border rounded-lg text-white focus:border-primary-500 focus:outline-none resize-none"></textarea>
                </div>

                <div class="space-y-2">
                    <p class="text-sm font-medium text-dark-text">Schedule</p>
                    <div v-for="day in settings.working_hours.schedule" :key="day.day"
                         class="flex items-center gap-4 p-3 bg-dark-bg rounded-lg">
                        <label class="flex items-center gap-2 w-24">
                            <input type="checkbox" v-model="day.is_working" 
                                   class="rounded bg-dark-border border-dark-border text-primary-500 focus:ring-primary-500">
                            <span class="text-white text-sm">{{ getDayName(day.day) }}</span>
                        </label>
                        <template v-if="day.is_working">
                            <input type="time" v-model="day.start_time"
                                   class="px-3 py-1 bg-dark-card border border-dark-border rounded text-white text-sm focus:border-primary-500 focus:outline-none">
                            <span class="text-dark-muted">to</span>
                            <input type="time" v-model="day.end_time"
                                   class="px-3 py-1 bg-dark-card border border-dark-border rounded text-white text-sm focus:border-primary-500 focus:outline-none">
                        </template>
                        <span v-else class="text-dark-muted text-sm">Day off</span>
                    </div>
                </div>
            </div>
        </div>

        <!-- Translation -->
        <div v-if="activeTab === 'translation'" class="space-y-4">
            <div class="flex items-center justify-between p-4 bg-dark-bg rounded-xl">
                <div>
                    <p class="text-white font-medium">Enable Auto-Translation</p>
                    <p class="text-sm text-dark-muted">Automatically translate messages</p>
                </div>
                <label class="relative inline-flex items-center cursor-pointer">
                    <input type="checkbox" v-model="settings.translation.enabled" class="sr-only peer">
                    <div class="w-11 h-6 bg-dark-border rounded-full peer peer-checked:bg-primary-500 
                                after:content-[''] after:absolute after:top-[2px] after:left-[2px] 
                                after:bg-white after:rounded-full after:h-5 after:w-5 after:transition-all 
                                peer-checked:after:translate-x-full"></div>
                </label>
            </div>

            <div v-if="settings.translation.enabled" class="space-y-4">
                <div>
                    <label class="block text-sm font-medium text-dark-text mb-2">Agent Language</label>
                    <select v-model="settings.translation.source_language"
                            class="w-full px-4 py-2 bg-dark-bg border border-dark-border rounded-lg text-white focus:border-primary-500 focus:outline-none">
                        <option value="en">English</option>
                        <option value="ru">Russian</option>
                        <option value="es">Spanish</option>
                        <option value="fr">French</option>
                        <option value="de">German</option>
                        <option value="zh">Chinese</option>
                        <option value="ja">Japanese</option>
                        <option value="ar">Arabic</option>
                    </select>
                </div>

                <div class="flex items-center gap-4">
                    <label class="flex items-center gap-2">
                        <input type="checkbox" v-model="settings.translation.auto_detect"
                               class="rounded bg-dark-border text-primary-500 focus:ring-primary-500">
                        <span class="text-white text-sm">Auto-detect incoming language</span>
                    </label>
                </div>

                <div class="flex items-center gap-4">
                    <label class="flex items-center gap-2">
                        <input type="checkbox" v-model="settings.translation.translate_incoming"
                               class="rounded bg-dark-border text-primary-500 focus:ring-primary-500">
                        <span class="text-white text-sm">Translate incoming messages</span>
                    </label>
                </div>

                <div class="flex items-center gap-4">
                    <label class="flex items-center gap-2">
                        <input type="checkbox" v-model="settings.translation.translate_outgoing"
                               class="rounded bg-dark-border text-primary-500 focus:ring-primary-500">
                        <span class="text-white text-sm">Translate outgoing responses</span>
                    </label>
                </div>
            </div>
        </div>

        <!-- Follow-up -->
        <div v-if="activeTab === 'followup'" class="space-y-4">
            <div class="flex items-center justify-between p-4 bg-dark-bg rounded-xl">
                <div>
                    <p class="text-white font-medium">Enable Follow-up Messages</p>
                    <p class="text-sm text-dark-muted">Auto-send messages after inactivity</p>
                </div>
                <label class="relative inline-flex items-center cursor-pointer">
                    <input type="checkbox" v-model="settings.follow_up.enabled" class="sr-only peer">
                    <div class="w-11 h-6 bg-dark-border rounded-full peer peer-checked:bg-primary-500 
                                after:content-[''] after:absolute after:top-[2px] after:left-[2px] 
                                after:bg-white after:rounded-full after:h-5 after:w-5 after:transition-all 
                                peer-checked:after:translate-x-full"></div>
                </label>
            </div>

            <div v-if="settings.follow_up.enabled" class="space-y-4">
                <div class="grid grid-cols-2 gap-4">
                    <div>
                        <label class="block text-sm font-medium text-dark-text mb-2">Delay (minutes)</label>
                        <input type="number" v-model.number="settings.follow_up.delay_minutes" min="1"
                               class="w-full px-4 py-2 bg-dark-bg border border-dark-border rounded-lg text-white focus:border-primary-500 focus:outline-none">
                    </div>
                    <div>
                        <label class="block text-sm font-medium text-dark-text mb-2">Max Follow-ups</label>
                        <input type="number" v-model.number="settings.follow_up.max_follow_ups" min="1" max="5"
                               class="w-full px-4 py-2 bg-dark-bg border border-dark-border rounded-lg text-white focus:border-primary-500 focus:outline-none">
                    </div>
                </div>

                <div class="flex items-center gap-4">
                    <label class="flex items-center gap-2">
                        <input type="checkbox" v-model="settings.follow_up.only_if_no_reply"
                               class="rounded bg-dark-border text-primary-500 focus:ring-primary-500">
                        <span class="text-white text-sm">Only send if user hasn't replied</span>
                    </label>
                </div>

                <div>
                    <label class="block text-sm font-medium text-dark-text mb-2">Follow-up Messages</label>
                    <div class="space-y-2">
                        <div v-for="(msg, index) in settings.follow_up.messages" :key="index"
                             class="flex gap-2">
                            <input type="text" v-model="settings.follow_up.messages[index]"
                                   class="flex-1 px-4 py-2 bg-dark-bg border border-dark-border rounded-lg text-white focus:border-primary-500 focus:outline-none">
                            <button @click="removeFollowUpMessage(index)" class="px-3 text-red-400 hover:text-red-300">
                                ✕
                            </button>
                        </div>
                        <button @click="addFollowUpMessage" class="text-primary-400 hover:text-primary-300 text-sm">
                            + Add message
                        </button>
                    </div>
                </div>
            </div>
        </div>

        <!-- AI Settings -->
        <div v-if="activeTab === 'ai'" class="space-y-4">
            <div>
                <label class="block text-sm font-medium text-dark-text mb-2">
                    Temperature: {{ settings.temperature }}
                </label>
                <input type="range" v-model.number="settings.temperature" min="0" max="1" step="0.1"
                       class="w-full h-2 bg-dark-border rounded-lg appearance-none cursor-pointer">
                <div class="flex justify-between text-xs text-dark-muted mt-1">
                    <span>Precise</span>
                    <span>Creative</span>
                </div>
            </div>

            <div>
                <label class="block text-sm font-medium text-dark-text mb-2">Max Tokens per Message</label>
                <input type="number" v-model.number="settings.max_tokens_per_msg" min="100" max="2000"
                       class="w-full px-4 py-2 bg-dark-bg border border-dark-border rounded-lg text-white focus:border-primary-500 focus:outline-none">
                <p class="text-xs text-dark-muted mt-1">Maximum length of AI responses (~750 words = 1000 tokens)</p>
            </div>

            <div class="flex items-center justify-between p-4 bg-dark-bg rounded-xl">
                <div>
                    <p class="text-white font-medium">Sentiment Analysis</p>
                    <p class="text-sm text-dark-muted">Analyze user message sentiment</p>
                </div>
                <label class="relative inline-flex items-center cursor-pointer">
                    <input type="checkbox" v-model="settings.sentiment.enabled" class="sr-only peer">
                    <div class="w-11 h-6 bg-dark-border rounded-full peer peer-checked:bg-primary-500 
                                after:content-[''] after:absolute after:top-[2px] after:left-[2px] 
                                after:bg-white after:rounded-full after:h-5 after:w-5 after:transition-all 
                                peer-checked:after:translate-x-full"></div>
                </label>
            </div>

            <div v-if="settings.sentiment.enabled" class="space-y-4 pl-4">
                <label class="flex items-center gap-2">
                    <input type="checkbox" v-model="settings.sentiment.alert_on_negative"
                           class="rounded bg-dark-border text-primary-500 focus:ring-primary-500">
                    <span class="text-white text-sm">Alert on negative sentiment</span>
                </label>
                <label class="flex items-center gap-2">
                    <input type="checkbox" v-model="settings.sentiment.escalate_on_very_negative"
                           class="rounded bg-dark-border text-primary-500 focus:ring-primary-500">
                    <span class="text-white text-sm">Auto-escalate very negative messages to human</span>
                </label>
            </div>
        </div>

        <!-- Save Button -->
        <div class="flex justify-end pt-4 border-t border-dark-border">
            <button @click="saveSettings" :disabled="saving"
                    class="px-6 py-2 bg-primary-600 hover:bg-primary-500 text-white font-medium rounded-xl transition-colors disabled:opacity-50">
                {{ saving ? 'Saving...' : 'Save Settings' }}
            </button>
        </div>
    </div>
    `,

    setup(props, { emit }) {
        const { ref, reactive, onMounted } = Vue;

        const activeTab = ref('hours');
        const saving = ref(false);
        
        const tabs = [
            { id: 'hours', label: 'Working Hours', icon: '🕐' },
            { id: 'translation', label: 'Translation', icon: '🌍' },
            { id: 'followup', label: 'Follow-up', icon: '📨' },
            { id: 'ai', label: 'AI Config', icon: '🤖' }
        ];

        const settings = reactive({
            working_hours: {
                enabled: false,
                timezone: 'UTC',
                away_message: '',
                schedule: [
                    { day: 0, is_working: false, start_time: '09:00', end_time: '18:00' },
                    { day: 1, is_working: true, start_time: '09:00', end_time: '18:00' },
                    { day: 2, is_working: true, start_time: '09:00', end_time: '18:00' },
                    { day: 3, is_working: true, start_time: '09:00', end_time: '18:00' },
                    { day: 4, is_working: true, start_time: '09:00', end_time: '18:00' },
                    { day: 5, is_working: true, start_time: '09:00', end_time: '18:00' },
                    { day: 6, is_working: false, start_time: '09:00', end_time: '18:00' }
                ]
            },
            translation: {
                enabled: false,
                source_language: 'en',
                auto_detect: true,
                translate_incoming: true,
                translate_outgoing: true
            },
            follow_up: {
                enabled: false,
                delay_minutes: 30,
                max_follow_ups: 2,
                only_if_no_reply: true,
                messages: ['Hi! Just checking if you have any questions?']
            },
            sentiment: {
                enabled: false,
                alert_on_negative: true,
                negative_threshold: 0.3,
                escalate_on_very_negative: false
            },
            max_tokens_per_msg: 500,
            temperature: 0.7
        });

        const dayNames = ['Sunday', 'Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday'];
        const getDayName = (day) => dayNames[day];

        const loadSettings = async () => {
            try {
                const response = await axios.get(`/api/agents/${props.agentId}/settings`);
                if (response.data.results) {
                    Object.assign(settings, response.data.results);
                }
            } catch (error) {
                console.error('Failed to load settings:', error);
            }
        };

        const saveSettings = async () => {
            saving.value = true;
            try {
                await axios.put(`/api/agents/${props.agentId}/settings`, settings);
                emit('close');
            } catch (error) {
                console.error('Failed to save settings:', error);
            } finally {
                saving.value = false;
            }
        };

        const addFollowUpMessage = () => {
            settings.follow_up.messages.push('');
        };

        const removeFollowUpMessage = (index) => {
            settings.follow_up.messages.splice(index, 1);
        };

        onMounted(loadSettings);

        return {
            activeTab, tabs, settings, saving,
            getDayName, saveSettings, addFollowUpMessage, removeFollowUpMessage
        };
    }
};

// Knowledge Base Component
const KnowledgeBase = {
    props: ['agentId'],
    emits: ['close'],
    
    template: `
    <div class="space-y-6">
        <div class="flex items-center justify-between">
            <h3 class="text-xl font-bold text-white">Knowledge Base</h3>
            <button @click="$emit('close')" class="text-dark-muted hover:text-white">
                <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                </svg>
            </button>
        </div>

        <p class="text-dark-muted">Upload documents for your AI to reference when answering questions.</p>

        <!-- Upload Form -->
        <div class="p-4 bg-dark-bg rounded-xl border-2 border-dashed border-dark-border">
            <div class="space-y-4">
                <div>
                    <label class="block text-sm font-medium text-dark-text mb-2">Document Name</label>
                    <input type="text" v-model="newDoc.name"
                           class="w-full px-4 py-2 bg-dark-card border border-dark-border rounded-lg text-white focus:border-primary-500 focus:outline-none"
                           placeholder="e.g., Company FAQ">
                </div>
                <div>
                    <label class="block text-sm font-medium text-dark-text mb-2">Content</label>
                    <textarea v-model="newDoc.content" rows="6"
                              class="w-full px-4 py-2 bg-dark-card border border-dark-border rounded-lg text-white focus:border-primary-500 focus:outline-none resize-none"
                              placeholder="Paste your document text here..."></textarea>
                </div>
                <button @click="uploadDocument" :disabled="uploading || !newDoc.name || !newDoc.content"
                        class="w-full py-2 bg-primary-600 hover:bg-primary-500 text-white font-medium rounded-xl transition-colors disabled:opacity-50">
                    {{ uploading ? 'Uploading...' : 'Upload Document' }}
                </button>
            </div>
        </div>

        <!-- Documents List -->
        <div class="space-y-3">
            <h4 class="text-sm font-semibold text-dark-muted uppercase tracking-wider">Documents</h4>
            
            <div v-if="documents.length === 0" class="text-center py-8 text-dark-muted">
                No documents uploaded yet
            </div>

            <div v-for="doc in documents" :key="doc.id"
                 class="flex items-center justify-between p-4 bg-dark-bg rounded-xl">
                <div class="flex items-center gap-3">
                    <div class="w-10 h-10 rounded-lg bg-primary-500/20 flex items-center justify-center">
                        📄
                    </div>
                    <div>
                        <p class="text-white font-medium">{{ doc.name }}</p>
                        <p class="text-xs text-dark-muted">
                            {{ formatSize(doc.size) }} • {{ formatStatus(doc.status) }}
                        </p>
                    </div>
                </div>
                <button @click="deleteDocument(doc.id)" class="text-red-400 hover:text-red-300">
                    <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                    </svg>
                </button>
            </div>
        </div>
    </div>
    `,

    setup(props) {
        const { ref, reactive, onMounted } = Vue;

        const documents = ref([]);
        const uploading = ref(false);
        const newDoc = reactive({
            name: '',
            content: ''
        });

        const loadDocuments = async () => {
            try {
                const response = await axios.get(`/api/agents/${props.agentId}/knowledge`);
                documents.value = response.data.results || [];
            } catch (error) {
                console.error('Failed to load documents:', error);
            }
        };

        const uploadDocument = async () => {
            uploading.value = true;
            try {
                await axios.post(`/api/agents/${props.agentId}/knowledge`, {
                    name: newDoc.name,
                    type: 'text',
                    content: newDoc.content
                });
                newDoc.name = '';
                newDoc.content = '';
                await loadDocuments();
            } catch (error) {
                console.error('Failed to upload document:', error);
            } finally {
                uploading.value = false;
            }
        };

        const deleteDocument = async (id) => {
            if (!confirm('Delete this document?')) return;
            try {
                await axios.delete(`/api/knowledge/${id}`);
                await loadDocuments();
            } catch (error) {
                console.error('Failed to delete document:', error);
            }
        };

        const formatSize = (bytes) => {
            if (bytes < 1024) return bytes + ' B';
            if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
            return (bytes / (1024 * 1024)).toFixed(1) + ' MB';
        };

        const formatStatus = (status) => {
            const statusMap = {
                'processing': '⏳ Processing...',
                'ready': '✅ Ready',
                'error': '❌ Error'
            };
            return statusMap[status] || status;
        };

        onMounted(loadDocuments);

        return {
            documents, uploading, newDoc,
            uploadDocument, deleteDocument, formatSize, formatStatus
        };
    }
};

// Broadcast Component
const Broadcast = {
    props: ['agentId'],
    emits: ['close'],
    
    template: `
    <div class="space-y-6">
        <div class="flex items-center justify-between">
            <h3 class="text-xl font-bold text-white">Broadcast Messages</h3>
            <button @click="$emit('close')" class="text-dark-muted hover:text-white">
                <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                </svg>
            </button>
        </div>

        <!-- Create Broadcast -->
        <div class="p-4 bg-dark-bg rounded-xl space-y-4">
            <h4 class="text-white font-medium">New Broadcast</h4>
            
            <div>
                <label class="block text-sm font-medium text-dark-text mb-2">Platform</label>
                <select v-model="newBroadcast.integration_type"
                        class="w-full px-4 py-2 bg-dark-card border border-dark-border rounded-lg text-white focus:border-primary-500 focus:outline-none">
                    <option value="whatsapp">WhatsApp</option>
                    <option value="telegram">Telegram</option>
                </select>
            </div>

            <div>
                <label class="block text-sm font-medium text-dark-text mb-2">Message</label>
                <textarea v-model="newBroadcast.message" rows="3"
                          class="w-full px-4 py-2 bg-dark-card border border-dark-border rounded-lg text-white focus:border-primary-500 focus:outline-none resize-none"
                          placeholder="Your broadcast message..."></textarea>
            </div>

            <div>
                <label class="block text-sm font-medium text-dark-text mb-2">Recipients (one per line)</label>
                <textarea v-model="recipientsText" rows="3"
                          class="w-full px-4 py-2 bg-dark-card border border-dark-border rounded-lg text-white focus:border-primary-500 focus:outline-none resize-none font-mono text-sm"
                          placeholder="79001234567&#10;79001234568&#10;..."></textarea>
            </div>

            <button @click="createBroadcast" :disabled="creating || !newBroadcast.message || !recipientsText"
                    class="w-full py-2 bg-primary-600 hover:bg-primary-500 text-white font-medium rounded-xl transition-colors disabled:opacity-50">
                {{ creating ? 'Creating...' : 'Create Broadcast' }}
            </button>
        </div>

        <!-- Broadcasts List -->
        <div class="space-y-3">
            <h4 class="text-sm font-semibold text-dark-muted uppercase tracking-wider">History</h4>
            
            <div v-if="broadcasts.length === 0" class="text-center py-8 text-dark-muted">
                No broadcasts yet
            </div>

            <div v-for="b in broadcasts" :key="b.id"
                 class="p-4 bg-dark-bg rounded-xl">
                <div class="flex items-start justify-between">
                    <div>
                        <p class="text-white">{{ b.message.substring(0, 50) }}{{ b.message.length > 50 ? '...' : '' }}</p>
                        <p class="text-xs text-dark-muted mt-1">
                            {{ b.integration_type }} • {{ b.total_recipients }} recipients
                        </p>
                    </div>
                    <span :class="['px-2 py-1 rounded text-xs', getStatusClass(b.status)]">
                        {{ b.status }}
                    </span>
                </div>
                <div class="flex items-center gap-4 mt-3">
                    <div class="text-xs text-dark-muted">
                        ✅ {{ b.sent_count }} sent • ❌ {{ b.failed_count }} failed
                    </div>
                    <button v-if="b.status === 'pending'" @click="sendBroadcast(b.id)"
                            class="ml-auto px-3 py-1 bg-primary-600 hover:bg-primary-500 text-white text-xs rounded">
                        Send Now
                    </button>
                </div>
            </div>
        </div>
    </div>
    `,

    setup(props) {
        const { ref, reactive, onMounted, computed } = Vue;

        const broadcasts = ref([]);
        const creating = ref(false);
        const recipientsText = ref('');
        const newBroadcast = reactive({
            integration_type: 'whatsapp',
            message: ''
        });

        const loadBroadcasts = async () => {
            try {
                const response = await axios.get(`/api/agents/${props.agentId}/broadcasts`);
                broadcasts.value = response.data.results || [];
            } catch (error) {
                console.error('Failed to load broadcasts:', error);
            }
        };

        const createBroadcast = async () => {
            creating.value = true;
            try {
                const recipients = recipientsText.value
                    .split('\n')
                    .map(r => r.trim())
                    .filter(r => r);
                
                await axios.post(`/api/agents/${props.agentId}/broadcasts`, {
                    integration_type: newBroadcast.integration_type,
                    message: newBroadcast.message,
                    recipients: recipients
                });
                
                newBroadcast.message = '';
                recipientsText.value = '';
                await loadBroadcasts();
            } catch (error) {
                console.error('Failed to create broadcast:', error);
            } finally {
                creating.value = false;
            }
        };

        const sendBroadcast = async (id) => {
            try {
                await axios.post(`/api/broadcasts/${id}/send`);
                await loadBroadcasts();
            } catch (error) {
                console.error('Failed to send broadcast:', error);
            }
        };

        const getStatusClass = (status) => {
            const classes = {
                'pending': 'bg-yellow-500/20 text-yellow-400',
                'sending': 'bg-blue-500/20 text-blue-400',
                'completed': 'bg-green-500/20 text-green-400',
                'failed': 'bg-red-500/20 text-red-400'
            };
            return classes[status] || 'bg-dark-border text-dark-muted';
        };

        onMounted(loadBroadcasts);

        return {
            broadcasts, creating, recipientsText, newBroadcast,
            createBroadcast, sendBroadcast, getStatusClass
        };
    }
};


