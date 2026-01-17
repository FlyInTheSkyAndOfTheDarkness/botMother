#!/bin/bash
# Deploy WhatsApp fix to server
# Run this script from the project root directory

echo "📦 Copying files to server..."

# Copy the fixed files
scp src/infrastructure/whatsapp/agent_handler.go root@194.32.142.228:/root/botMother/src/infrastructure/whatsapp/
scp src/infrastructure/whatsapp/device_manager.go root@194.32.142.228:/root/botMother/src/infrastructure/whatsapp/

echo "🔨 Rebuilding on server..."
ssh root@194.32.142.228 "cd /root/botMother && docker compose up --build -d"

echo "✅ Deployment complete!"
echo "📋 Check logs: ssh root@194.32.142.228 'cd /root/botMother && docker compose logs -f app'"
