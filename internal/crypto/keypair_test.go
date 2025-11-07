package crypto

import (
	"crypto/rsa"
	"strings"
	"testing"
)

func TestGenerateRSAKeypair(t *testing.T) {
	privateKey, err := GenerateRSAKeypair()
	if err != nil {
		t.Fatalf("Failed to generate keypair: %v", err)
	}

	if privateKey == nil {
		t.Fatal("Private key is nil")
	}

	if privateKey.N.BitLen() != RSAKeySize {
		t.Errorf("Expected key size %d, got %d", RSAKeySize, privateKey.N.BitLen())
	}
}

func TestEncodeDecodePrivateKeyPEM(t *testing.T) {
	privateKey, err := GenerateRSAKeypair()
	if err != nil {
		t.Fatalf("Failed to generate keypair: %v", err)
	}

	// Encode
	pemData, err := EncodePrivateKeyPEM(privateKey)
	if err != nil {
		t.Fatalf("Failed to encode private key: %v", err)
	}

	if !strings.Contains(string(pemData), "BEGIN RSA PRIVATE KEY") {
		t.Error("PEM data doesn't contain expected header")
	}

	// Decode
	decodedKey, err := DecodePrivateKeyPEM(pemData)
	if err != nil {
		t.Fatalf("Failed to decode private key: %v", err)
	}

	// Compare
	if privateKey.N.Cmp(decodedKey.N) != 0 {
		t.Error("Decoded key doesn't match original")
	}
}

func TestEncodeDecodePublicKeyPEM(t *testing.T) {
	privateKey, err := GenerateRSAKeypair()
	if err != nil {
		t.Fatalf("Failed to generate keypair: %v", err)
	}

	publicKey := &privateKey.PublicKey

	// Encode
	pemData, err := EncodePublicKeyPEM(publicKey)
	if err != nil {
		t.Fatalf("Failed to encode public key: %v", err)
	}

	if !strings.Contains(string(pemData), "BEGIN PUBLIC KEY") {
		t.Error("PEM data doesn't contain expected header")
	}

	// Decode
	decodedKey, err := DecodePublicKeyPEM(pemData)
	if err != nil {
		t.Fatalf("Failed to decode public key: %v", err)
	}

	// Compare
	if publicKey.N.Cmp(decodedKey.N) != 0 {
		t.Error("Decoded key doesn't match original")
	}
}

func TestGenerateFingerprint(t *testing.T) {
	privateKey, err := GenerateRSAKeypair()
	if err != nil {
		t.Fatalf("Failed to generate keypair: %v", err)
	}

	publicKey := &privateKey.PublicKey

	fingerprint, err := GenerateFingerprint(publicKey)
	if err != nil {
		t.Fatalf("Failed to generate fingerprint: %v", err)
	}

	if !strings.HasPrefix(fingerprint, "SHA256:") {
		t.Errorf("Fingerprint should start with 'SHA256:', got: %s", fingerprint)
	}

	// Generate again and verify consistency
	fingerprint2, err := GenerateFingerprint(publicKey)
	if err != nil {
		t.Fatalf("Failed to generate fingerprint: %v", err)
	}

	if fingerprint != fingerprint2 {
		t.Error("Fingerprints should be consistent")
	}

	// Different key should produce different fingerprint
	privateKey2, err := GenerateRSAKeypair()
	if err != nil {
		t.Fatalf("Failed to generate second keypair: %v", err)
	}

	fingerprint3, err := GenerateFingerprint(&privateKey2.PublicKey)
	if err != nil {
		t.Fatalf("Failed to generate fingerprint: %v", err)
	}

	if fingerprint == fingerprint3 {
		t.Error("Different keys should produce different fingerprints")
	}
}

func TestDecodePrivateKeyPEMInvalidData(t *testing.T) {
	tests := []struct {
		name    string
		pemData []byte
	}{
		{"empty data", []byte("")},
		{"invalid pem", []byte("not a pem")},
		{"wrong type", []byte("-----BEGIN CERTIFICATE-----\nAAA\n-----END CERTIFICATE-----")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := DecodePrivateKeyPEM(tt.pemData)
			if err == nil {
				t.Error("Expected error for invalid data")
			}
		})
	}
}

func TestDecodePublicKeyPEMInvalidData(t *testing.T) {
	tests := []struct {
		name    string
		pemData []byte
	}{
		{"empty data", []byte("")},
		{"invalid pem", []byte("not a pem")},
		{"wrong type", []byte("-----BEGIN CERTIFICATE-----\nAAA\n-----END CERTIFICATE-----")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := DecodePublicKeyPEM(tt.pemData)
			if err == nil {
				t.Error("Expected error for invalid data")
			}
		})
	}
}

func BenchmarkGenerateRSAKeypair(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := GenerateRSAKeypair()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGenerateFingerprint(b *testing.B) {
	privateKey, err := GenerateRSAKeypair()
	if err != nil {
		b.Fatal(err)
	}
	publicKey := &privateKey.PublicKey

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := GenerateFingerprint(publicKey)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Helper function for tests
func generateTestKeypair(t *testing.T) *rsa.PrivateKey {
	t.Helper()
	key, err := GenerateRSAKeypair()
	if err != nil {
		t.Fatalf("Failed to generate test keypair: %v", err)
	}
	return key
}
