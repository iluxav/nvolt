package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"fmt"
)

// WrapKey wraps a symmetric key using RSA-OAEP with a public key
func WrapKey(publicKey *rsa.PublicKey, key []byte) ([]byte, error) {
	if publicKey == nil {
		return nil, fmt.Errorf("public key is nil")
	}

	wrappedKey, err := rsa.EncryptOAEP(
		sha256.New(),
		rand.Reader,
		publicKey,
		key,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to wrap key: %w", err)
	}

	return wrappedKey, nil
}

// UnwrapKey unwraps a symmetric key using RSA-OAEP with a private key
func UnwrapKey(privateKey *rsa.PrivateKey, wrappedKey []byte) ([]byte, error) {
	if privateKey == nil {
		return nil, fmt.Errorf("private key is nil")
	}

	key, err := rsa.DecryptOAEP(
		sha256.New(),
		rand.Reader,
		privateKey,
		wrappedKey,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to unwrap key: %w", err)
	}

	return key, nil
}
