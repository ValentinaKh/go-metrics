package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"github.com/stretchr/testify/assert"
	"math/big"
	"os"
	"path/filepath"
	"testing"
)

func createTestKeys(t *testing.T) (publicPath, privatePath string) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate RSA key: %v", err)
	}

	publicKey := &privateKey.PublicKey

	privBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}
	privBytes := pem.EncodeToMemory(privBlock)

	template := &x509.Certificate{
		SerialNumber: randomSerial(),
	}
	certBytes, err := x509.CreateCertificate(rand.Reader, template, template, publicKey, privateKey)
	if err != nil {
		t.Fatalf("Failed to create self-signed cert: %v", err)
	}
	pubBlock := &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	}
	pubBytes := pem.EncodeToMemory(pubBlock)

	tempDir := t.TempDir()
	publicPath = filepath.Join(tempDir, "public.pem")
	privatePath = filepath.Join(tempDir, "private.pem")

	if err := os.WriteFile(publicPath, pubBytes, 0600); err != nil {
		t.Fatalf("Failed to write public key: %v", err)
	}
	if err := os.WriteFile(privatePath, privBytes, 0600); err != nil {
		t.Fatalf("Failed to write private key: %v", err)
	}

	return publicPath, privatePath
}

func randomSerial() *big.Int {
	serial, _ := rand.Int(rand.Reader, new(big.Int).Exp(big.NewInt(2), big.NewInt(128), nil))
	return serial
}

func TestInitCertificate_Valid(t *testing.T) {
	pubPath, _ := createTestKeys(t)

	err := InitCertificate(pubPath)
	if err != nil {
		t.Fatalf("InitCertificate failed: %v", err)
	}
	if public == nil {
		t.Fatal("public key should not be nil")
	}
}

func TestInitCertificate_InvalidPath(t *testing.T) {
	err := InitCertificate("/test/path.pem")
	if err == nil {
		t.Fatal("Expected error for non-existent file")
	}
}

func TestInitPrivateKey_Valid(t *testing.T) {
	_, privatePath := createTestKeys(t)

	err := InitPrivateKey(privatePath)
	if err != nil {
		t.Fatalf("InitPrivateKey failed: %v", err)
	}
	if private == nil {
		t.Fatal("private key should not be nil")
	}
}

func TestInitPrivateKey_InvalidPath(t *testing.T) {
	err := InitPrivateKey("/test/path.pem")
	if err == nil {
		t.Fatal("Expected error for non-existent file")
	}
}

func TestEncryptDecrypt(t *testing.T) {
	pubPath, privatePath := createTestKeys(t)

	if err := InitCertificate(pubPath); err != nil {
		t.Fatalf("InitCertificate: %v", err)
	}
	if err := InitPrivateKey(privatePath); err != nil {
		t.Fatalf("InitPrivateKey: %v", err)
	}

	original := []byte("secret message")

	encrypted, err := Encrypt(original)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	decrypted, err := Decrypt(encrypted)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}

	if string(decrypted) != string(original) {
		t.Errorf("Decrypted != Original: got %q, want %q", decrypted, original)
	}
}

func TestEncrypt_WithNoPublicKey(t *testing.T) {
	public = nil

	msg := []byte("test")
	encrypted, err := Encrypt(msg)
	if err != nil {
		t.Fatalf("Encrypt with nil public key should not fail (per your code), got error: %v", err)
	}
	assert.Equal(t, string(msg), string(encrypted), "Expected original message when public is nil")
}

func TestDecrypt_WithNoPrivateKey(t *testing.T) {
	private = nil

	msg := []byte("test")
	decrypted, err := Decrypt(msg)
	if err != nil {
		t.Fatalf("Decrypt with nil private key should not fail, got error: %v", err)
	}
	assert.Equal(t, string(msg), string(decrypted), "Expected original message when private is nil")
}
