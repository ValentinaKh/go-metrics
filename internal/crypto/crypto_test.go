package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestLoadCertificate_Valid(t *testing.T) {
	pubPath, privatePath := createTestKeys(t)

	cert, err := loadKey[*x509.Certificate](pubPath, x509.ParseCertificate)
	assert.NoError(t, err)
	assert.NotNil(t, cert)

	cert1, err1 := loadKey[*rsa.PrivateKey](privatePath, x509.ParsePKCS1PrivateKey)
	assert.NoError(t, err1)
	assert.NotNil(t, cert1)

}

func TestLoadCertificate_InvalidPath(t *testing.T) {
	cert, err := loadKey[*x509.Certificate]("/test/path.pem", x509.ParseCertificate)
	assert.Error(t, err)
	assert.Nil(t, cert)
}

func TestEncryptDecrypt(t *testing.T) {
	pubPath, privatePath := createTestKeys(t)

	public, err := NewPublicKeyService(pubPath)
	if err != nil {
		t.Fatalf("InitCertificate: %v", err)
	}
	private, err := NewPrivateKeyService(privatePath)
	if err != nil {
		t.Fatalf("InitPrivateKey: %v", err)
	}

	original := []byte("secret message")

	encrypted, err := public.Transform(original)
	require.NoError(t, err)

	decrypted, err := private.Transform(encrypted)
	require.NoError(t, err)
	require.Equal(t, original, decrypted)
}
