package crypto

import (
	"crypto/rand"
	"errors"
	"fmt"

	"golang.org/x/crypto/argon2"
)

// Default salt length in bytes (32 bytes = 256 bits)
const DefaultSaltLength = 32

// Minimum salt length for security
const MinSaltLength = 8

// Argon2Params defines parameters for Argon2id key derivation
type Argon2Params struct {
	Time        uint32 // Number of iterations (time cost)
	Memory      uint32 // Memory cost in KB
	Parallelism uint8  // Number of parallel threads
	KeyLen      uint32 // Length of derived key in bytes
}

// DefaultArgon2Params returns recommended Argon2id parameters
// Based on RFC 9106 recommendations for interactive use
func DefaultArgon2Params() Argon2Params {
	return Argon2Params{
		Time:        3,        // 3 iterations
		Memory:      64 * 1024, // 64 MB
		Parallelism: 4,        // 4 threads
		KeyLen:      32,       // 32 bytes (256 bits) for AES-256
	}
}

// Validate checks if Argon2Params are valid
func (p Argon2Params) Validate() error {
	if p.Time == 0 {
		return errors.New("time cost must be greater than 0")
	}

	if p.Memory == 0 {
		return errors.New("memory cost must be greater than 0")
	}

	if p.Parallelism == 0 {
		return errors.New("parallelism must be greater than 0")
	}

	if p.KeyLen == 0 {
		return errors.New("key length must be greater than 0")
	}

	// Enforce minimum key length for security
	if p.KeyLen < 16 {
		return errors.New("key length must be at least 16 bytes")
	}

	// Sanity check: memory should be reasonable
	if p.Memory < 8*1024 {
		return errors.New("memory cost should be at least 8 MB (8192 KB)")
	}

	return nil
}

// DeriveKey derives a cryptographic key from a password using Argon2id
// Argon2id is the recommended variant as it provides protection against
// both side-channel attacks (Argon2i) and GPU cracking attacks (Argon2d)
func DeriveKey(password string, salt []byte, params Argon2Params) ([]byte, error) {
	// Validate inputs
	if password == "" {
		return nil, errors.New("password cannot be empty")
	}

	if salt == nil {
		return nil, errors.New("salt cannot be nil")
	}

	if len(salt) < MinSaltLength {
		return nil, fmt.Errorf("salt must be at least %d bytes, got %d", MinSaltLength, len(salt))
	}

	// Validate parameters
	if err := params.Validate(); err != nil {
		return nil, fmt.Errorf("invalid Argon2 parameters: %w", err)
	}

	// Derive key using Argon2id
	// Argon2id combines the memory-hard properties of Argon2i and Argon2d
	key := argon2.IDKey(
		[]byte(password),
		salt,
		params.Time,
		params.Memory,
		params.Parallelism,
		params.KeyLen,
	)

	return key, nil
}

// GenerateSalt generates a cryptographically secure random salt
func GenerateSalt() ([]byte, error) {
	return GenerateSaltWithLength(DefaultSaltLength)
}

// GenerateSaltWithLength generates a random salt of specified length
func GenerateSaltWithLength(length int) ([]byte, error) {
	if length < MinSaltLength {
		return nil, fmt.Errorf("salt length must be at least %d bytes", MinSaltLength)
	}

	salt := make([]byte, length)

	// Use crypto/rand for cryptographically secure randomness
	_, err := rand.Read(salt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate random salt: %w", err)
	}

	return salt, nil
}
