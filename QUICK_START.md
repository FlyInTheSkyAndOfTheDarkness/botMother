# Быстрый запуск AI Chatbot для WhatsApp

## Шаги для запуска:

### 1. Установка зависимостей (если нужно)

Если Go не установлен, установите его с https://golang.org/dl/

### 2. Установка зависимостей проекта

```bash
cd src
go mod download
```

### 3. Запуск приложения

```bash
go run . rest
```

Приложение запустится на `http://localhost:3000`

### 4. Настройка AI Chatbot

#### Вариант 1: Через REST API (рекомендуется)

Откройте браузер и перейдите на `http://localhost:3000`

Затем используйте API для настройки:

**GET /app/ai-config** - Получить текущую конфигурацию
```
GET http://localhost:3000/app/ai-config
```

**PUT /app/ai-config** - Настроить AI Chatbot
```json
PUT http://localhost:3000/app/ai-config
Content-Type: application/json

{
  "enabled": true,
  "api_token": "sk-your-openai-api-token-here",
  "model": "gpt-4o-mini",
  "system_prompt": "You are a helpful assistant. Respond concisely and helpfully to user messages."
}
```

#### Вариант 2: Через переменные окружения

Создайте файл `.env` в папке `src`:

```env
AI_CHATBOT_ENABLED=true
AI_CHATBOT_API_TOKEN=sk-your-openai-api-token-here
AI_CHATBOT_MODEL=gpt-4o-mini
AI_CHATBOT_SYSTEM_PROMPT=You are a helpful assistant. Respond concisely and helpfully to user messages.
```

#### Вариант 3: Через аргументы командной строки

```bash
go run . rest --ai-chatbot-enabled=true --ai-chatbot-api-token=sk-your-token --ai-chatbot-model=gpt-4o-mini --ai-chatbot-system-prompt="You are a helpful assistant"
```

### 5. Подключение WhatsApp

1. Откройте `http://localhost:3000` в браузере
2. Нажмите на "Login" или используйте API: `GET /app/login`
3. Отсканируйте QR-код с помощью WhatsApp на телефоне
4. После подключения AI Chatbot начнет отвечать на входящие сообщения

### 6. Тестирование

1. Отправьте текстовое сообщение на подключенный номер WhatsApp
2. AI Chatbot автоматически ответит (если включен)
3. Отправьте голосовое сообщение (аудио) - оно будет транскрибировано и AI ответит

### Доступные модели OpenAI:

- `gpt-4o-mini` (по умолчанию, быстрый и недорогой)
- `gpt-4o`
- `gpt-4-turbo`
- `gpt-3.5-turbo`
- И другие модели OpenAI

### Примечания:

- Для работы с аудио сообщениями требуется `WhatsappAutoDownloadMedia=true` (включено по умолчанию)
- AI Chatbot работает только с прямыми сообщениями (1:1), не отвечает в группах
- AI Chatbot имеет приоритет над автоответом (если AI включен, автоответ не работает)

