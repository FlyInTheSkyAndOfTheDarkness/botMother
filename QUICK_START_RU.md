# Быстрый запуск AI Chatbot для WhatsApp

## 🚀 Варианты запуска

### Вариант 1: Через Docker (Рекомендуется, если Go не установлен)

#### Требования:
- Docker и Docker Compose

#### Шаги:

1. **Запуск через Docker Compose:**
   ```bash
   docker-compose up -d --build
   ```

2. **Или запуск через Docker с переменными окружения:**
   ```bash
   docker run -d -p 3000:3000 --name whatsapp \
     -e AI_CHATBOT_ENABLED=true \
     -e AI_CHATBOT_API_TOKEN=sk-your-token-here \
     -e AI_CHATBOT_MODEL=gpt-4o-mini \
     -e AI_CHATBOT_SYSTEM_PROMPT="You are a helpful assistant" \
     aldinokemal2104/go-whatsapp-web-multidevice rest
   ```

3. **Откройте браузер:** `http://localhost:3000`

4. **Настройте AI через API:**
   ```bash
   curl -X PUT http://localhost:3000/app/ai-config \
     -H "Content-Type: application/json" \
     -d '{
       "enabled": true,
       "api_token": "sk-your-token-here",
       "model": "gpt-4o-mini",
       "system_prompt": "You are a helpful assistant"
     }'
   ```

---

### Вариант 2: Локальный запуск (Требуется Go)

#### Требования:
- Go 1.24.0 или выше
- Скачать: https://golang.org/dl/

#### Шаги:

1. **Установите зависимости:**
   ```bash
   cd src
   go mod download
   ```

2. **Запустите приложение:**
   
   **С переменными окружения:**
   ```bash
   $env:AI_CHATBOT_ENABLED="true"
   $env:AI_CHATBOT_API_TOKEN="sk-your-token-here"
   $env:AI_CHATBOT_MODEL="gpt-4o-mini"
   $env:AI_CHATBOT_SYSTEM_PROMPT="You are a helpful assistant"
   go run . rest
   ```
   
   **Или через аргументы командной строки:**
   ```bash
   go run . rest --ai-chatbot-enabled=true --ai-chatbot-api-token=sk-your-token --ai-chatbot-model=gpt-4o-mini --ai-chatbot-system-prompt="You are a helpful assistant"
   ```

3. **Откройте браузер:** `http://localhost:3000`

---

## ⚙️ Настройка AI Chatbot

### Через REST API (после запуска)

1. **Получить текущую конфигурацию:**
   ```bash
   GET http://localhost:3000/app/ai-config
   ```

2. **Обновить конфигурацию:**
   ```bash
   PUT http://localhost:3000/app/ai-config
   Content-Type: application/json
   
   {
     "enabled": true,
     "api_token": "sk-your-openai-api-token",
     "model": "gpt-4o-mini",
     "system_prompt": "You are a helpful assistant. Respond concisely."
   }
   ```

### Через файл .env (только для локального запуска)

Создайте файл `src/.env`:
```env
AI_CHATBOT_ENABLED=true
AI_CHATBOT_API_TOKEN=sk-your-openai-api-token
AI_CHATBOT_MODEL=gpt-4o-mini
AI_CHATBOT_SYSTEM_PROMPT=You are a helpful assistant. Respond concisely and helpfully.
```

---

## 📱 Подключение WhatsApp

1. Откройте `http://localhost:3000` в браузере
2. Нажмите "Login" или используйте API: `GET /app/login`
3. Отсканируйте QR-код в WhatsApp (Настройки → Связанные устройства)
4. После подключения AI Chatbot начнет автоматически отвечать

---

## 🧪 Тестирование

1. **Текстовые сообщения:**
   - Отправьте текстовое сообщение на подключенный номер
   - AI автоматически ответит

2. **Голосовые сообщения (аудио):**
   - Отправьте голосовое сообщение
   - Оно будет автоматически транскрибировано через Whisper API
   - AI ответит на транскрибированный текст

---

## 📝 Доступные модели OpenAI

- `gpt-4o-mini` (по умолчанию, быстрый и недорогой)
- `gpt-4o`
- `gpt-4-turbo`
- `gpt-3.5-turbo`
- Другие модели OpenAI

---

## ⚠️ Важные замечания

- **AI Chatbot работает только с прямыми сообщениями (1:1)**, не отвечает в группах
- **Для аудио сообщений** требуется `WhatsappAutoDownloadMedia=true` (включено по умолчанию)
- **AI Chatbot имеет приоритет** над автоответом (если AI включен, автоответ отключен)
- **API токен маскируется** в ответах API для безопасности (показываются только последние 4 символа)

---

## 🔧 Отладка

Если что-то не работает:

1. Проверьте логи приложения
2. Убедитесь, что OpenAI API токен действителен
3. Проверьте, что WhatsApp подключен (статус в UI)
4. Проверьте, что AI Chatbot включен: `GET /app/ai-config`

