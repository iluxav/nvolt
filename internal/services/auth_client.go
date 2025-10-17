package services

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"iluxav/nvolt/internal/helpers"
	"iluxav/nvolt/internal/types"
)

type AuthClient struct {
	config *types.MachineLocalConfig
}

func NewAuthClient(config *types.MachineLocalConfig) *AuthClient {
	return &AuthClient{config: config}
}

// RequestChallenge requests an encrypted challenge from server
func (a *AuthClient) RequestChallenge(machineName string) (string, string, error) {

	url := fmt.Sprintf("%s/api/v1/auth/challenge", a.config.ServerURL)

	req := map[string]string{"machine_name": machineName}
	resp, err := helpers.CallAPIWithPayload[map[string]interface{}](url, "POST", "", &req)
	if err != nil {
		return "", "", err
	}

	challenge := (*resp)["challenge"].(string)
	challengeID := (*resp)["challenge_id"].(string)
	return challenge, challengeID, nil
}

// SignChallenge decrypts challenge with private key and signs it
func (a *AuthClient) SignChallenge(privateKeyPEM, encryptedChallenge string) (string, error) {
	// Parse private key
	block, _ := pem.Decode([]byte(privateKeyPEM))
	if block == nil {
		return "", fmt.Errorf("failed to decode PEM")
	}

	privKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return "", fmt.Errorf("failed to parse private key: %w", err)
	}

	fmt.Printf("DEBUG: Private key size: %d bits\n", privKey.N.BitLen())
	fmt.Printf("DEBUG: Encrypted challenge length: %d bytes\n", len(encryptedChallenge))

	// Decrypt challenge
	encrypted, err := base64.StdEncoding.DecodeString(encryptedChallenge)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64 challenge: %w", err)
	}
	fmt.Printf("DEBUG: Decoded encrypted data length: %d bytes\n", len(encrypted))

	decrypted, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, privKey, encrypted, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt challenge: %w", err)
	}

	// Sign the decrypted challenge
	hashed := sha256.Sum256(decrypted)
	signature, err := rsa.SignPSS(rand.Reader, privKey, crypto.SHA256, hashed[:], nil)
	if err != nil {
		return "", fmt.Errorf("failed to sign challenge: %w", err)
	}

	return base64.StdEncoding.EncodeToString(signature), nil
}

// VerifySignature sends signature to server and gets JWT
func (a *AuthClient) VerifySignature(machineName, challengeID, signature string) (string, error) {
	url := fmt.Sprintf("%s/api/v1/auth/verify", a.config.ServerURL)

	req := map[string]string{
		"machine_name": machineName,
		"challenge_id": challengeID,
		"signature":    signature,
	}

	resp, err := helpers.CallAPIWithPayload[map[string]interface{}](url, "POST", "", &req)
	if err != nil {
		return "", err
	}

	if !(*resp)["success"].(bool) {
		return "", fmt.Errorf("verification failed: %s", (*resp)["message"].(string))
	}

	return (*resp)["token"].(string), nil
}
