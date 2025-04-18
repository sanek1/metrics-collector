// Package crypto предоставляет функции для асимметричного шифрования
package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

// LoadPublicKey загружает публичный ключ RSA из файла
func LoadPublicKey(filePath string) (*rsa.PublicKey, error) {
	pemData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error loading public key: %w", err)
	}

	block, _ := pem.Decode(pemData)
	if block == nil || block.Type != "PUBLIC KEY" {
		return nil, fmt.Errorf("invalid PEM block with public key")
	}

	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("error parsing public key: %w", err)
	}

	pub, ok := pubInterface.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("error converting key to RSA public key")
	}

	return pub, nil
}

// LoadPrivateKey загружает приватный ключ RSA из файла
func LoadPrivateKey(filePath string) (*rsa.PrivateKey, error) {
	pemData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading private key: %w", err)
	}

	block, _ := pem.Decode(pemData)
	if block == nil || block.Type != "PRIVATE KEY" {
		return nil, fmt.Errorf("invalid PEM block with private key")
	}

	priv, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("error parsing private key: %w", err)
	}

	privateKey, ok := priv.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("error converting key to RSA private key")
	}

	return privateKey, nil
}

// EncryptData шифрует данные с помощью публичного ключа RSA
func EncryptData(publicKey *rsa.PublicKey, data []byte) ([]byte, error) {
	// Используем OAEP шифрование с SHA-256
	encrypted, err := rsa.EncryptOAEP(
		sha256.New(),
		rand.Reader,
		publicKey,
		data,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("error encrypting data: %w", err)
	}
	return encrypted, nil
}

// DecryptData расшифровывает данные с помощью приватного ключа RSA
func DecryptData(privateKey *rsa.PrivateKey, data []byte) ([]byte, error) {
	// Используем OAEP расшифровку с SHA-256
	decrypted, err := rsa.DecryptOAEP(
		sha256.New(),
		rand.Reader,
		privateKey,
		data,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("error decrypting data: %w", err)
	}
	return decrypted, nil
}

// GenerateKeyPair создает новую пару ключей RSA и сохраняет их в файлы
func GenerateKeyPair(privateKeyPath, publicKeyPath string) error {
	// Генерируем приватный ключ
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("error generating keys: %w", err)
	}

	// Создаем PEM блок для приватного ключа
	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return fmt.Errorf("error marshalling private key: %w", err)
	}

	privatePEM := pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privateKeyBytes,
	}

	privateFile, err := os.Create(privateKeyPath)
	if err != nil {
		return fmt.Errorf("error creating private key file: %w", err)
	}
	defer privateFile.Close()

	if err := pem.Encode(privateFile, &privatePEM); err != nil {
		return fmt.Errorf("error saving private key: %w", err)
	}

	// Создаем PEM блок для публичного ключа
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return fmt.Errorf("error marshalling public key: %w", err)
	}

	publicPEM := pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	}

	publicFile, err := os.Create(publicKeyPath)
	if err != nil {
		return fmt.Errorf("error creating public key file: %w", err)
	}
	defer publicFile.Close()

	if err := pem.Encode(publicFile, &publicPEM); err != nil {
		return fmt.Errorf("error saving public key: %w", err)
	}

	return nil
}
