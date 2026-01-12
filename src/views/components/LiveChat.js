// Live Chat Component - View all conversations in real-time
const LiveChat = {
    props: ['agentId'],
    
    template: `
    <div class="flex h-[calc(100vh-12rem)]">
        <!-- Conversations List -->
        <div class="w-80 bg-dark-card border-r border-dark-border flex flex-col">
            <div class="p-4 border-b border-dark-border">
                <input v-model="searchQuery" type="text" placeholder="Search conversations..."
                       class="w-full px-3 py-2 bg-dark-bg border border-dark-border rounded-lg text-white text-sm focus:border-primary-500 focus:outline-none">
            </div>
            
            <div class="flex-1 overflow-y-auto">
                <div v-if="filteredConversations.length === 0" class="p-4 text-center text-dark-muted">
                    No conversations yet
                </div>
                
                <div v-for="conv in filteredConversations" :key="conv.id"
                     @click="selectConversation(conv)"
                     :class="['p-4 border-b border-dark-border/50 cursor-pointer transition-colors',
                              selectedConversation?.id === conv.id ? 'bg-primary-500/10 border-l-2 border-l-primary-500' : 'hover:bg-dark-border/30']">
                    <div class="flex items-center gap-3">
                        <div class="w-10 h-10 rounded-full bg-dark-border flex items-center justify-center">
                            <span class="text-lg">👤</span>
                        </div>
                        <div class="flex-1 min-w-0">
                            <div class="flex items-center justify-between">
                                <p class="text-white font-medium truncate">{{ formatJID(conv.remote_jid) }}</p>
                                <span class="text-xs text-dark-muted">{{ formatTime(conv.updated_at) }}</span>
                            </div>
                            <p class="text-sm text-dark-muted truncate">{{ conv.last_message || 'No messages' }}</p>
                        </div>
                        <div v-if="conv.unread_count" class="w-5 h-5 rounded-full bg-primary-500 text-white text-xs flex items-center justify-center">
                            {{ conv.unread_count }}
                        </div>
                    </div>
                </div>
            </div>
        </div>

        <!-- Chat Area -->
        <div class="flex-1 flex flex-col bg-dark-bg">
            <template v-if="selectedConversation">
                <!-- Chat Header -->
                <div class="p-4 bg-dark-card border-b border-dark-border flex items-center justify-between">
                    <div class="flex items-center gap-3">
                        <div class="w-10 h-10 rounded-full bg-dark-border flex items-center justify-center">
                            <span class="text-lg">👤</span>
                        </div>
                        <div>
                            <p class="text-white font-medium">{{ formatJID(selectedConversation.remote_jid) }}</p>
                            <p class="text-xs text-dark-muted">{{ selectedConversation.integration_type || 'WhatsApp' }}</p>
                        </div>
                    </div>
                    <div class="flex items-center gap-2">
                        <button @click="takeOver" v-if="!isManualMode"
                                class="px-3 py-1.5 bg-yellow-600 hover:bg-yellow-500 text-white text-sm rounded-lg transition-colors">
                            Take Over
                        </button>
                        <button @click="releaseControl" v-else
                                class="px-3 py-1.5 bg-green-600 hover:bg-green-500 text-white text-sm rounded-lg transition-colors">
                            Release to AI
                        </button>
                    </div>
                </div>

                <!-- Messages -->
                <div class="flex-1 overflow-y-auto p-4 space-y-4" ref="messagesContainer">
                    <div v-for="msg in messages" :key="msg.id"
                         :class="['flex', msg.role === 'user' ? 'justify-start' : 'justify-end']">
                        <div :class="['max-w-[70%] rounded-2xl px-4 py-2',
                                      msg.role === 'user' ? 'bg-dark-card text-white rounded-bl-md' : 'bg-primary-600 text-white rounded-br-md']">
                            <p class="text-sm whitespace-pre-wrap">{{ msg.content }}</p>
                            <p :class="['text-xs mt-1', msg.role === 'user' ? 'text-dark-muted' : 'text-primary-200']">
                                {{ formatMessageTime(msg.timestamp) }}
                                <span v-if="msg.role === 'assistant'" class="ml-1">
                                    {{ msg.is_manual ? '👤' : '🤖' }}
                                </span>
                            </p>
                        </div>
                    </div>
                </div>

                <!-- Input (only in manual mode) -->
                <div v-if="isManualMode" class="p-4 bg-dark-card border-t border-dark-border">
                    <div class="flex gap-2">
                        <input v-model="newMessage" type="text" placeholder="Type a message..."
                               @keyup.enter="sendMessage"
                               class="flex-1 px-4 py-2 bg-dark-bg border border-dark-border rounded-lg text-white focus:border-primary-500 focus:outline-none">
                        <button @click="sendMessage" :disabled="!newMessage.trim() || sending"
                                class="px-4 py-2 bg-primary-600 hover:bg-primary-500 text-white rounded-lg transition-colors disabled:opacity-50">
                            {{ sending ? '...' : 'Send' }}
                        </button>
                    </div>
                    <p class="text-xs text-yellow-400 mt-2">⚠️ Manual mode: AI is paused for this conversation</p>
                </div>
            </template>

            <!-- No conversation selected -->
            <div v-else class="flex-1 flex items-center justify-center">
                <div class="text-center">
                    <div class="w-16 h-16 rounded-full bg-dark-card flex items-center justify-center mx-auto mb-4">
                        <span class="text-3xl">💬</span>
                    </div>
                    <p class="text-white font-medium">Select a conversation</p>
                    <p class="text-sm text-dark-muted">Choose from the list to view messages</p>
                </div>
            </div>
        </div>

        <!-- Conversation Info -->
        <div v-if="selectedConversation" class="w-64 bg-dark-card border-l border-dark-border p-4">
            <h3 class="text-sm font-semibold text-dark-muted uppercase tracking-wider mb-4">Info</h3>
            
            <div class="space-y-4">
                <div>
                    <p class="text-xs text-dark-muted">Contact</p>
                    <p class="text-white text-sm">{{ formatJID(selectedConversation.remote_jid) }}</p>
                </div>
                <div>
                    <p class="text-xs text-dark-muted">Agent</p>
                    <p class="text-white text-sm">{{ selectedConversation.agent_name || 'Unknown' }}</p>
                </div>
                <div>
                    <p class="text-xs text-dark-muted">Started</p>
                    <p class="text-white text-sm">{{ formatDate(selectedConversation.created_at) }}</p>
                </div>
                <div>
                    <p class="text-xs text-dark-muted">Messages</p>
                    <p class="text-white text-sm">{{ messages.length }}</p>
                </div>
                <div>
                    <p class="text-xs text-dark-muted">Status</p>
                    <p :class="['text-sm', isManualMode ? 'text-yellow-400' : 'text-green-400']">
                        {{ isManualMode ? '👤 Manual Control' : '🤖 AI Active' }}
                    </p>
                </div>
            </div>

            <div class="mt-6 pt-4 border-t border-dark-border">
                <h3 class="text-sm font-semibold text-dark-muted uppercase tracking-wider mb-3">Quick Actions</h3>
                <div class="space-y-2">
                    <button class="w-full px-3 py-2 bg-dark-bg hover:bg-dark-border text-white text-sm rounded-lg transition-colors text-left">
                        📋 Add Note
                    </button>
                    <button class="w-full px-3 py-2 bg-dark-bg hover:bg-dark-border text-white text-sm rounded-lg transition-colors text-left">
                        🏷️ Add Tag
                    </button>
                    <button class="w-full px-3 py-2 bg-dark-bg hover:bg-dark-border text-white text-sm rounded-lg transition-colors text-left">
                        📤 Export Chat
                    </button>
                </div>
            </div>
        </div>
    </div>
    `,

    setup(props) {
        const { ref, computed, onMounted, watch, nextTick } = Vue;

        const conversations = ref([]);
        const selectedConversation = ref(null);
        const messages = ref([]);
        const searchQuery = ref('');
        const newMessage = ref('');
        const sending = ref(false);
        const isManualMode = ref(false);
        const messagesContainer = ref(null);

        const filteredConversations = computed(() => {
            if (!searchQuery.value) return conversations.value;
            const query = searchQuery.value.toLowerCase();
            return conversations.value.filter(c => 
                c.remote_jid?.toLowerCase().includes(query) ||
                c.last_message?.toLowerCase().includes(query)
            );
        });

        const loadConversations = async () => {
            try {
                let url = '/api/conversations';
                if (props.agentId) {
                    url += `?agent_id=${props.agentId}`;
                }
                const response = await axios.get(url);
                conversations.value = response.data.results || [];
            } catch (error) {
                console.error('Failed to load conversations:', error);
                conversations.value = [];
            }
        };

        const selectConversation = async (conv) => {
            selectedConversation.value = conv;
            await loadMessages(conv.id);
        };

        const loadMessages = async (conversationId) => {
            try {
                const response = await axios.get(`/api/conversations/${conversationId}/messages`);
                messages.value = response.data.results || [];
                await nextTick();
                scrollToBottom();
            } catch (error) {
                console.error('Failed to load messages:', error);
                messages.value = [];
            }
        };

        const scrollToBottom = () => {
            if (messagesContainer.value) {
                messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight;
            }
        };

        const sendMessage = async () => {
            if (!newMessage.value.trim() || !selectedConversation.value) return;
            
            sending.value = true;
            try {
                await axios.post(`/api/conversations/${selectedConversation.value.id}/messages`, {
                    content: newMessage.value,
                    is_manual: true
                });
                newMessage.value = '';
                await loadMessages(selectedConversation.value.id);
            } catch (error) {
                console.error('Failed to send message:', error);
            } finally {
                sending.value = false;
            }
        };

        const takeOver = () => {
            isManualMode.value = true;
            // TODO: API call to pause AI for this conversation
        };

        const releaseControl = () => {
            isManualMode.value = false;
            // TODO: API call to resume AI for this conversation
        };

        const formatJID = (jid) => {
            if (!jid) return 'Unknown';
            return jid.replace('@s.whatsapp.net', '').replace('@g.us', ' (Group)');
        };

        const formatTime = (timestamp) => {
            if (!timestamp) return '';
            const date = new Date(timestamp);
            const now = new Date();
            const diff = now - date;
            
            if (diff < 60000) return 'Now';
            if (diff < 3600000) return Math.floor(diff / 60000) + 'm';
            if (diff < 86400000) return Math.floor(diff / 3600000) + 'h';
            return date.toLocaleDateString();
        };

        const formatMessageTime = (timestamp) => {
            if (!timestamp) return '';
            return new Date(timestamp).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
        };

        const formatDate = (timestamp) => {
            if (!timestamp) return '';
            return new Date(timestamp).toLocaleString();
        };

        onMounted(() => {
            loadConversations();
            // Refresh every 30 seconds
            setInterval(loadConversations, 30000);
        });

        watch(selectedConversation, () => {
            isManualMode.value = false;
        });

        return {
            conversations, selectedConversation, messages, searchQuery,
            newMessage, sending, isManualMode, messagesContainer,
            filteredConversations,
            loadConversations, selectConversation, sendMessage,
            takeOver, releaseControl,
            formatJID, formatTime, formatMessageTime, formatDate
        };
    }
};

