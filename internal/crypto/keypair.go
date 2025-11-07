package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
)

const (
	// RSAKeySize is the size of RSA keys in bits
	RSAKeySize = 4096
)

// Type alias for cleaner code
type RSAPrivateKey = rsa.PrivateKey

// GenerateRSAKeypair generates a new RSA keypair
func GenerateRSAKeypair() (*rsa.PrivateKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, RSAKeySize)
	if err != nil {
		return nil, fmt.Errorf("failed to generate RSA key: %w", err)
	}
	return privateKey, nil
}

// EncodePrivateKeyPEM encodes a private key to PEM format
func EncodePrivateKeyPEM(privateKey *rsa.PrivateKey) ([]byte, error) {
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	})
	if privateKeyPEM == nil {
		return nil, fmt.Errorf("failed to encode private key to PEM")
	}
	return privateKeyPEM, nil
}

// DecodePrivateKeyPEM decodes a private key from PEM format
func DecodePrivateKeyPEM(pemData []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}
	if block.Type != "RSA PRIVATE KEY" {
		return nil, fmt.Errorf("unexpected PEM block type: %s", block.Type)
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}
	return privateKey, nil
}

// EncodePublicKeyPEM encodes a public key to PEM format
func EncodePublicKeyPEM(publicKey *rsa.PublicKey) ([]byte, error) {
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal public key: %w", err)
	}

	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	})
	if publicKeyPEM == nil {
		return nil, fmt.Errorf("failed to encode public key to PEM")
	}
	return publicKeyPEM, nil
}

// DecodePublicKeyPEM decodes a public key from PEM format
func DecodePublicKeyPEM(pemData []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}
	if block.Type != "PUBLIC KEY" {
		return nil, fmt.Errorf("unexpected PEM block type: %s", block.Type)
	}

	publicKeyInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	publicKey, ok := publicKeyInterface.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA public key")
	}

	return publicKey, nil
}

// GenerateFingerprint generates a SHA256 fingerprint of a public key
func GenerateFingerprint(publicKey *rsa.PublicKey) (string, error) {
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return "", fmt.Errorf("failed to marshal public key: %w", err)
	}

	hash := sha256.Sum256(publicKeyBytes)
	fingerprint := base64.StdEncoding.EncodeToString(hash[:])
	return fmt.Sprintf("SHA256:%s", fingerprint), nil
}
