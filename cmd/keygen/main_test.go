package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sanek1/metrics-collector/internal/crypto"
	"github.com/stretchr/testify/require"
)

func TestGenerateKeyPair(t *testing.T) {
	// Создаем временные пути для ключей
	tempDir := t.TempDir()
	privateKeyPath := filepath.Join(tempDir, "test_private.pem")
	publicKeyPath := filepath.Join(tempDir, "test_public.pem")

	// Непосредственно генерируем ключи, минуя основную функцию main
	err := crypto.GenerateKeyPair(privateKeyPath, publicKeyPath)
	require.NoError(t, err, "Ошибка при генерации ключей")

	// Проверяем, что файлы ключей были созданы
	_, err = os.Stat(privateKeyPath)
	require.NoError(t, err, "Приватный ключ должен быть создан")

	_, err = os.Stat(publicKeyPath)
	require.NoError(t, err, "Публичный ключ должен быть создан")

	// Проверяем перезапись существующих файлов
	err = crypto.GenerateKeyPair(privateKeyPath, publicKeyPath)
	require.NoError(t, err, "Ошибка при перезаписи ключей")
}
