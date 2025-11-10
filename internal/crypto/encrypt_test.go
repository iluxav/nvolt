package crypto

import (
	"bytes"
	"testing"
)

func TestGenerateAESKey(t *testing.T) {
	key, err := GenerateAESKey()
	if err != nil {
		t.Fatalf("Failed to generate AES key: %v", err)
	}

	if len(key) != AESKeySize {
		t.Errorf("Expected key size %d, got %d", AESKeySize, len(key))
	}

	// Generate another key and ensure they're different
	key2, err := GenerateAESKey()
	if err != nil {
		t.Fatalf("Failed to generate second AES key: %v", err)
	}

	if bytes.Equal(key, key2) {
		t.Error("Two generated keys should not be identical")
	}
}

func TestEncryptDecryptAESGCM(t *testing.T) {
	key, err := GenerateAESKey()
	if err != nil {
		t.Fatalf("Failed to generate AES key: %v", err)
	}

	plaintext := []byte("This is a secret message!")

	// Encrypt
	ciphertext, nonce, err := EncryptAESGCM(key, plaintext)
	if err != nil {
		t.Fatalf("Failed to encrypt: %v", err)
	}

	if len(ciphertext) == 0 {
		t.Error("Ciphertext is empty")
	}

	if len(nonce) == 0 {
		t.Error("Nonce is empty")
	}

	// Decrypt
	decrypted, err := DecryptAESGCM(key, ciphertext, nonce)
	if err != nil {
		t.Fatalf("Failed to decrypt: %v", err)
	}

	if !bytes.Equal(plaintext, decrypted) {
		t.Errorf("Decrypted text doesn't match original.\nExpected: %s\nGot: %s", plaintext, decrypted)
	}
}

func TestEncryptAESGCMDifferentNonces(t *testing.T) {
	key, err := GenerateAESKey()
	if err != nil {
		t.Fatalf("Failed to generate AES key: %v", err)
	}

	plaintext := []byte("Same message")

	// Encrypt twice
	ciphertext1, nonce1, err := EncryptAESGCM(key, plaintext)
	if err != nil {
		t.Fatalf("Failed to encrypt first time: %v", err)
	}

	ciphertext2, nonce2, err := EncryptAESGCM(key, plaintext)
	if err != nil {
		t.Fatalf("Failed to encrypt second time: %v", err)
	}

	// Nonces should be different
	if bytes.Equal(nonce1, nonce2) {
		t.Error("Nonces should be different for each encryption")
	}

	// Ciphertexts should be different (due to different nonces)
	if bytes.Equal(ciphertext1, ciphertext2) {
		t.Error("Ciphertexts should be different for each encryption")
	}

	// But both should decrypt to the same plaintext
	decrypted1, err := DecryptAESGCM(key, ciphertext1, nonce1)
	if err != nil {
		t.Fatalf("Failed to decrypt first ciphertext: %v", err)
	}

	decrypted2, err := DecryptAESGCM(key, ciphertext2, nonce2)
	if err != nil {
		t.Fatalf("Failed to decrypt second ciphertext: %v", err)
	}

	if !bytes.Equal(decrypted1, plaintext) || !bytes.Equal(decrypted2, plaintext) {
		t.Error("Both decryptions should produce original plaintext")
	}
}

func TestDecryptAESGCMWrongKey(t *testing.T) {
	key1, err := GenerateAESKey()
	if err != nil {
		t.Fatalf("Failed to generate AES key: %v", err)
	}

	key2, err := GenerateAESKey()
	if err != nil {
		t.Fatalf("Failed to generate second AES key: %v", err)
	}

	plaintext := []byte("Secret message")

	// Encrypt with key1
	ciphertext, nonce, err := EncryptAESGCM(key1, plaintext)
	if err != nil {
		t.Fatalf("Failed to encrypt: %v", err)
	}

	// Try to decrypt with key2
	_, err = DecryptAESGCM(key2, ciphertext, nonce)
	if err == nil {
		t.Error("Expected error when decrypting with wrong key")
	}
}

func TestDecryptAESGCMWrongNonce(t *testing.T) {
	key, err := GenerateAESKey()
	if err != nil {
		t.Fatalf("Failed to generate AES key: %v", err)
	}

	plaintext := []byte("Secret message")

	// Encrypt
	ciphertext, _, err := EncryptAESGCM(key, plaintext)
	if err != nil {
		t.Fatalf("Failed to encrypt: %v", err)
	}

	// Create a different nonce
	_, wrongNonce, err := EncryptAESGCM(key, []byte("different"))
	if err != nil {
		t.Fatalf("Failed to create wrong nonce: %v", err)
	}

	// Try to decrypt with wrong nonce
	_, err = DecryptAESGCM(key, ciphertext, wrongNonce)
	if err == nil {
		t.Error("Expected error when decrypting with wrong nonce")
	}
}

func TestEncryptAESGCMInvalidKeySize(t *testing.T) {
	invalidKey := []byte("short")
	plaintext := []byte("message")

	_, _, err := EncryptAESGCM(invalidKey, plaintext)
	if err == nil {
		t.Error("Expected error for invalid key size")
	}
}

func TestDecryptAESGCMInvalidKeySize(t *testing.T) {
	invalidKey := []byte("short")
	ciphertext := []byte("encrypted")
	nonce := make([]byte, 12)

	_, err := DecryptAESGCM(invalidKey, ciphertext, nonce)
	if err == nil {
		t.Error("Expected error for invalid key size")
	}
}

func TestEncryptDecryptEmptyData(t *testing.T) {
	key, err := GenerateAESKey()
	if err != nil {
		t.Fatalf("Failed to generate AES key: %v", err)
	}

	plaintext := []byte("")

	// Encrypt
	ciphertext, nonce, err := EncryptAESGCM(key, plaintext)
	if err != nil {
		t.Fatalf("Failed to encrypt empty data: %v", err)
	}

	// Decrypt
	decrypted, err := DecryptAESGCM(key, ciphertext, nonce)
	if err != nil {
		t.Fatalf("Failed to decrypt empty data: %v", err)
	}

	if !bytes.Equal(plaintext, decrypted) {
		t.Error("Decrypted empty data doesn't match")
	}
}

func TestEncryptDecryptLargeData(t *testing.T) {
	key, err := GenerateAESKey()
	if err != nil {
		t.Fatalf("Failed to generate AES key: %v", err)
	}

	// Create 1MB of data
	plaintext := make([]byte, 1024*1024)
	for i := range plaintext {
		plaintext[i] = byte(i % 256)
	}

	// Encrypt
	ciphertext, nonce, err := EncryptAESGCM(key, plaintext)
	if err != nil {
		t.Fatalf("Failed to encrypt large data: %v", err)
	}

	// Decrypt
	decrypted, err := DecryptAESGCM(key, ciphertext, nonce)
	if err != nil {
		t.Fatalf("Failed to decrypt large data: %v", err)
	}

	if !bytes.Equal(plaintext, decrypted) {
		t.Error("Decrypted large data doesn't match")
	}
}

func BenchmarkEncryptAESGCM(b *testing.B) {
	key, err := GenerateAESKey()
	if err != nil {
		b.Fatal(err)
	}

	plaintext := []byte("This is a benchmark message for encryption performance testing")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := EncryptAESGCM(key, plaintext)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDecryptAESGCM(b *testing.B) {
	key, err := GenerateAESKey()
	if err != nil {
		b.Fatal(err)
	}

	plaintext := []byte("This is a benchmark message for decryption performance testing")
	ciphertext, nonce, err := EncryptAESGCM(key, plaintext)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := DecryptAESGCM(key, ciphertext, nonce)
		if err != nil {
			b.Fatal(err)
		}
	}
}
