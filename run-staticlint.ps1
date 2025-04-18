# Скрипт для запуска статического анализатора, исключая кэш сборки
# Убедитесь, что у вас собран staticlint
# go build -o staticlint.exe ./cmd/staticlint

# Получаем список всех .go файлов, исключая кэш сборки и .git
$goFiles = Get-ChildItem -Path . -Recurse -Include "*.go" | Where-Object { 
    $_.FullName -notmatch '[\\/]\.git[\\/]' -and 
    $_.FullName -notmatch '[\\/]go-build[\\/]' -and
    $_.FullName -notmatch 'AppData[\\/]Local[\\/]go-build[\\/]'
}

# Создаем временный файл со списком файлов
$tempFile = [System.IO.Path]::GetTempFileName()
$goFiles.FullName | Out-File -FilePath $tempFile -Encoding utf8

# Запускаем статический анализатор с полученным списком файлов
Write-Host "Запуск staticlint с анализом исходного кода..."
.\staticlint.exe @($tempFile)

# Запускаем отдельно exitchecker
Write-Host "Запуск exitchecker..."
.\staticlint.exe exitchecker @($tempFile)

# Удаляем временный файл
Remove-Item -Path $tempFile

Write-Host "Готово!" 