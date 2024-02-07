package keystore

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"ctb-cli/types"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"golang.org/x/crypto/chacha20poly1305"
)

// DeserializePrivateKey encrypts and serializes the private key
func (*KeyStore) DeserializePrivateKey(cipheredEncoded string, rootKey *types.Key) (*rsa.PrivateKey, error) {
	ciphered, err := base64.StdEncoding.DecodeString(cipheredEncoded)
	if err != nil {
		return nil, fmt.Errorf("failed to decode ciphered data: %w", err)
	}

	aead, err := chacha20poly1305.New(rootKey[:])
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	nonce := make([]byte, chacha20poly1305.NonceSize)

	deciphered, err := aead.Open(nil, nonce, ciphered, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt data: %w", err)
	}

	block, _ := pem.Decode(deciphered)
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing the key")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return privateKey, nil
}

// SerializePrivateKey encrypts and serializes the private key
func (*KeyStore) SerializePrivateKey(privateKey *rsa.PrivateKey, rootKey *types.Key) (string, error) {
	bytes := x509.MarshalPKCS1PrivateKey(privateKey)
	pem := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: bytes,
		},
	)

	aead, err := chacha20poly1305.New(rootKey[:])
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	nonce := make([]byte, chacha20poly1305.NonceSize)

	ciphered := aead.Seal(nil, nonce, pem, nil)
	cipheredEncoded := base64.StdEncoding.EncodeToString(ciphered)

	return cipheredEncoded, nil
}

// SerializeDataKey encrypts and serializes the key pair
func (*KeyStore) SerializeDataKey(key []byte, publicKey *rsa.PublicKey) ([]byte, error) {
	encrypted, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, publicKey, key[:], nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	return encrypted, nil
}

// DeserializeDataKey decrypts and deserializes the key pair
func (*KeyStore) DeserializeDataKey(ciphered []byte, privateKey *rsa.PrivateKey) (*Key, error) {
	deciphered, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, privateKey, ciphered, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	key := Key{}
	copy(key[:], deciphered)

	return &key, nil
}
