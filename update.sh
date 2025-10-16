#!/bin/bash

echo "🔄 Updating Dolphin application..."

# Stop current process
pkill dolphin 2>/dev/null || true

# Update from git
echo "📥 Pulling latest changes..."
git pull origin main

# Update dependencies
echo "📦 Updating dependencies..."
go mod tidy

# Rebuild application
echo "🔨 Building application..."
go build -o dolphin cmd/dolphin/main.go

# Run migrations
echo "🗄️ Running migrations..."
./dolphin migrate:run

# Clear caches
echo "🧹 Clearing caches..."
./dolphin cache:clear
./dolphin static:clear-cache

echo "✅ Update complete! Starting application..."
./dolphin serve
