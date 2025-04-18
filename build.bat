@echo off
REM Скрипт для сборки приложений с информацией о версии, дате и коммите

REM Создаем директорию bin, если она не существует
if not exist bin mkdir bin

REM Получаем текущую версию из git тегов (или устанавливаем дефолтную)
for /f "tokens=*" %%i in ('git describe --tags --abbrev^=0 2^>nul') do set VERSION=%%i
if "%VERSION%"=="" set VERSION=v0.1.0

REM Получаем хеш последнего коммита
for /f "tokens=*" %%i in ('git rev-parse --short HEAD 2^>nul') do set COMMIT=%%i
if "%COMMIT%"=="" set COMMIT=unknown

REM Устанавливаем дату сборки в формате ISO 8601
for /f "tokens=2 delims==" %%I in ('wmic os get localdatetime /format:list') do set datetime=%%I
set DATE=%datetime:~0,4%-%datetime:~4,2%-%datetime:~6,2%T%datetime:~8,2%:%datetime:~10,2%:%datetime:~12,2%Z

echo Building with the following information:
echo Version: %VERSION%
echo Date: %DATE%
echo Commit: %COMMIT%

REM Сборка сервера
echo Building server...
go build -o bin\server.exe -ldflags "-X main.buildVersion=%VERSION% -X main.buildDate=%DATE% -X main.buildCommit=%COMMIT%" .\cmd\server

REM Сборка агента
echo Building agent...
go build -o bin\agent.exe -ldflags "-X main.buildVersion=%VERSION% -X main.buildDate=%DATE% -X main.buildCommit=%COMMIT%" .\cmd\agent

echo Build completed successfully! 