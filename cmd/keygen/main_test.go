package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sanek1/metrics-collector/internal/crypto"
	"github.com/stretchr/testify/require"
)

func TestGenerateKeyPair(t *testing.T) {
	tempDir := t.TempDir()
	privateKeyPath := filepath.Join(tempDir, "test_private.pem")
	publicKeyPath := filepath.Join(tempDir, "test_public.pem")

	err := crypto.GenerateKeyPair(privateKeyPath, publicKeyPath)
	require.NoError(t, err, "error generating key pair")

	_, err = os.Stat(privateKeyPath)
	require.NoError(t, err, "private key should be created")

	_, err = os.Stat(publicKeyPath)
	require.NoError(t, err, "public key should be created")

	err = crypto.GenerateKeyPair(privateKeyPath, publicKeyPath)
	require.NoError(t, err, "error rewriting key pair")
}
