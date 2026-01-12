export default {
    name: 'AIChatbotConfig',
    data() {
        return {
            loading: false,
            saving: false,
            config: {
                enabled: false,
                api_token: '',
                model: 'gpt-4o-mini',
                system_prompt: 'You are a helpful assistant. Respond concisely and helpfully to user messages.'
            },
            availableModels: [
                { value: 'gpt-4o', name: 'GPT-4o (Лучший)' },
                { value: 'gpt-4o-mini', name: 'GPT-4o Mini (Быстрый)' },
                { value: 'gpt-4-turbo', name: 'GPT-4 Turbo' },
                { value: 'gpt-4', name: 'GPT-4' },
                { value: 'gpt-3.5-turbo', name: 'GPT-3.5 Turbo (Дешевый)' },
                { value: 'o1-preview', name: 'o1-preview (Reasoning)' },
                { value: 'o1-mini', name: 'o1-mini (Fast Reasoning)' }
            ],
            showApiToken: false,
            newApiToken: '',
        }
    },
    methods: {
        async openModal() {
            await this.loadConfig();
            $('#modalAIChatbot').modal({
                onApprove: () => false
            }).modal('show');
        },
        async loadConfig() {
            this.loading = true;
            try {
                const response = await window.http.get('/app/ai-config');
                this.config = response.data.results;
                this.newApiToken = ''; // Reset the new token field
            } catch (error) {
                if (error.response) {
                    showErrorInfo(error.response.data.message);
                } else {
                    showErrorInfo(error.message);
                }
            } finally {
                this.loading = false;
            }
        },
        async saveConfig() {
            this.saving = true;
            try {
                const payload = {
                    enabled: this.config.enabled,
                    model: this.config.model,
                    system_prompt: this.config.system_prompt
                };
                
                // Only send API token if a new one was entered
                if (this.newApiToken && this.newApiToken.trim()) {
                    payload.api_token = this.newApiToken.trim();
                }
                
                const response = await window.http.put('/app/ai-config', payload);
                this.config = response.data.results;
                this.newApiToken = ''; // Reset after save
                showSuccessInfo('✅ AI Chatbot настройки сохранены!');
            } catch (error) {
                if (error.response) {
                    showErrorInfo(error.response.data.message);
                } else {
                    showErrorInfo(error.message);
                }
            } finally {
                this.saving = false;
            }
        },
        toggleApiTokenVisibility() {
            this.showApiToken = !this.showApiToken;
        }
    },
    computed: {
        statusText() {
            return this.config.enabled ? 'Включен' : 'Выключен';
        },
        statusColor() {
            return this.config.enabled ? 'green' : 'grey';
        },
        hasApiToken() {
            return this.config.api_token && this.config.api_token !== '';
        }
    },
    template: `
    <div class="purple card" @click="openModal" style="cursor: pointer">
        <div class="content">
            <a class="ui purple right ribbon label">AI</a>
            <div class="header">🤖 AI Chatbot</div>
            <div class="description">
                Настройте автоматические ответы AI на входящие сообщения.
            </div>
            <div class="meta" style="margin-top: 8px;">
                <span :class="['ui', statusColor, 'label']">{{ statusText }}</span>
            </div>
        </div>
    </div>
    
    <!-- Modal AI Chatbot Config -->
    <div class="ui large modal" id="modalAIChatbot">
        <i class="close icon"></i>
        <div class="header">
            <i class="robot icon"></i> Настройки AI Chatbot
        </div>
        <div class="content">
            <div v-if="loading" class="ui active centered inline loader"></div>
            
            <div v-else class="ui form">
                <!-- Enable/Disable Toggle -->
                <div class="field">
                    <div class="ui toggle checkbox">
                        <input type="checkbox" v-model="config.enabled" id="aiEnabled">
                        <label for="aiEnabled">
                            <strong>Включить AI Chatbot</strong>
                            <p style="color: #888; font-weight: normal;">
                                Когда включено, AI будет автоматически отвечать на входящие личные сообщения
                            </p>
                        </label>
                    </div>
                </div>
                
                <div class="ui divider"></div>
                
                <!-- API Token -->
                <div class="field">
                    <label>
                        <i class="key icon"></i> OpenAI API Token
                    </label>
                    <div v-if="hasApiToken" class="ui message info">
                        <p><i class="check circle icon"></i> API токен установлен: <code>{{ config.api_token }}</code></p>
                    </div>
                    <div class="ui action input">
                        <input 
                            :type="showApiToken ? 'text' : 'password'" 
                            v-model="newApiToken"
                            placeholder="Введите новый API токен (sk-...)"
                        >
                        <button class="ui icon button" @click="toggleApiTokenVisibility" type="button">
                            <i :class="showApiToken ? 'eye slash icon' : 'eye icon'"></i>
                        </button>
                    </div>
                    <small style="color: #888;">
                        Получите API ключ на <a href="https://platform.openai.com/api-keys" target="_blank">platform.openai.com/api-keys</a>
                    </small>
                </div>
                
                <!-- Model Selection -->
                <div class="field">
                    <label>
                        <i class="microchip icon"></i> Модель GPT
                    </label>
                    <select class="ui dropdown" v-model="config.model">
                        <option v-for="model in availableModels" :key="model.value" :value="model.value">
                            {{ model.name }}
                        </option>
                    </select>
                </div>
                
                <!-- System Prompt -->
                <div class="field">
                    <label>
                        <i class="comment alternate icon"></i> Системный промпт (инструкции для AI)
                    </label>
                    <textarea 
                        v-model="config.system_prompt" 
                        rows="6"
                        placeholder="Опишите роль AI, как он должен отвечать, его приоритеты..."
                    ></textarea>
                    <small style="color: #888;">
                        Здесь вы описываете кто такой AI, как он должен отвечать, его роль, приоритеты и стиль общения.
                    </small>
                </div>
                
                <!-- Info Box -->
                <div class="ui info message">
                    <div class="header">
                        <i class="info circle icon"></i> Возможности AI Chatbot
                    </div>
                    <ul class="list">
                        <li><strong>Текстовые сообщения:</strong> AI отвечает на все текстовые сообщения</li>
                        <li><strong>Голосовые сообщения:</strong> AI автоматически расшифровывает аудио (Whisper) и отвечает на них</li>
                        <li><strong>Личные чаты:</strong> Работает только в личных чатах (не в группах)</li>
                    </ul>
                </div>
                
                <!-- Warning if not configured -->
                <div v-if="config.enabled && !hasApiToken && !newApiToken" class="ui warning message">
                    <div class="header">
                        <i class="exclamation triangle icon"></i> API токен не установлен
                    </div>
                    <p>Для работы AI Chatbot необходимо указать API токен OpenAI.</p>
                </div>
            </div>
        </div>
        <div class="actions">
            <button class="ui cancel button">
                Отмена
            </button>
            <button 
                class="ui positive right labeled icon button" 
                @click="saveConfig"
                :class="{ loading: saving }"
                :disabled="saving"
            >
                Сохранить
                <i class="save icon"></i>
            </button>
        </div>
    </div>
    `
}


