package helpers

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
)

func GetEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func GetLocalMachineName() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return hostname
}

func GetLocalMachineID() string {
	return GetLocalMachineName() + "-" + GenerateShortUniqueID()
}

func GenerateShortUniqueID() string {
	// Generate a random string of 6 characters not from uuid
	randomString := make([]byte, 6)
	rand.Read(randomString)
	return hex.EncodeToString(randomString)
}

func ParseKeyValue(kv string) (string, string, error) {
	// Split on first '=' sign
	for i := 0; i < len(kv); i++ {
		if kv[i] == '=' {
			if i == 0 {
				return "", "", fmt.Errorf("key cannot be empty")
			}
			key := kv[:i]
			value := kv[i+1:]
			return key, value, nil
		}
	}
	return "", "", fmt.Errorf("missing '=' separator")
}
