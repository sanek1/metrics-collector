package crypto

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"os"
	"testing"
)

func TestGenerateAndLoadKeys(t *testing.T) {
	privateKeyPath := "test_private.pem"
	publicKeyPath := "test_public.pem"

	defer func() {
		os.Remove(privateKeyPath)
		os.Remove(publicKeyPath)
	}()

	err := GenerateKeyPair(privateKeyPath, publicKeyPath)
	if err != nil {
		t.Fatalf("error generating keys: %v", err)
	}

	if _, err := os.Stat(privateKeyPath); os.IsNotExist(err) {
		t.Fatalf("private key file not created: %v", err)
	}
	if _, err := os.Stat(publicKeyPath); os.IsNotExist(err) {
		t.Fatalf("public key file not created: %v", err)
	}

	privateKey, err := LoadPrivateKey(privateKeyPath)
	if err != nil {
		t.Fatalf("error loading private key: %v", err)
	}
	publicKey, err := LoadPublicKey(publicKeyPath)
	if err != nil {
		t.Fatalf("error loading public key: %v", err)
	}
	if privateKey == nil {
		t.Fatal("loaded private key is nil")
	}
	if publicKey == nil {
		t.Fatal("loaded public key is nil")
	}
}

func TestEncryptionAndDecryption(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("error generating test keys: %v", err)
	}
	publicKey := &privateKey.PublicKey
	testData := []byte("This is test data for encryption and decryption")

	encrypted, err := EncryptData(publicKey, testData)
	if err != nil {
		t.Fatalf("error encrypting data: %v", err)
	}

	if bytes.Equal(encrypted, testData) {
		t.Fatal("encrypted data is identical to original data, which is impossible with asymmetric encryption")
	}

	decrypted, err := DecryptData(privateKey, encrypted)
	if err != nil {
		t.Fatalf("error decrypting data: %v", err)
	}

	if !bytes.Equal(decrypted, testData) {
		t.Fatal("decrypted data does not match original data")
	}
}

func TestLoadKeys_FileNotExist(t *testing.T) {
	_, err := LoadPrivateKey("non_existent_private.pem")
	if err == nil {
		t.Fatal("expected error when loading non-existent private key")
	}

	_, err = LoadPublicKey("non_existent_public.pem")
	if err == nil {
		t.Fatal("expected error when loading non-existent public key")
	}
}

func TestEncryptDecrypt_SmallChunks(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("error generating test keys: %v", err)
	}
	publicKey := &privateKey.PublicKey

	testCases := [][]byte{
		[]byte("a"),
		[]byte("abc"),
		[]byte("12345"),
		[]byte{0, 1, 2, 3, 4, 5},
		[]byte(""),
	}

	for _, testData := range testCases {
		encrypted, err := EncryptData(publicKey, testData)
		if err != nil {
			t.Fatalf("error encrypting data: %v (data length: %d)", err, len(testData))
		}

		decrypted, err := DecryptData(privateKey, encrypted)
		if err != nil {
			t.Fatalf("error decrypting data: %v (data length: %d)", err, len(testData))
		}

		if !bytes.Equal(decrypted, testData) {
			t.Fatalf("decrypted data does not match original data (data length: %d)", len(testData))
		}
	}
}
