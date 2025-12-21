package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"hash"
	"io"
	"os"
)

type CryptoKey interface {
	*x509.Certificate | *rsa.PrivateKey | *rsa.PublicKey
}

type CryptoService[T CryptoKey, V CryptoKey] struct {
	key       T
	extract   func(T) V
	transform func(hash hash.Hash, random io.Reader, v V, ciphertext []byte, label []byte) ([]byte, error)
}

func (cs *CryptoService[T, V]) Transform(message []byte) ([]byte, error) {

	value := cs.extract(cs.key)
	return cs.transform(sha256.New(), rand.Reader, value, message, nil)
}

func NewPublicKeyService(filePath string) (*CryptoService[*x509.Certificate, *rsa.PublicKey], error) {
	cert, err := loadKey(filePath, x509.ParseCertificate)
	if err != nil {
		return nil, err
	}
	service := CryptoService[*x509.Certificate, *rsa.PublicKey]{
		key: cert,
		extract: func(cert *x509.Certificate) *rsa.PublicKey {
			return cert.PublicKey.(*rsa.PublicKey)
		},
		transform: rsa.EncryptOAEP,
	}
	return &service, nil
}

func NewPrivateKeyService(filePath string) (*CryptoService[*rsa.PrivateKey, *rsa.PrivateKey], error) {
	cert, err := loadKey(filePath, x509.ParsePKCS1PrivateKey)
	if err != nil {
		return nil, err
	}
	service := CryptoService[*rsa.PrivateKey, *rsa.PrivateKey]{
		key: cert,
		extract: func(cert *rsa.PrivateKey) *rsa.PrivateKey {
			return cert
		},
		transform: rsa.DecryptOAEP,
	}
	return &service, nil
}

func loadKey[T CryptoKey](filePath string, parse func(der []byte) (T, error)) (T, error) {
	certBytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	pemBlock, _ := pem.Decode(certBytes)
	if pemBlock == nil {
		return nil, fmt.Errorf("certificate not found")
	}

	certificate, err := parse(pemBlock.Bytes)
	if err != nil {
		return nil, err
	}
	return certificate, nil
}
