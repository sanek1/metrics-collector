@echo off
echo ====== Запуск статического анализатора =====

REM Собираем анализатор
echo * Сборка анализатора...
go build -o staticlint.exe ./cmd/staticlint

REM Создаем список файлов для анализа
echo * Создание списка файлов для анализа...
set TEMPFILE=%TEMP%\gofiles-%RANDOM%.txt
dir /s /b *.go | findstr /v "go-build" | findstr /v ".git" | findstr /v "AppData\Local\go-build" > %TEMPFILE%

REM Запускаем анализатор
echo * Запуск полного анализа кода...
staticlint.exe -- @%TEMPFILE%

echo * Запуск проверки exitchecker...
staticlint.exe -exitchecker -- @%TEMPFILE%

REM Удаляем временный файл
del %TEMPFILE%

echo ====== Анализ завершен ======
pause 