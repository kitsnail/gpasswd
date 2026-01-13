package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
)

// DefaultNonceSize is the standard nonce size for GCM (12 bytes / 96 bits)
const DefaultNonceSize = 12

// Minimum ciphertext size (nonce + tag)
const minCiphertextSize = DefaultNonceSize + 16 // 12 bytes nonce + 16 bytes GCM tag

// Encrypt encrypts plaintext using AES-256-GCM with the provided key
// The nonce is randomly generated and prepended to the ciphertext
// Format: [nonce (12 bytes)][encrypted data + GCM tag (16 bytes)]
//
// AES-256-GCM provides:
// - Confidentiality: Data is encrypted
// - Authenticity: Data cannot be tampered without detection
// - Integrity: Any modification will be detected during decryption
//
// Key must be 32 bytes (256 bits) for AES-256
func Encrypt(plaintext, key []byte) ([]byte, error) {
	// Validate inputs
	if plaintext == nil {
		return nil, errors.New("plaintext cannot be nil")
	}

	if key == nil {
		return nil, errors.New("key cannot be nil")
	}

	if len(key) != 32 {
		return nil, fmt.Errorf("key must be 32 bytes for AES-256, got %d", len(key))
	}

	// Create AES cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM mode: %w", err)
	}

	// Generate random nonce
	nonce, err := GenerateNonce()
	if err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt and authenticate
	// gcm.Seal appends the encrypted plaintext and authentication tag to nonce
	// We allocate the exact size needed: nonce + plaintext + tag
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)

	return ciphertext, nil
}

// Decrypt decrypts ciphertext using AES-256-GCM with the provided key
// The nonce is expected to be prepended to the ciphertext
// Format: [nonce (12 bytes)][encrypted data + GCM tag (16 bytes)]
//
// Returns error if:
// - Key is invalid
// - Ciphertext is too short
// - GCM authentication fails (wrong key or tampered data)
func Decrypt(ciphertext, key []byte) ([]byte, error) {
	// Validate inputs
	if ciphertext == nil {
		return nil, errors.New("ciphertext cannot be nil")
	}

	if key == nil {
		return nil, errors.New("key cannot be nil")
	}

	if len(key) != 32 {
		return nil, fmt.Errorf("key must be 32 bytes for AES-256, got %d", len(key))
	}

	// Check minimum ciphertext size
	if len(ciphertext) < minCiphertextSize {
		return nil, fmt.Errorf("ciphertext too short: must be at least %d bytes (nonce + tag), got %d",
			minCiphertextSize, len(ciphertext))
	}

	// Create AES cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM mode: %w", err)
	}

	// Extract nonce from the beginning of ciphertext
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short to contain nonce")
	}

	nonce := ciphertext[:nonceSize]
	encryptedData := ciphertext[nonceSize:]

	// Decrypt and verify authentication tag
	// gcm.Open will verify the authentication tag and return error if tampered
	plaintext, err := gcm.Open(nil, nonce, encryptedData, nil)
	if err != nil {
		return nil, fmt.Errorf("decryption failed (wrong key or tampered data): %w", err)
	}

	return plaintext, nil
}

// GenerateNonce generates a cryptographically secure random nonce
// for AES-GCM encryption. The nonce size is 12 bytes (96 bits) which
// is the standard size for GCM mode.
//
// Note: It's critical that nonces are NEVER reused with the same key.
// Each encryption operation must use a fresh random nonce.
func GenerateNonce() ([]byte, error) {
	nonce := make([]byte, DefaultNonceSize)

	// Use crypto/rand for cryptographically secure randomness
	_, err := rand.Read(nonce)
	if err != nil {
		return nil, fmt.Errorf("failed to generate random nonce: %w", err)
	}

	return nonce, nil
}
