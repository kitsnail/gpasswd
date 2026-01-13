package crypto

import (
	"bytes"
	"testing"
)

func TestEncrypt(t *testing.T) {
	tests := []struct {
		name      string
		plaintext []byte
		key       []byte
		wantErr   bool
	}{
		{
			name:      "valid - standard encryption",
			plaintext: []byte("Hello, World!"),
			key:       make([]byte, 32), // 256-bit key
			wantErr:   false,
		},
		{
			name:      "valid - empty plaintext",
			plaintext: []byte(""),
			key:       make([]byte, 32),
			wantErr:   false,
		},
		{
			name:      "valid - long plaintext",
			plaintext: bytes.Repeat([]byte("a"), 10000),
			key:       make([]byte, 32),
			wantErr:   false,
		},
		{
			name:      "invalid - nil plaintext",
			plaintext: nil,
			key:       make([]byte, 32),
			wantErr:   true,
		},
		{
			name:      "invalid - nil key",
			plaintext: []byte("test"),
			key:       nil,
			wantErr:   true,
		},
		{
			name:      "invalid - short key (128-bit)",
			plaintext: []byte("test"),
			key:       make([]byte, 16),
			wantErr:   true,
		},
		{
			name:      "invalid - wrong key size",
			plaintext: []byte("test"),
			key:       make([]byte, 30),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ciphertext, err := Encrypt(tt.plaintext, tt.key)

			if tt.wantErr {
				if err == nil {
					t.Error("Encrypt() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Encrypt() unexpected error: %v", err)
				return
			}

			// Verify ciphertext is not empty
			if len(ciphertext) == 0 {
				t.Error("Encrypt() returned empty ciphertext")
			}

			// Verify ciphertext is different from plaintext
			if bytes.Equal(ciphertext, tt.plaintext) {
				t.Error("Encrypt() ciphertext should be different from plaintext")
			}

			// Verify ciphertext is longer than plaintext (nonce + tag + data)
			// GCM nonce: 12 bytes, GCM tag: 16 bytes
			if len(ciphertext) < len(tt.plaintext)+12+16 {
				t.Errorf("Encrypt() ciphertext length = %d, want >= %d",
					len(ciphertext), len(tt.plaintext)+12+16)
			}
		})
	}
}

func TestDecrypt(t *testing.T) {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}

	tests := []struct {
		name       string
		ciphertext []byte
		key        []byte
		wantErr    bool
	}{
		{
			name:       "invalid - nil ciphertext",
			ciphertext: nil,
			key:        key,
			wantErr:    true,
		},
		{
			name:       "invalid - nil key",
			ciphertext: []byte("test"),
			key:        nil,
			wantErr:    true,
		},
		{
			name:       "invalid - short ciphertext",
			ciphertext: make([]byte, 10), // Less than nonce size
			key:        key,
			wantErr:    true,
		},
		{
			name:       "invalid - wrong key size",
			ciphertext: make([]byte, 50),
			key:        make([]byte, 16),
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Decrypt(tt.ciphertext, tt.key)

			if tt.wantErr && err == nil {
				t.Error("Decrypt() expected error, got nil")
			}

			if !tt.wantErr && err != nil {
				t.Errorf("Decrypt() unexpected error: %v", err)
			}
		})
	}
}

func TestEncryptDecryptRoundTrip(t *testing.T) {
	tests := []struct {
		name      string
		plaintext []byte
	}{
		{
			name:      "simple text",
			plaintext: []byte("Hello, World!"),
		},
		{
			name:      "empty text",
			plaintext: []byte(""),
		},
		{
			name:      "unicode text",
			plaintext: []byte("你好世界 🔐 Password Manager"),
		},
		{
			name:      "binary data",
			plaintext: []byte{0x00, 0x01, 0x02, 0xFF, 0xFE, 0xFD},
		},
		{
			name:      "long text",
			plaintext: bytes.Repeat([]byte("abcdefghijklmnopqrstuvwxyz"), 100),
		},
		{
			name:      "json data",
			plaintext: []byte(`{"username":"admin","password":"secret123","url":"https://example.com"}`),
		},
	}

	// Generate a proper encryption key
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encrypt
			ciphertext, err := Encrypt(tt.plaintext, key)
			if err != nil {
				t.Fatalf("Encrypt() error: %v", err)
			}

			// Decrypt
			decrypted, err := Decrypt(ciphertext, key)
			if err != nil {
				t.Fatalf("Decrypt() error: %v", err)
			}

			// Verify
			if !bytes.Equal(decrypted, tt.plaintext) {
				t.Errorf("Decrypt() = %v, want %v", decrypted, tt.plaintext)
			}
		})
	}
}

func TestEncryptDeterminism(t *testing.T) {
	// Encryption should NOT be deterministic (different nonce each time)
	plaintext := []byte("test message")
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}

	ciphertext1, err := Encrypt(plaintext, key)
	if err != nil {
		t.Fatalf("Encrypt() error: %v", err)
	}

	ciphertext2, err := Encrypt(plaintext, key)
	if err != nil {
		t.Fatalf("Encrypt() error: %v", err)
	}

	// Same plaintext + key should produce DIFFERENT ciphertext (random nonce)
	if bytes.Equal(ciphertext1, ciphertext2) {
		t.Error("Encrypt() should produce different ciphertext each time (random nonce)")
	}

	// But both should decrypt to the same plaintext
	decrypted1, err := Decrypt(ciphertext1, key)
	if err != nil {
		t.Fatalf("Decrypt() error: %v", err)
	}

	decrypted2, err := Decrypt(ciphertext2, key)
	if err != nil {
		t.Fatalf("Decrypt() error: %v", err)
	}

	if !bytes.Equal(decrypted1, plaintext) || !bytes.Equal(decrypted2, plaintext) {
		t.Error("Both ciphertexts should decrypt to original plaintext")
	}
}

func TestDecryptWithWrongKey(t *testing.T) {
	plaintext := []byte("secret message")

	// Encrypt with key1
	key1 := make([]byte, 32)
	for i := range key1 {
		key1[i] = byte(i)
	}

	ciphertext, err := Encrypt(plaintext, key1)
	if err != nil {
		t.Fatalf("Encrypt() error: %v", err)
	}

	// Try to decrypt with key2
	key2 := make([]byte, 32)
	for i := range key2 {
		key2[i] = byte(i + 1)
	}

	_, err = Decrypt(ciphertext, key2)
	if err == nil {
		t.Error("Decrypt() with wrong key should fail (GCM authentication)")
	}
}

func TestDecryptWithTamperedCiphertext(t *testing.T) {
	plaintext := []byte("important data")
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}

	ciphertext, err := Encrypt(plaintext, key)
	if err != nil {
		t.Fatalf("Encrypt() error: %v", err)
	}

	// Tamper with ciphertext (modify one byte)
	if len(ciphertext) > 20 {
		ciphertext[20] ^= 0xFF
	}

	// Decryption should fail due to GCM authentication tag mismatch
	_, err = Decrypt(ciphertext, key)
	if err == nil {
		t.Error("Decrypt() with tampered ciphertext should fail (GCM authentication)")
	}
}

func TestGenerateNonce(t *testing.T) {
	// Test default nonce generation
	nonce, err := GenerateNonce()
	if err != nil {
		t.Fatalf("GenerateNonce() error: %v", err)
	}

	if len(nonce) != DefaultNonceSize {
		t.Errorf("GenerateNonce() length = %d, want %d", len(nonce), DefaultNonceSize)
	}

	// Verify nonce is not all zeros
	allZeros := true
	for _, b := range nonce {
		if b != 0 {
			allZeros = false
			break
		}
	}
	if allZeros {
		t.Error("GenerateNonce() produced all-zero nonce")
	}
}

func TestGenerateNonceUniqueness(t *testing.T) {
	// Generate multiple nonces and ensure they're different
	nonces := make(map[string]bool)
	iterations := 1000

	for i := 0; i < iterations; i++ {
		nonce, err := GenerateNonce()
		if err != nil {
			t.Fatalf("GenerateNonce() error: %v", err)
		}

		nonceStr := string(nonce)
		if nonces[nonceStr] {
			t.Errorf("GenerateNonce() produced duplicate nonce on iteration %d", i)
		}
		nonces[nonceStr] = true
	}

	if len(nonces) != iterations {
		t.Errorf("GenerateNonce() uniqueness = %d, want %d", len(nonces), iterations)
	}
}

func TestIntegrationKeyDerivationAndEncryption(t *testing.T) {
	// Integration test: derive key from password, then encrypt/decrypt
	password := "my-secure-master-password-123!"

	// Generate salt
	salt, err := GenerateSalt()
	if err != nil {
		t.Fatalf("GenerateSalt() error: %v", err)
	}

	// Derive key
	params := DefaultArgon2Params()
	key, err := DeriveKey(password, salt, params)
	if err != nil {
		t.Fatalf("DeriveKey() error: %v", err)
	}

	// Encrypt data
	plaintext := []byte(`{"username":"admin","password":"P@ssw0rd!","url":"https://bank.com"}`)
	ciphertext, err := Encrypt(plaintext, key)
	if err != nil {
		t.Fatalf("Encrypt() error: %v", err)
	}

	// Decrypt data
	decrypted, err := Decrypt(ciphertext, key)
	if err != nil {
		t.Fatalf("Decrypt() error: %v", err)
	}

	// Verify
	if !bytes.Equal(decrypted, plaintext) {
		t.Errorf("Integration test failed: decrypted data doesn't match original")
	}

	// Test with wrong password
	wrongPassword := "wrong-password"
	wrongKey, err := DeriveKey(wrongPassword, salt, params)
	if err != nil {
		t.Fatalf("DeriveKey() with wrong password error: %v", err)
	}

	// Should fail to decrypt with wrong key
	_, err = Decrypt(ciphertext, wrongKey)
	if err == nil {
		t.Error("Integration test: should fail to decrypt with wrong password")
	}
}

// Benchmark tests
func BenchmarkEncrypt(b *testing.B) {
	plaintext := []byte("This is a secret message for benchmarking encryption performance")
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Encrypt(plaintext, key)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDecrypt(b *testing.B) {
	plaintext := []byte("This is a secret message for benchmarking decryption performance")
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}

	ciphertext, err := Encrypt(plaintext, key)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Decrypt(ciphertext, key)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGenerateNonce(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := GenerateNonce()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkEncryptLargeData(b *testing.B) {
	plaintext := bytes.Repeat([]byte("a"), 1024*1024) // 1 MB
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Encrypt(plaintext, key)
		if err != nil {
			b.Fatal(err)
		}
	}
}
