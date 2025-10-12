package crypto

import (
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"iluxav/nvolt/internal/types"
)

// KeyPair represents a public/private key pair

// Encrypt encrypts plaintext using the provided RSA public key (supports both PEM and base64-encoded PEM)
func Encrypt(publicKeyInput string, plaintext string) (string, error) {
	var block *pem.Block

	// Try to decode as plain PEM first
	block, _ = pem.Decode([]byte(publicKeyInput))

	// If that fails, try base64 decoding first (backward compatibility)
	if block == nil {
		decoded, err := base64.StdEncoding.DecodeString(publicKeyInput)
		if err == nil {
			block, _ = pem.Decode(decoded)
		}
	}

	if block == nil {
		return "", fmt.Errorf("failed to decode PEM block")
	}

	// Parse public key
	parsedKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return "", fmt.Errorf("failed to parse public key: %w", err)
	}

	// Cast to RSA public key
	rsaPublicKey, ok := parsedKey.(*rsa.PublicKey)
	if !ok {
		return "", fmt.Errorf("not an RSA public key")
	}

	// Encrypt using RSA OAEP (more secure than PKCS1v15)
	encrypted, err := rsa.EncryptOAEP(
		sha256.New(),
		rand.Reader,
		rsaPublicKey,
		[]byte(plaintext),
		nil,
	)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt: %w", err)
	}

	// Return base64-encoded encrypted data
	return base64.StdEncoding.EncodeToString(encrypted), nil
}

// Decrypt decrypts ciphertext using the provided RSA private key (supports both PEM and base64-encoded PEM)
func Decrypt(privateKeyInput string, ciphertextBase64 string) (string, error) {
	var block *pem.Block

	// Try to decode as plain PEM first
	block, _ = pem.Decode([]byte(privateKeyInput))

	// If that fails, try base64 decoding first (backward compatibility)
	if block == nil {
		decoded, err := base64.StdEncoding.DecodeString(privateKeyInput)
		if err == nil {
			block, _ = pem.Decode(decoded)
		}
	}

	if block == nil {
		return "", fmt.Errorf("failed to decode PEM block")
	}

	// Parse private key
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return "", fmt.Errorf("failed to parse private key: %w", err)
	}

	// Decode base64-encoded ciphertext
	ciphertext, err := base64.StdEncoding.DecodeString(ciphertextBase64)
	if err != nil {
		return "", fmt.Errorf("failed to decode ciphertext: %w", err)
	}

	// Decrypt using RSA OAEP
	decrypted, err := rsa.DecryptOAEP(
		sha256.New(),
		rand.Reader,
		privateKey,
		ciphertext,
		nil,
	)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}

	return string(decrypted), nil
}

// GenerateKeyPair generates a new RSA key pair for asymmetric encryption
func GenerateKeyPair() (*types.KeyPair, error) {
	// Generate RSA key pair (2048 bits)
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate RSA key: %w", err)
	}

	// Encode private key to PEM format
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyPEMBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	}
	privateKeyPEMBytes := pem.EncodeToMemory(privateKeyPEMBlock)

	// Encode public key to PEM format
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal public key: %w", err)
	}

	publicKeyPEMBlock := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	}
	publicKeyPEMBytes := pem.EncodeToMemory(publicKeyPEMBlock)

	// Return as PEM strings (not base64 encoded)
	return &types.KeyPair{
		PrivateKey: string(privateKeyPEMBytes),
		PublicKey:  string(publicKeyPEMBytes),
	}, nil
}

// ExtractPublicKey extracts the public key from a private key (supports both PEM and base64-encoded PEM)
func ExtractPublicKey(privateKeyInput string) (string, error) {
	var block *pem.Block

	// Try to decode as plain PEM first
	block, _ = pem.Decode([]byte(privateKeyInput))

	// If that fails, try base64 decoding first (backward compatibility with old format)
	if block == nil {
		decoded, err := base64.StdEncoding.DecodeString(privateKeyInput)
		if err == nil {
			block, _ = pem.Decode(decoded)
		}
	}

	if block == nil {
		return "", fmt.Errorf("failed to decode PEM block")
	}

	// Parse private key
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return "", fmt.Errorf("failed to parse private key: %w", err)
	}

	// Extract public key
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return "", fmt.Errorf("failed to marshal public key: %w", err)
	}

	publicKeyPEMBlock := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	}
	publicKeyPEMBytes := pem.EncodeToMemory(publicKeyPEMBlock)

	// Return as PEM string
	return string(publicKeyPEMBytes), nil
}

// ValidatePrivateKey validates that a private key is in the correct format (supports both PEM and base64-encoded PEM)
func ValidatePrivateKey(privateKeyInput string) error {
	var block *pem.Block

	// Try to decode as plain PEM first
	block, _ = pem.Decode([]byte(privateKeyInput))

	// If that fails, try base64 decoding first (backward compatibility)
	if block == nil {
		decoded, err := base64.StdEncoding.DecodeString(privateKeyInput)
		if err == nil {
			block, _ = pem.Decode(decoded)
		}
	}

	if block == nil {
		return fmt.Errorf("failed to decode PEM block")
	}

	// Parse private key
	_, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse private key: %w", err)
	}

	return nil
}

// ValidatePublicKey validates that a public key is in the correct format
func ValidatePublicKey(publicKey string) error {
	// Decode from base64
	pemBytes, err := base64.StdEncoding.DecodeString(publicKey)
	if err != nil {
		return fmt.Errorf("invalid base64 encoding: %w", err)
	}

	// Decode PEM
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return fmt.Errorf("failed to decode PEM block")
	}

	// Parse public key
	_, err = x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse public key: %w", err)
	}

	return nil
}

func DecryptAllVars(privateKey string, vars map[string]string) (map[string]string, error) {
	decryptedVars := make(map[string]string)

	for key, encryptedValue := range vars {
		decryptedValue, err := Decrypt(privateKey, encryptedValue)
		if err != nil {
			fmt.Printf("Failed to decrypt %s: %v", key, err)
			continue
		}
		decryptedVars[key] = decryptedValue
	}
	return decryptedVars, nil
}

// GenerateMasterKey generates a random 256-bit AES key for symmetric encryption
func GenerateMasterKey() ([]byte, error) {
	key := make([]byte, 32) // 256 bits
	_, err := rand.Read(key)
	if err != nil {
		return nil, fmt.Errorf("failed to generate master key: %w", err)
	}
	return key, nil
}

// EncryptWithMasterKey encrypts plaintext using AES-256-GCM with the provided master key
func EncryptWithMasterKey(masterKey []byte, plaintext string) (string, error) {
	block, err := aes.NewCipher(masterKey)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptWithMasterKey decrypts ciphertext using AES-256-GCM with the provided master key
func DecryptWithMasterKey(masterKey []byte, ciphertextBase64 string) (string, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(ciphertextBase64)
	if err != nil {
		return "", fmt.Errorf("failed to decode ciphertext: %w", err)
	}

	block, err := aes.NewCipher(masterKey)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}

	return string(plaintext), nil
}

// WrapMasterKey encrypts the master key with a user's public RSA key
func WrapMasterKey(publicKeyBase64 string, masterKey []byte) (string, error) {
	// Encode master key to base64 first
	masterKeyBase64 := base64.StdEncoding.EncodeToString(masterKey)
	// Use existing RSA encryption
	return Encrypt(publicKeyBase64, masterKeyBase64)
}

// UnwrapMasterKey decrypts the wrapped master key with a user's private RSA key
func UnwrapMasterKey(privateKeyBase64 string, wrappedKeyBase64 string) ([]byte, error) {
	// Decrypt using RSA
	masterKeyBase64, err := Decrypt(privateKeyBase64, wrappedKeyBase64)
	if err != nil {
		return nil, fmt.Errorf("failed to unwrap master key: %w", err)
	}
	// Decode from base64
	masterKey, err := base64.StdEncoding.DecodeString(masterKeyBase64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode master key: %w", err)
	}
	return masterKey, nil
}

// SignMessage signs a message using RSA private key (supports both PEM and base64-encoded PEM) with SHA256
func SignMessage(privateKeyInput string, message string) (string, error) {
	var block *pem.Block

	// Try to decode as plain PEM first
	block, _ = pem.Decode([]byte(privateKeyInput))

	// If that fails, try base64 decoding first (backward compatibility)
	if block == nil {
		decoded, err := base64.StdEncoding.DecodeString(privateKeyInput)
		if err == nil {
			block, _ = pem.Decode(decoded)
		}
	}

	if block == nil {
		return "", fmt.Errorf("failed to decode PEM block")
	}

	// Parse private key
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return "", fmt.Errorf("failed to parse private key: %w", err)
	}

	// Hash the message
	hashed := sha256.Sum256([]byte(message))

	// Sign the hash
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hashed[:])
	if err != nil {
		return "", fmt.Errorf("failed to sign message: %w", err)
	}

	// Return base64-encoded signature
	return base64.StdEncoding.EncodeToString(signature), nil
}

// VerifySignature verifies a signature using RSA public key (supports both PEM and base64-encoded PEM) with SHA256
func VerifySignature(publicKeyInput string, message string, signatureBase64 string) error {
	var block *pem.Block

	// Try to decode as plain PEM first
	block, _ = pem.Decode([]byte(publicKeyInput))

	// If that fails, try base64 decoding first (backward compatibility)
	if block == nil {
		decoded, err := base64.StdEncoding.DecodeString(publicKeyInput)
		if err == nil {
			block, _ = pem.Decode(decoded)
		}
	}

	if block == nil {
		return fmt.Errorf("failed to decode PEM block")
	}

	// Parse public key
	parsedKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse public key: %w", err)
	}

	// Cast to RSA public key
	rsaPublicKey, ok := parsedKey.(*rsa.PublicKey)
	if !ok {
		return fmt.Errorf("not an RSA public key")
	}

	// Decode signature
	signature, err := base64.StdEncoding.DecodeString(signatureBase64)
	if err != nil {
		return fmt.Errorf("failed to decode signature: %w", err)
	}

	// Hash the message
	hashed := sha256.Sum256([]byte(message))

	// Verify the signature
	err = rsa.VerifyPKCS1v15(rsaPublicKey, crypto.SHA256, hashed[:], signature)
	if err != nil {
		return fmt.Errorf("signature verification failed: %w", err)
	}

	return nil
}
