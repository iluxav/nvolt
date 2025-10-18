package helpers

import (
	"fmt"
	"os"
	"time"

	"github.com/oklog/ulid/v2"
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
	return GenerateShortUniqueID()
}

func GenerateShortUniqueID() string {
	return ulid.Make().String()
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

// FormatTimestamp formats a timestamp string in a readable format
// If the timestamp is empty or invalid, returns "N/A"
func FormatTimestamp(timestamp string) string {
	if timestamp == "" {
		return "N/A"
	}

	// Try parsing as RFC3339 (ISO 8601)
	t, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		// Try parsing as RFC3339Nano
		t, err = time.Parse(time.RFC3339Nano, timestamp)
		if err != nil {
			return "N/A"
		}
	}

	// Format as relative time if recent, otherwise as absolute date
	now := time.Now()
	diff := now.Sub(t)

	// If less than 24 hours ago, show relative time
	if diff < 24*time.Hour {
		if diff < time.Minute {
			return "just now"
		} else if diff < time.Hour {
			minutes := int(diff.Minutes())
			if minutes == 1 {
				return "1 minute ago"
			}
			return fmt.Sprintf("%d minutes ago", minutes)
		} else {
			hours := int(diff.Hours())
			if hours == 1 {
				return "1 hour ago"
			}
			return fmt.Sprintf("%d hours ago", hours)
		}
	}

	// If less than a week ago, show "X days ago"
	if diff < 7*24*time.Hour {
		days := int(diff.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	}

	// Otherwise show the date in a nice format: "Jan 2, 2006 15:04"
	return t.Format("Jan 2, 2006 15:04")
}
