// Sidebar Component - Hierarchical Navigation
const Sidebar = {
    template: `
    <aside class="sidebar-container w-64 h-screen bg-sidebar-bg border-r border-sidebar-border flex flex-col fixed left-0 top-0 z-50">
        <!-- User Profile -->
        <div class="p-4 border-b border-sidebar-border">
            <div class="flex items-center gap-3">
                <div class="w-10 h-10 rounded-full bg-gradient-to-br from-purple-500 to-pink-500 flex items-center justify-center text-white font-semibold">
                    {{ userInitials }}
                </div>
                <div class="flex-1 min-w-0">
                    <p class="text-white font-medium text-sm truncate">{{ userName }}</p>
                    <p class="text-sidebar-muted text-xs truncate">{{ userEmail }}</p>
                </div>
            </div>
        </div>

        <!-- Quick Links -->
        <div class="p-2 border-b border-sidebar-border">
            <button @click="$emit('navigate', 'notifications')" 
                    class="w-full flex items-center gap-3 px-3 py-2 text-sidebar-text hover:bg-sidebar-hover rounded-lg transition-colors text-sm">
                <span>🔔</span>
                <span>Уведомления</span>
            </button>
            <button @click="$emit('navigate', 'billing')" 
                    class="w-full flex items-center gap-3 px-3 py-2 text-sidebar-text hover:bg-sidebar-hover rounded-lg transition-colors text-sm">
                <span>💳</span>
                <span>Биллинг</span>
            </button>
        </div>

        <!-- Navigation Groups -->
        <nav class="flex-1 overflow-y-auto p-2">
            <!-- Рабочее пространство -->
            <div class="mb-2">
                <button @click="toggleGroup('workspace')" 
                        class="w-full flex items-center justify-between px-3 py-2 text-sidebar-muted hover:text-white text-xs font-semibold uppercase tracking-wider">
                    <span>Рабочее пространство</span>
                    <svg :class="['w-4 h-4 transition-transform', expandedGroups.workspace ? 'rotate-180' : '']" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
                    </svg>
                </button>
                <div v-show="expandedGroups.workspace" class="space-y-1">
                    <button @click="$emit('navigate', 'dashboard')" 
                            :class="['w-full flex items-center gap-3 px-3 py-2 rounded-lg transition-colors text-sm',
                                     activeView === 'dashboard' ? 'bg-purple-500/20 text-purple-400' : 'text-sidebar-text hover:bg-sidebar-hover']">
                        <span>📊</span>
                        <span>Дашборд</span>
                    </button>
                </div>
            </div>

            <!-- ИИ-Агенты -->
            <div class="mb-2">
                <button @click="toggleGroup('agents')" 
                        class="w-full flex items-center justify-between px-3 py-2 text-sidebar-muted hover:text-white text-xs font-semibold uppercase tracking-wider">
                    <div class="flex items-center gap-2">
                        <span>ИИ-Агенты</span>
                        <button @click.stop="$emit('create-agent')" class="text-purple-400 hover:text-purple-300">
                            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
                            </svg>
                        </button>
                    </div>
                    <svg :class="['w-4 h-4 transition-transform', expandedGroups.agents ? 'rotate-180' : '']" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
                    </svg>
                </button>
                <div v-show="expandedGroups.agents" class="space-y-1">
                    <!-- Agent List -->
                    <div v-for="agent in agents" :key="agent.id" class="ml-2">
                        <button @click="toggleAgent(agent.id)" 
                                :class="['w-full flex items-center justify-between px-3 py-2 rounded-lg transition-colors text-sm',
                                         selectedAgentId === agent.id ? 'bg-purple-500/20 text-purple-400' : 'text-sidebar-text hover:bg-sidebar-hover']">
                            <div class="flex items-center gap-2 min-w-0">
                                <span class="text-lg">🤖</span>
                                <span class="truncate">{{ agent.name }}</span>
                            </div>
                            <svg :class="['w-4 h-4 transition-transform flex-shrink-0', expandedAgents[agent.id] ? 'rotate-180' : '']" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
                            </svg>
                        </button>
                        
                        <!-- Agent Sub-menu -->
                        <div v-show="expandedAgents[agent.id]" class="ml-6 space-y-1 mt-1">
                            <button @click="$emit('agent-action', { agentId: agent.id, action: 'settings' })"
                                    :class="['w-full flex items-center gap-2 px-3 py-1.5 rounded-lg transition-colors text-xs',
                                             activeSubView === 'settings' && selectedAgentId === agent.id ? 'bg-purple-500/10 text-purple-400' : 'text-sidebar-muted hover:bg-sidebar-hover hover:text-white']">
                                <span>⚙️</span>
                                <span>Настройки</span>
                            </button>
                            <button @click="$emit('agent-action', { agentId: agent.id, action: 'prompting' })"
                                    :class="['w-full flex items-center gap-2 px-3 py-1.5 rounded-lg transition-colors text-xs',
                                             activeSubView === 'prompting' && selectedAgentId === agent.id ? 'bg-purple-500/10 text-purple-400' : 'text-sidebar-muted hover:bg-sidebar-hover hover:text-white']">
                                <span>💬</span>
                                <span>Промптинг</span>
                            </button>
                            <button @click="$emit('agent-action', { agentId: agent.id, action: 'messages' })"
                                    :class="['w-full flex items-center gap-2 px-3 py-1.5 rounded-lg transition-colors text-xs',
                                             activeSubView === 'messages' && selectedAgentId === agent.id ? 'bg-purple-500/10 text-purple-400' : 'text-sidebar-muted hover:bg-sidebar-hover hover:text-white']">
                                <span>📝</span>
                                <span>Сообщения</span>
                            </button>
                            <button @click="$emit('agent-action', { agentId: agent.id, action: 'llm' })"
                                    :class="['w-full flex items-center gap-2 px-3 py-1.5 rounded-lg transition-colors text-xs',
                                             activeSubView === 'llm' && selectedAgentId === agent.id ? 'bg-purple-500/10 text-purple-400' : 'text-sidebar-muted hover:bg-sidebar-hover hover:text-white']">
                                <span>🧠</span>
                                <span>LLM-Модели</span>
                            </button>
                            <button @click="$emit('agent-action', { agentId: agent.id, action: 'control' })"
                                    :class="['w-full flex items-center gap-2 px-3 py-1.5 rounded-lg transition-colors text-xs',
                                             activeSubView === 'control' && selectedAgentId === agent.id ? 'bg-purple-500/10 text-purple-400' : 'text-sidebar-muted hover:bg-sidebar-hover hover:text-white']">
                                <span>🎛️</span>
                                <span>Контроль</span>
                            </button>
                            <button @click="$emit('agent-action', { agentId: agent.id, action: 'functions' })"
                                    :class="['w-full flex items-center gap-2 px-3 py-1.5 rounded-lg transition-colors text-xs',
                                             activeSubView === 'functions' && selectedAgentId === agent.id ? 'bg-purple-500/10 text-purple-400' : 'text-sidebar-muted hover:bg-sidebar-hover hover:text-white']">
                                <span>⚡</span>
                                <span>Функции</span>
                            </button>
                            <button @click="$emit('agent-action', { agentId: agent.id, action: 'knowledge' })"
                                    :class="['w-full flex items-center gap-2 px-3 py-1.5 rounded-lg transition-colors text-xs',
                                             activeSubView === 'knowledge' && selectedAgentId === agent.id ? 'bg-purple-500/10 text-purple-400' : 'text-sidebar-muted hover:bg-sidebar-hover hover:text-white']">
                                <span>📚</span>
                                <span>База знаний</span>
                            </button>
                            <button @click="$emit('agent-action', { agentId: agent.id, action: 'integrations' })"
                                    :class="['w-full flex items-center gap-2 px-3 py-1.5 rounded-lg transition-colors text-xs',
                                             activeSubView === 'integrations' && selectedAgentId === agent.id ? 'bg-purple-500/10 text-purple-400' : 'text-sidebar-muted hover:bg-sidebar-hover hover:text-white']">
                                <span>🔗</span>
                                <span>Интеграции</span>
                            </button>
                            <button @click="$emit('agent-action', { agentId: agent.id, action: 'channels' })"
                                    :class="['w-full flex items-center gap-2 px-3 py-1.5 rounded-lg transition-colors text-xs',
                                             activeSubView === 'channels' && selectedAgentId === agent.id ? 'bg-purple-500/10 text-purple-400' : 'text-sidebar-muted hover:bg-sidebar-hover hover:text-white']">
                                <span>📡</span>
                                <span>Каналы</span>
                            </button>
                            <button @click="$emit('agent-action', { agentId: agent.id, action: 'test-chat' })"
                                    :class="['w-full flex items-center gap-2 px-3 py-1.5 rounded-lg transition-colors text-xs',
                                             activeSubView === 'test-chat' && selectedAgentId === agent.id ? 'bg-purple-500/10 text-purple-400' : 'text-sidebar-muted hover:bg-sidebar-hover hover:text-white']">
                                <span>💭</span>
                                <span>Тестовый чат</span>
                            </button>
                        </div>
                    </div>
                </div>
            </div>

            <!-- Работа с клиентами -->
            <div class="mb-2">
                <button @click="toggleGroup('clients')" 
                        class="w-full flex items-center justify-between px-3 py-2 text-sidebar-muted hover:text-white text-xs font-semibold uppercase tracking-wider">
                    <span>Работа с клиентами</span>
                    <svg :class="['w-4 h-4 transition-transform', expandedGroups.clients ? 'rotate-180' : '']" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
                    </svg>
                </button>
                <div v-show="expandedGroups.clients" class="space-y-1">
                    <button @click="$emit('navigate', 'broadcasts')" 
                            :class="['w-full flex items-center gap-3 px-3 py-2 rounded-lg transition-colors text-sm',
                                     activeView === 'broadcasts' ? 'bg-purple-500/20 text-purple-400' : 'text-sidebar-text hover:bg-sidebar-hover']">
                        <span>📤</span>
                        <span>Рассылки</span>
                    </button>
                    <button @click="$emit('navigate', 'dialogs')" 
                            :class="['w-full flex items-center gap-3 px-3 py-2 rounded-lg transition-colors text-sm',
                                     activeView === 'dialogs' ? 'bg-purple-500/20 text-purple-400' : 'text-sidebar-text hover:bg-sidebar-hover']">
                        <span>💬</span>
                        <span>Диалоги</span>
                    </button>
                    <button @click="$emit('navigate', 'analytics')" 
                            :class="['w-full flex items-center gap-3 px-3 py-2 rounded-lg transition-colors text-sm',
                                     activeView === 'analytics' ? 'bg-purple-500/20 text-purple-400' : 'text-sidebar-text hover:bg-sidebar-hover']">
                        <span>📊</span>
                        <span>Аналитика</span>
                    </button>
                </div>
            </div>
        </nav>
    </aside>
    `,

    props: {
        agents: {
            type: Array,
            default: () => []
        },
        activeView: {
            type: String,
            default: 'dashboard'
        },
        activeSubView: {
            type: String,
            default: ''
        },
        selectedAgentId: {
            type: String,
            default: null
        },
        userName: {
            type: String,
            default: 'User'
        },
        userEmail: {
            type: String,
            default: 'user@example.com'
        }
    },

    emits: ['navigate', 'create-agent', 'agent-action'],

    setup(props) {
        const { ref, reactive, computed } = Vue;

        const expandedGroups = reactive({
            workspace: true,
            agents: true,
            clients: true
        });

        const expandedAgents = reactive({});

        const userInitials = computed(() => {
            if (!props.userName) return 'U';
            return props.userName.split(' ')
                .map(n => n[0])
                .slice(0, 2)
                .join('')
                .toUpperCase();
        });

        const toggleGroup = (group) => {
            expandedGroups[group] = !expandedGroups[group];
        };

        const toggleAgent = (agentId) => {
            expandedAgents[agentId] = !expandedAgents[agentId];
        };

        return {
            expandedGroups,
            expandedAgents,
            userInitials,
            toggleGroup,
            toggleAgent
        };
    }
};
