#!/bin/bash
echo "Deploying Audio & Delete fixes..."

# Copy modified files
scp src/infrastructure/agent/repository.go root@194.32.142.228:~/go-whatsapp/src/infrastructure/agent/
scp src/infrastructure/telegram/bot.go root@194.32.142.228:~/go-whatsapp/src/infrastructure/telegram/

# Rebuild and restart
ssh root@194.32.142.228 "cd ~/go-whatsapp && docker-compose build app && docker-compose up -d app"

echo "Deployment complete!"
