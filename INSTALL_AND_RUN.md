# Инструкция по установке и запуску

## Проблема решена! ✅

Зависимость `github.com/sashabaranov/go-openai` добавлена в `go.mod`.

## Запуск приложения

### 1. Установите зависимости (если еще не сделали):

```powershell
cd src
go mod download
```

Или просто:
```powershell
cd src
go mod tidy
```

### 2. Запустите приложение:

```powershell
go run . rest
```

### 3. Откройте в браузере:

```
http://localhost:3000
```

## Настройка AI Chatbot

После запуска приложения:

### Вариант 1: Через REST API

Используйте PowerShell для настройки:

```powershell
$body = @{
    enabled = $true
    api_token = "sk-your-openai-api-token-here"
    model = "gpt-4o-mini"
    system_prompt = "You are a helpful assistant. Respond concisely and helpfully."
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:3000/app/ai-config" -Method PUT -ContentType "application/json" -Body $body
```

### Вариант 2: Через переменные окружения

Создайте файл `src/.env` (если его еще нет):

```env
AI_CHATBOT_ENABLED=true
AI_CHATBOT_API_TOKEN=sk-your-openai-api-token-here
AI_CHATBOT_MODEL=gpt-4o-mini
AI_CHATBOT_SYSTEM_PROMPT=You are a helpful assistant. Respond concisely and helpfully.
```

### Вариант 3: Через аргументы командной строки

```powershell
go run . rest --ai-chatbot-enabled=true --ai-chatbot-api-token=sk-your-token --ai-chatbot-model=gpt-4o-mini --ai-chatbot-system-prompt="You are a helpful assistant"
```

## Подключение WhatsApp

1. Откройте `http://localhost:3000`
2. Нажмите "Login"
3. Отсканируйте QR-код в WhatsApp
4. Готово! AI Chatbot начнет отвечать на сообщения

## Проверка работы

1. Отправьте текстовое сообщение на подключенный номер
2. AI автоматически ответит
3. Отправьте голосовое сообщение - оно будет транскрибировано и AI ответит

## Проверка конфигурации AI

Получить текущую конфигурацию:

```powershell
Invoke-RestMethod -Uri "http://localhost:3000/app/ai-config" -Method GET
```

