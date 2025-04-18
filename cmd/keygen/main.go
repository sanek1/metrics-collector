// Package main предоставляет утилиту для генерации ключей шифрования
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/sanek1/metrics-collector/internal/crypto"
)

func main() {
	var privateKeyPath string
	var publicKeyPath string

	flag.StringVar(&privateKeyPath, "private", "private.pem", "путь для сохранения приватного ключа")
	flag.StringVar(&publicKeyPath, "public", "public.pem", "путь для сохранения публичного ключа")
	flag.Parse()

	fmt.Printf("Генерация ключей RSA:\n")
	fmt.Printf("Приватный ключ будет сохранен в: %s\n", privateKeyPath)
	fmt.Printf("Публичный ключ будет сохранен в: %s\n", publicKeyPath)

	// Проверка наличия файлов
	if _, err := os.Stat(privateKeyPath); err == nil {
		fmt.Printf("Файл %s уже существует, будет перезаписан\n", privateKeyPath)
	}
	if _, err := os.Stat(publicKeyPath); err == nil {
		fmt.Printf("Файл %s уже существует, будет перезаписан\n", publicKeyPath)
	}

	// Генерация ключей
	if err := crypto.GenerateKeyPair(privateKeyPath, publicKeyPath); err != nil {
		fmt.Printf("Ошибка при генерации ключей: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Ключи успешно сгенерированы!\n")
	fmt.Printf("\nИспользование на сервере:\n")
	fmt.Printf("  Запустите сервер с флагом: -crypto-key=%s\n", privateKeyPath)
	fmt.Printf("  Или установите переменную окружения: CRYPTO_KEY=%s\n", privateKeyPath)

	fmt.Printf("\nИспользование на агенте:\n")
	fmt.Printf("  Запустите агент с флагом: -crypto-key=%s\n", publicKeyPath)
	fmt.Printf("  Или установите переменную окружения: CRYPTO_KEY=%s\n", publicKeyPath)
}
