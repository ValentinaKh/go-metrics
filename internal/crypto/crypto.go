package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

var public *x509.Certificate
var private *rsa.PrivateKey

func readKey[T any](filePath string, parse func(der []byte) (*T, error)) (*T, error) {
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

func InitCertificate(publicKeyPath string) error {
	res, err := readKey(publicKeyPath, x509.ParseCertificate)
	if err != nil {
		return err
	}
	public = res
	return nil
}

func InitPrivateKey(privateKeyPath string) error {
	res, err := readKey(privateKeyPath, x509.ParsePKCS1PrivateKey)
	if err != nil {
		return err
	}
	private = res
	return nil
}

func Encrypt(message []byte) ([]byte, error) {
	if public == nil {
		return message, nil
	}
	return rsa.EncryptPKCS1v15(rand.Reader, public.PublicKey.(*rsa.PublicKey), message)
}

func Decrypt(message []byte) ([]byte, error) {
	if private == nil {
		return message, nil
	}
	return rsa.DecryptPKCS1v15(rand.Reader, private, message)
}
