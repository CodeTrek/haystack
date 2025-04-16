package storage

import (
	"context"
	"haystack/conf"
	"os"
	"testing"
)

func TestCodec(t *testing.T) {
	// Set up a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "haystack-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Set configuration
	conf.Get().Global.DataPath = tempDir

	shutdown, cancel := context.WithCancel(context.Background())
	// Initialize storage
	err = Init(shutdown)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer CloseAndWait()
	defer cancel()

	// Test encoding and decoding
	testCases := []struct {
		name     string
		key      string
		value    string
		expected string
	}{
		{
			name:     "simple key value",
			key:      "test-key",
			value:    "test-value",
			expected: "test-value",
		},
		{
			name:     "empty value",
			key:      "empty-key",
			value:    "",
			expected: "",
		},
		{
			name:     "special characters",
			key:      "special-key",
			value:    "test\nvalue\twith\rspecial chars",
			expected: "test\nvalue\twith\rspecial chars",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Encode and save
			encodedKey := EncodeWorkspaceKey(tc.key)
			encodedValue := []byte(tc.value)
			err := db.Put(encodedKey, encodedValue)
			if err != nil {
				t.Fatalf("Failed to set value: %v", err)
			}

			// Read and decode
			value, err := db.Get(encodedKey)
			if err != nil {
				t.Fatalf("Failed to get value: %v", err)
			}

			if string(value) != tc.expected {
				t.Errorf("Value mismatch, got %s, want %s", string(value), tc.expected)
			}
		})
	}
}

func TestKeyEncoding(t *testing.T) {
	testCases := []struct {
		name     string
		key      string
		expected string
	}{
		{
			name:     "simple key",
			key:      "test-key",
			expected: "test-key",
		},
		{
			name:     "empty key",
			key:      "",
			expected: "",
		},
		{
			name:     "special characters",
			key:      "test\nkey\twith\rspecial chars",
			expected: "test\nkey\twith\rspecial chars",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			encoded := EncodeWorkspaceKey(tc.key)
			decoded := DecodeWorkspaceKey(string(encoded))
			if decoded != tc.expected {
				t.Errorf("Key mismatch, got %s, want %s", decoded, tc.expected)
			}
		})
	}
}
