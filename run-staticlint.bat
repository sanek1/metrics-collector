@echo off
echo ====== Запуск статического анализатора =====

REM Собираем анализатор
echo * Сборка анализатора...
go build -buildvcs=false -o staticlint.exe ./cmd/staticlint

REM Создаем список файлов для анализа
echo * Создание списка файлов для анализа...
set TEMPFILE=%TEMP%\gofiles-%RANDOM%.txt
dir /s /b *.go | findstr /v "go-build" | findstr /v ".git" | findstr /v "AppData\Local\go-build" > %TEMPFILE%

REM Проверка скриптов сборки
echo * Проверка скриптов сборки...
where /q powershell
if %ERRORLEVEL% EQU 0 (
    echo   - Проверка build.bat с помощью PowerShell...
    powershell -Command "if ((Get-Content build.bat) -match 'goto :eof' -or (Get-Content build.bat) -match 'exit /b') { echo '    OK' } else { echo '    Предупреждение: build.bat может не иметь правильного завершения' }"
) else (
    echo   - PowerShell не найден, пропускаем проверку скриптов...
)

REM Запускаем анализатор
echo * Запуск полного анализа кода...
staticlint.exe -- @%TEMPFILE%

echo * Запуск проверки exitchecker...
staticlint.exe -exitchecker -- @%TEMPFILE%

REM Удаляем временный файл
del %TEMPFILE%

echo ====== Анализ завершен ======
pause 