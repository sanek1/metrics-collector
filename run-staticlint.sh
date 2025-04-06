#!/bin/bash

echo "====== Запуск статического анализатора ======"

# Собираем анализатор
echo "* Сборка анализатора..."
go build -o staticlint ./cmd/staticlint

# Создаем список файлов для анализа
echo "* Создание списка файлов для анализа..."
TEMPFILE=$(mktemp /tmp/gofiles-XXXXXX.txt)
find . -path "*/\.git/*" -prune -o -path "*/go-build/*" -prune -o -path "*/vendor/*" -prune -o -name "*.go" -print > $TEMPFILE

# Запускаем анализатор
echo "* Запуск полного анализа кода..."
./staticlint -- @$TEMPFILE

echo "* Запуск проверки exitchecker..."
./staticlint -exitchecker -- @$TEMPFILE

# Удаляем временный файл
rm $TEMPFILE

echo "====== Анализ завершен ======" 