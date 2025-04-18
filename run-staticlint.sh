#!/bin/bash

echo "====== Запуск статического анализатора ======"

# Собираем анализатор
echo "* Сборка анализатора..."
go build -buildvcs=false -o staticlint ./cmd/staticlint

# Создаем список файлов для анализа
echo "* Создание списка файлов для анализа..."
TEMPFILE=$(mktemp /tmp/gofiles-XXXXXX.txt)
find . -path "*/\.git/*" -prune -o -path "*/go-build/*" -prune -o -path "*/vendor/*" -prune -o -name "*.go" -print > $TEMPFILE

# Проверка скриптов сборки
echo "* Проверка скриптов сборки..."
if command -v shellcheck &> /dev/null; then
    echo "  - Проверка build.sh с помощью shellcheck..."
    shellcheck build.sh || echo "    Предупреждение: найдены проблемы в build.sh"
else
    echo "  - shellcheck не установлен, пропускаем проверку скриптов..."
fi

# Запускаем анализатор
echo "* Запуск полного анализа кода..."
./staticlint -- @$TEMPFILE

echo "* Запуск проверки exitchecker..."
./staticlint -exitchecker -- @$TEMPFILE

# Удаляем временный файл
rm $TEMPFILE

echo "====== Анализ завершен ======" 