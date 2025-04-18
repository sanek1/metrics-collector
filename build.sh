#!/bin/bash

# Скрипт для сборки приложений с информацией о версии, дате и коммите

# Получаем текущую версию из git тегов (или устанавливаем дефолтную)
VERSION=$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.1.0")

# Получаем хеш последнего коммита
COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Устанавливаем дату сборки в формате ISO 8601
DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

echo "Building with the following information:"
echo "Version: $VERSION"
echo "Date: $DATE"
echo "Commit: $COMMIT"

# Сборка сервера
echo "Building server..."
go build -o bin/server -ldflags "-X main.buildVersion=$VERSION -X main.buildDate=$DATE -X main.buildCommit=$COMMIT" ./cmd/server

# Сборка агента
echo "Building agent..."
go build -o bin/agent -ldflags "-X main.buildVersion=$VERSION -X main.buildDate=$DATE -X main.buildCommit=$COMMIT" ./cmd/agent

echo "Build completed successfully!" 