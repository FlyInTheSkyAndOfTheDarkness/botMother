// Dashboard Component
const Dashboard = {
    template: `
    <div class="space-y-6">
        <!-- Header -->
        <div class="flex items-center justify-between">
            <div>
                <h2 class="text-2xl font-bold text-white">Dashboard</h2>
                <p class="text-dark-muted">Overview of your AI agents performance</p>
            </div>
            <select v-model="period" @change="loadData" 
                    class="px-4 py-2 bg-dark-card border border-dark-border rounded-lg text-white text-sm focus:border-primary-500 focus:outline-none">
                <option value="today">Today</option>
                <option value="7days">Last 7 days</option>
                <option value="30days">Last 30 days</option>
                <option value="month">This month</option>
            </select>
        </div>

        <!-- Stats Cards -->
        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
            <div class="glass rounded-xl p-5">
                <div class="flex items-center justify-between">
                    <div>
                        <p class="text-dark-muted text-sm">Total Agents</p>
                        <p class="text-3xl font-bold text-white mt-1">{{ stats.total_agents || 0 }}</p>
                    </div>
                    <div class="w-12 h-12 rounded-xl bg-primary-500/20 flex items-center justify-center">
                        <span class="text-2xl">🤖</span>
                    </div>
                </div>
                <p class="text-xs text-green-400 mt-2">{{ stats.active_agents || 0 }} active</p>
            </div>

            <div class="glass rounded-xl p-5">
                <div class="flex items-center justify-between">
                    <div>
                        <p class="text-dark-muted text-sm">Messages Today</p>
                        <p class="text-3xl font-bold text-white mt-1">{{ stats.messages_today || 0 }}</p>
                    </div>
                    <div class="w-12 h-12 rounded-xl bg-green-500/20 flex items-center justify-center">
                        <span class="text-2xl">💬</span>
                    </div>
                </div>
                <p class="text-xs text-dark-muted mt-2">{{ stats.total_messages || 0 }} total</p>
            </div>

            <div class="glass rounded-xl p-5">
                <div class="flex items-center justify-between">
                    <div>
                        <p class="text-dark-muted text-sm">Conversations</p>
                        <p class="text-3xl font-bold text-white mt-1">{{ stats.total_conversations || 0 }}</p>
                    </div>
                    <div class="w-12 h-12 rounded-xl bg-purple-500/20 flex items-center justify-center">
                        <span class="text-2xl">👥</span>
                    </div>
                </div>
                <p class="text-xs text-dark-muted mt-2">{{ stats.active_chats || 0 }} active now</p>
            </div>

            <div class="glass rounded-xl p-5">
                <div class="flex items-center justify-between">
                    <div>
                        <p class="text-dark-muted text-sm">This Week</p>
                        <p class="text-3xl font-bold text-white mt-1">{{ stats.messages_this_week || 0 }}</p>
                    </div>
                    <div class="w-12 h-12 rounded-xl bg-yellow-500/20 flex items-center justify-center">
                        <span class="text-2xl">📈</span>
                    </div>
                </div>
                <p class="text-xs text-dark-muted mt-2">messages</p>
            </div>
        </div>

        <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
            <!-- Chart -->
            <div class="lg:col-span-2 glass rounded-xl p-5">
                <h3 class="text-lg font-semibold text-white mb-4">Messages Over Time</h3>
                <div class="h-64 flex items-end gap-1">
                    <div v-for="(item, index) in chartData" :key="index"
                         class="flex-1 bg-primary-500/30 hover:bg-primary-500/50 rounded-t transition-all relative group"
                         :style="{ height: getBarHeight(item.count) + '%' }">
                        <div class="absolute bottom-full mb-2 left-1/2 -translate-x-1/2 bg-dark-card px-2 py-1 rounded text-xs text-white opacity-0 group-hover:opacity-100 transition-opacity whitespace-nowrap">
                            {{ item.count }} messages
                            <br>
                            <span class="text-dark-muted">{{ formatChartLabel(item) }}</span>
                        </div>
                    </div>
                </div>
                <div class="flex justify-between mt-2 text-xs text-dark-muted">
                    <span>{{ chartData[0]?.date || chartData[0]?.hour + ':00' || '' }}</span>
                    <span>{{ chartData[chartData.length - 1]?.date || (chartData[chartData.length - 1]?.hour + ':00') || '' }}</span>
                </div>
            </div>

            <!-- Recent Activity -->
            <div class="glass rounded-xl p-5">
                <h3 class="text-lg font-semibold text-white mb-4">Recent Activity</h3>
                <div class="space-y-3 max-h-64 overflow-y-auto">
                    <div v-if="activity.length === 0" class="text-center py-8 text-dark-muted">
                        No activity yet
                    </div>
                    <div v-for="item in activity" :key="item.id"
                         class="flex items-start gap-3 p-2 hover:bg-dark-border/30 rounded-lg transition-colors">
                        <div :class="['w-8 h-8 rounded-lg flex items-center justify-center text-sm',
                                      item.type === 'message_received' ? 'bg-blue-500/20' : 'bg-green-500/20']">
                            {{ item.type === 'message_received' ? '📥' : '📤' }}
                        </div>
                        <div class="flex-1 min-w-0">
                            <p class="text-sm text-white truncate">{{ item.description }}</p>
                            <p class="text-xs text-dark-muted">
                                {{ item.agent_name || 'Unknown' }} • {{ formatTime(item.timestamp) }}
                            </p>
                        </div>
                    </div>
                </div>
            </div>
        </div>

        <!-- Agent Performance -->
        <div class="glass rounded-xl p-5">
            <h3 class="text-lg font-semibold text-white mb-4">Agent Performance</h3>
            <div class="overflow-x-auto">
                <table class="w-full">
                    <thead>
                        <tr class="text-left text-dark-muted text-sm border-b border-dark-border">
                            <th class="pb-3 font-medium">Agent</th>
                            <th class="pb-3 font-medium">Conversations</th>
                            <th class="pb-3 font-medium">Messages</th>
                            <th class="pb-3 font-medium">Received</th>
                            <th class="pb-3 font-medium">Sent</th>
                            <th class="pb-3 font-medium">Response Rate</th>
                        </tr>
                    </thead>
                    <tbody>
                        <tr v-if="agentStats.length === 0">
                            <td colspan="6" class="py-8 text-center text-dark-muted">No agents yet</td>
                        </tr>
                        <tr v-for="agent in agentStats" :key="agent.agent_id" 
                            class="border-b border-dark-border/50 hover:bg-dark-border/20">
                            <td class="py-3">
                                <div class="flex items-center gap-2">
                                    <div class="w-8 h-8 rounded-lg bg-primary-500/20 flex items-center justify-center">
                                        🤖
                                    </div>
                                    <span class="text-white">{{ agent.agent_name || 'Unnamed' }}</span>
                                </div>
                            </td>
                            <td class="py-3 text-white">{{ agent.total_conversations || 0 }}</td>
                            <td class="py-3 text-white">{{ agent.total_messages || 0 }}</td>
                            <td class="py-3 text-blue-400">{{ agent.messages_received || 0 }}</td>
                            <td class="py-3 text-green-400">{{ agent.messages_sent || 0 }}</td>
                            <td class="py-3">
                                <div class="flex items-center gap-2">
                                    <div class="flex-1 h-2 bg-dark-border rounded-full overflow-hidden">
                                        <div class="h-full bg-primary-500 rounded-full" 
                                             :style="{ width: getResponseRate(agent) + '%' }"></div>
                                    </div>
                                    <span class="text-xs text-dark-muted">{{ getResponseRate(agent) }}%</span>
                                </div>
                            </td>
                        </tr>
                    </tbody>
                </table>
            </div>
        </div>
    </div>
    `,

    setup() {
        const { ref, onMounted } = Vue;

        const period = ref('7days');
        const stats = ref({});
        const chartData = ref([]);
        const activity = ref([]);
        const agentStats = ref([]);

        const loadData = async () => {
            try {
                // Load dashboard stats
                const dashRes = await axios.get('/api/analytics/dashboard');
                stats.value = dashRes.data.results || {};

                // Load chart data
                const chartRes = await axios.get(`/api/analytics/messages/daily?period=${period.value}`);
                chartData.value = chartRes.data.results || [];

                // Load activity
                const actRes = await axios.get('/api/analytics/activity?limit=10');
                activity.value = actRes.data.results || [];

                // Load agent stats
                const agentRes = await axios.get(`/api/analytics/agents?period=${period.value}`);
                agentStats.value = agentRes.data.results || [];
            } catch (error) {
                console.error('Failed to load analytics:', error);
            }
        };

        const getBarHeight = (count) => {
            if (!chartData.value.length) return 0;
            const max = Math.max(...chartData.value.map(d => d.count || 0));
            if (max === 0) return 5;
            return Math.max(5, (count / max) * 100);
        };

        const formatChartLabel = (item) => {
            if (item.date) return item.date;
            if (item.hour !== undefined) return item.hour + ':00';
            return '';
        };

        const formatTime = (timestamp) => {
            if (!timestamp) return '';
            const date = new Date(timestamp);
            const now = new Date();
            const diff = now - date;
            
            if (diff < 60000) return 'Just now';
            if (diff < 3600000) return Math.floor(diff / 60000) + 'm ago';
            if (diff < 86400000) return Math.floor(diff / 3600000) + 'h ago';
            return date.toLocaleDateString();
        };

        const getResponseRate = (agent) => {
            if (!agent.messages_received) return 0;
            return Math.round((agent.messages_sent / agent.messages_received) * 100);
        };

        onMounted(() => {
            loadData();
        });

        return {
            period, stats, chartData, activity, agentStats,
            loadData, getBarHeight, formatChartLabel, formatTime, getResponseRate
        };
    }
};

