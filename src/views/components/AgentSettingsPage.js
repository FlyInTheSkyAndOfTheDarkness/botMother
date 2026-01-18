// Agent Settings Page Component - Matches reference design
const AgentSettingsPage = {
    props: ['agentId'],
    emits: ['close', 'save'],

    template: `
    <div class="max-w-4xl mx-auto">
        <!-- Header with Avatar and Test Chat button -->
        <div class="flex items-center justify-between mb-8">
            <h1 class="text-2xl font-bold text-content-text">Настройки агента</h1>
            <button @click="$emit('test-chat')" 
                    class="px-4 py-2 bg-white border border-gray-200 text-content-muted text-sm font-medium rounded-lg hover:bg-gray-50 transition-colors">
                Тестовый чат
            </button>
        </div>

        <!-- Settings Form -->
        <div class="space-y-8">
            <!-- Общие настройки -->
            <section class="bg-white rounded-2xl border border-gray-200 p-6">
                <h2 class="text-lg font-semibold text-content-text mb-6">Общие настройки</h2>
                
                <!-- Avatar/Emoji Selector -->
                <div class="mb-6">
                    <button @click="showEmojiPicker = !showEmojiPicker" 
                            class="w-16 h-16 rounded-xl bg-yellow-50 flex items-center justify-center text-3xl hover:bg-yellow-100 transition-colors border-2 border-dashed border-yellow-200">
                        {{ form.avatar || '😍' }}
                    </button>
                    <!-- Simple Emoji Picker -->
                    <div v-if="showEmojiPicker" class="mt-2 p-3 bg-white rounded-xl shadow-lg border border-gray-200 inline-flex gap-2 flex-wrap max-w-xs">
                        <button v-for="emoji in emojis" :key="emoji" 
                                @click="selectEmoji(emoji)"
                                class="w-10 h-10 rounded-lg hover:bg-gray-100 flex items-center justify-center text-xl transition-colors">
                            {{ emoji }}
                        </button>
                    </div>
                </div>

                <!-- Название -->
                <div class="mb-6">
                    <label class="block text-sm font-medium text-content-text mb-2">Название</label>
                    <input v-model="form.name" type="text"
                           class="w-full px-4 py-3 bg-gray-50 border border-gray-200 rounded-xl text-content-text placeholder-content-muted focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 focus:outline-none transition-all"
                           placeholder="Имя агента">
                </div>

                <!-- Статус бота -->
                <div class="flex items-center justify-between p-4 bg-gray-50 rounded-xl mb-4">
                    <div>
                        <p class="font-medium text-content-text">Статус бота</p>
                        <p class="text-sm text-content-muted">Активировать или деактивировать</p>
                    </div>
                    <label class="relative inline-flex items-center cursor-pointer">
                        <input type="checkbox" v-model="form.is_active" class="sr-only peer">
                        <div class="w-14 h-7 bg-gray-300 rounded-full peer peer-checked:bg-primary-500 
                                    after:content-[''] after:absolute after:top-[3px] after:left-[3px] 
                                    after:bg-white after:rounded-full after:h-[22px] after:w-[22px] after:transition-all after:shadow-sm
                                    peer-checked:after:translate-x-7"></div>
                    </label>
                </div>

                <!-- Состояние чата по умолчанию -->
                <div class="flex items-center justify-between p-4 bg-gray-50 rounded-xl mb-4">
                    <div>
                        <p class="font-medium text-content-text">Состояние чата по умолчанию</p>
                        <p class="text-sm text-content-muted">Активировать или деактивировать</p>
                    </div>
                    <label class="relative inline-flex items-center cursor-pointer">
                        <input type="checkbox" v-model="form.auto_reply" class="sr-only peer">
                        <div class="w-14 h-7 bg-gray-300 rounded-full peer peer-checked:bg-primary-500 
                                    after:content-[''] after:absolute after:top-[3px] after:left-[3px] 
                                    after:bg-white after:rounded-full after:h-[22px] after:w-[22px] after:transition-all after:shadow-sm
                                    peer-checked:after:translate-x-7"></div>
                    </label>
                </div>

                <!-- Часовой пояс -->
                <div class="mb-6">
                    <label class="block text-sm font-medium text-content-text mb-2">Часовой пояс</label>
                    <div class="relative">
                        <select v-model="form.timezone"
                                class="w-full px-4 py-3 bg-gray-50 border border-gray-200 rounded-xl text-content-text focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 focus:outline-none transition-all appearance-none">
                            <option value="Asia/Almaty">Asia/Almaty</option>
                            <option value="Europe/Moscow">Europe/Moscow</option>
                            <option value="UTC">UTC</option>
                            <option value="Europe/London">Europe/London</option>
                            <option value="America/New_York">America/New_York</option>
                            <option value="Asia/Tokyo">Asia/Tokyo</option>
                        </select>
                        <svg class="absolute right-4 top-1/2 -translate-y-1/2 w-5 h-5 text-content-muted pointer-events-none" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
                        </svg>
                    </div>
                </div>
            </section>

            <!-- Защита от спама -->
            <section class="bg-white rounded-2xl border border-gray-200 p-6">
                <h2 class="text-lg font-semibold text-content-text mb-6">Защита от спама</h2>
                
                <!-- Защита от спама пользователя -->
                <div class="flex items-center justify-between p-4 bg-gray-50 rounded-xl mb-4">
                    <div>
                        <p class="font-medium text-content-text">Защита от спама пользователя</p>
                        <p class="text-sm text-content-muted">Активировать или деактивировать</p>
                    </div>
                    <label class="relative inline-flex items-center cursor-pointer">
                        <input type="checkbox" v-model="form.spam_protection" class="sr-only peer">
                        <div class="w-14 h-7 bg-gray-300 rounded-full peer peer-checked:bg-primary-500 
                                    after:content-[''] after:absolute after:top-[3px] after:left-[3px] 
                                    after:bg-white after:rounded-full after:h-[22px] after:w-[22px] after:transition-all after:shadow-sm
                                    peer-checked:after:translate-x-7"></div>
                    </label>
                </div>

                <!-- Spam Settings (shown when enabled) -->
                <div v-if="form.spam_protection" class="space-y-4 mt-4 pl-4 border-l-2 border-primary-200">
                    <div>
                        <label class="block text-sm font-medium text-content-text mb-2">Стандартное сообщение</label>
                        <input v-model="form.spam_message" type="text"
                               class="w-full px-4 py-3 bg-gray-50 border border-gray-200 rounded-xl text-content-text placeholder-content-muted focus:border-primary-500 focus:outline-none"
                               placeholder="Введите значение...">
                    </div>
                    <div class="grid grid-cols-2 gap-4">
                        <div>
                            <label class="block text-sm font-medium text-content-text mb-2">Количество</label>
                            <input v-model.number="form.spam_count" type="number" min="1"
                                   class="w-full px-4 py-3 bg-gray-50 border border-gray-200 rounded-xl text-content-text focus:border-primary-500 focus:outline-none"
                                   placeholder="0">
                        </div>
                        <div>
                            <label class="block text-sm font-medium text-content-text mb-2">Длительность</label>
                            <input v-model.number="form.spam_duration" type="number" min="1"
                                   class="w-full px-4 py-3 bg-gray-50 border border-gray-200 rounded-xl text-content-text focus:border-primary-500 focus:outline-none"
                                   placeholder="0">
                        </div>
                    </div>
                </div>
            </section>

            <!-- Save Button -->
            <div class="flex justify-end">
                <button @click="saveSettings" :disabled="saving"
                        class="px-8 py-3 bg-primary-500 hover:bg-primary-600 text-white font-medium rounded-xl transition-colors disabled:opacity-50 shadow-lg shadow-primary-500/25">
                    {{ saving ? 'Сохранение...' : 'Сохранить настройки' }}
                </button>
            </div>
        </div>
    </div>
    `,

    setup(props, { emit }) {
        const { ref, reactive, onMounted } = Vue;

        const saving = ref(false);
        const showEmojiPicker = ref(false);

        const emojis = ['😍', '🤖', '💼', '🎯', '🚀', '💡', '🌟', '🔥', '💬', '📱', '🎨', '🛒'];

        const form = reactive({
            name: '',
            avatar: '😍',
            is_active: true,
            auto_reply: true,
            timezone: 'Asia/Almaty',
            spam_protection: false,
            spam_message: '',
            spam_count: 0,
            spam_duration: 0
        });

        const selectEmoji = (emoji) => {
            form.avatar = emoji;
            showEmojiPicker.value = false;
        };

        const loadSettings = async () => {
            if (!props.agentId) return;
            try {
                const response = await axios.get(`/api/agents/${props.agentId}`);
                const agent = response.data.results;
                if (agent) {
                    form.name = agent.name || '';
                    form.avatar = agent.avatar || '😍';
                    form.is_active = agent.is_active ?? true;
                    form.timezone = agent.timezone || 'Asia/Almaty';
                }

                // Load additional settings
                try {
                    const settingsRes = await axios.get(`/api/agents/${props.agentId}/settings`);
                    const settings = settingsRes.data.results;
                    if (settings) {
                        form.auto_reply = settings.auto_reply ?? true;
                        form.spam_protection = settings.spam_protection?.enabled ?? false;
                        form.spam_message = settings.spam_protection?.message || '';
                        form.spam_count = settings.spam_protection?.count || 0;
                        form.spam_duration = settings.spam_protection?.duration || 0;
                    }
                } catch (e) {
                    console.log('Settings not found, using defaults');
                }
            } catch (error) {
                console.error('Failed to load agent:', error);
            }
        };

        const saveSettings = async () => {
            saving.value = true;
            try {
                // Update agent basic info
                await axios.put(`/api/agents/${props.agentId}`, {
                    name: form.name,
                    is_active: form.is_active,
                    avatar: form.avatar,
                    timezone: form.timezone
                });

                // Update agent settings
                await axios.put(`/api/agents/${props.agentId}/settings`, {
                    auto_reply: form.auto_reply,
                    spam_protection: {
                        enabled: form.spam_protection,
                        message: form.spam_message,
                        count: form.spam_count,
                        duration: form.spam_duration
                    }
                });

                emit('save');
            } catch (error) {
                console.error('Failed to save settings:', error);
            } finally {
                saving.value = false;
            }
        };

        onMounted(loadSettings);

        return {
            form,
            saving,
            showEmojiPicker,
            emojis,
            selectEmoji,
            saveSettings
        };
    }
};
