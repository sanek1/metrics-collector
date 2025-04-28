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

	flag.StringVar(&privateKeyPath, "private", "private.pem", "path to save private key")
	flag.StringVar(&publicKeyPath, "public", "public.pem", "path to save public key")
	flag.Parse()

	fmt.Printf("generation RSA keys:\n")
	fmt.Printf("private key: %s\n", privateKeyPath)
	fmt.Printf("public key: %s\n", publicKeyPath)

	if _, err := os.Stat(privateKeyPath); err == nil {
		fmt.Printf("file %s already exists, will be overwritten\n", privateKeyPath)
	}
	if _, err := os.Stat(publicKeyPath); err == nil {
		fmt.Printf("file %s already exists, will be overwritten\n", publicKeyPath)
	}

	if err := crypto.GenerateKeyPair(privateKeyPath, publicKeyPath); err != nil {
		fmt.Printf("error generating keys: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("keys generated successfully!\n")
	fmt.Printf("\nusage on server:\n")
	fmt.Printf("  run server with flag: -crypto-key=%s\n", privateKeyPath)
	fmt.Printf("  or set environment variable: CRYPTO_KEY=%s\n", privateKeyPath)

	fmt.Printf("\nusage on agent:\n")
	fmt.Printf("  run agent with flag: -crypto-key=%s\n", publicKeyPath)
	fmt.Printf("  or set environment variable: CRYPTO_KEY=%s\n", publicKeyPath)
}
