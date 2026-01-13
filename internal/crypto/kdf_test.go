package crypto

import (
	"bytes"
	"testing"
)

func TestDeriveKey(t *testing.T) {
	tests := []struct {
		name     string
		password string
		salt     []byte
		params   Argon2Params
		wantErr  bool
	}{
		{
			name:     "valid - standard params",
			password: "my-strong-password-123!",
			salt:     make([]byte, 32),
			params: Argon2Params{
				Time:        3,
				Memory:      64 * 1024, // 64 MB
				Parallelism: 4,
				KeyLen:      32,
			},
			wantErr: false,
		},
		{
			name:     "valid - minimum params",
			password: "test",
			salt:     make([]byte, 16),
			params: Argon2Params{
				Time:        1,
				Memory:      8 * 1024, // 8 MB
				Parallelism: 1,
				KeyLen:      32,
			},
			wantErr: false,
		},
		{
			name:     "invalid - empty password",
			password: "",
			salt:     make([]byte, 32),
			params:   DefaultArgon2Params(),
			wantErr:  true,
		},
		{
			name:     "invalid - nil salt",
			password: "password",
			salt:     nil,
			params:   DefaultArgon2Params(),
			wantErr:  true,
		},
		{
			name:     "invalid - short salt",
			password: "password",
			salt:     make([]byte, 7),
			params:   DefaultArgon2Params(),
			wantErr:  true,
		},
		{
			name:     "invalid - zero time cost",
			password: "password",
			salt:     make([]byte, 32),
			params: Argon2Params{
				Time:        0,
				Memory:      64 * 1024,
				Parallelism: 4,
				KeyLen:      32,
			},
			wantErr: true,
		},
		{
			name:     "invalid - zero memory cost",
			password: "password",
			salt:     make([]byte, 32),
			params: Argon2Params{
				Time:        3,
				Memory:      0,
				Parallelism: 4,
				KeyLen:      32,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, err := DeriveKey(tt.password, tt.salt, tt.params)

			if tt.wantErr {
				if err == nil {
					t.Error("DeriveKey() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("DeriveKey() unexpected error: %v", err)
				return
			}

			// Verify key length
			if len(key) != int(tt.params.KeyLen) {
				t.Errorf("DeriveKey() key length = %d, want %d", len(key), tt.params.KeyLen)
			}

			// Verify key is not all zeros
			allZeros := true
			for _, b := range key {
				if b != 0 {
					allZeros = false
					break
				}
			}
			if allZeros {
				t.Error("DeriveKey() produced all-zero key")
			}
		})
	}
}

func TestDeriveKeyDeterministic(t *testing.T) {
	// Same password + salt + params should produce same key
	password := "test-password-123"
	salt := []byte("this-is-a-32-byte-salt-value")
	params := DefaultArgon2Params()

	key1, err := DeriveKey(password, salt, params)
	if err != nil {
		t.Fatalf("DeriveKey() error: %v", err)
	}

	key2, err := DeriveKey(password, salt, params)
	if err != nil {
		t.Fatalf("DeriveKey() error: %v", err)
	}

	if !bytes.Equal(key1, key2) {
		t.Error("DeriveKey() should be deterministic for same inputs")
	}
}

func TestDeriveKeyUniqueness(t *testing.T) {
	password := "test-password"
	params := DefaultArgon2Params()

	// Different salts should produce different keys
	salt1 := []byte("salt-one-32-bytes-long-valueXX")
	salt2 := []byte("salt-two-32-bytes-long-valueYY")

	key1, err := DeriveKey(password, salt1, params)
	if err != nil {
		t.Fatalf("DeriveKey() error: %v", err)
	}

	key2, err := DeriveKey(password, salt2, params)
	if err != nil {
		t.Fatalf("DeriveKey() error: %v", err)
	}

	if bytes.Equal(key1, key2) {
		t.Error("DeriveKey() should produce different keys for different salts")
	}

	// Different passwords should produce different keys
	password2 := "different-password"
	key3, err := DeriveKey(password2, salt1, params)
	if err != nil {
		t.Fatalf("DeriveKey() error: %v", err)
	}

	if bytes.Equal(key1, key3) {
		t.Error("DeriveKey() should produce different keys for different passwords")
	}
}

func TestGenerateSalt(t *testing.T) {
	// Test default salt generation
	salt, err := GenerateSalt()
	if err != nil {
		t.Fatalf("GenerateSalt() error: %v", err)
	}

	if len(salt) != DefaultSaltLength {
		t.Errorf("GenerateSalt() length = %d, want %d", len(salt), DefaultSaltLength)
	}

	// Verify salt is not all zeros
	allZeros := true
	for _, b := range salt {
		if b != 0 {
			allZeros = false
			break
		}
	}
	if allZeros {
		t.Error("GenerateSalt() produced all-zero salt")
	}
}

func TestGenerateSaltUniqueness(t *testing.T) {
	// Generate multiple salts and ensure they're different
	salts := make(map[string]bool)
	iterations := 100

	for i := 0; i < iterations; i++ {
		salt, err := GenerateSalt()
		if err != nil {
			t.Fatalf("GenerateSalt() error: %v", err)
		}

		saltStr := string(salt)
		if salts[saltStr] {
			t.Errorf("GenerateSalt() produced duplicate salt on iteration %d", i)
		}
		salts[saltStr] = true
	}

	if len(salts) != iterations {
		t.Errorf("GenerateSalt() uniqueness = %d, want %d", len(salts), iterations)
	}
}

func TestDefaultArgon2Params(t *testing.T) {
	params := DefaultArgon2Params()

	// Verify default values match documentation
	if params.Time != 3 {
		t.Errorf("DefaultArgon2Params() Time = %d, want 3", params.Time)
	}

	if params.Memory != 64*1024 {
		t.Errorf("DefaultArgon2Params() Memory = %d, want %d", params.Memory, 64*1024)
	}

	if params.Parallelism != 4 {
		t.Errorf("DefaultArgon2Params() Parallelism = %d, want 4", params.Parallelism)
	}

	if params.KeyLen != 32 {
		t.Errorf("DefaultArgon2Params() KeyLen = %d, want 32", params.KeyLen)
	}
}

func TestArgon2ParamsValidate(t *testing.T) {
	tests := []struct {
		name    string
		params  Argon2Params
		wantErr bool
	}{
		{
			name:    "valid - default params",
			params:  DefaultArgon2Params(),
			wantErr: false,
		},
		{
			name: "valid - minimum params",
			params: Argon2Params{
				Time:        1,
				Memory:      8 * 1024,
				Parallelism: 1,
				KeyLen:      16,
			},
			wantErr: false,
		},
		{
			name: "invalid - zero time",
			params: Argon2Params{
				Time:        0,
				Memory:      64 * 1024,
				Parallelism: 4,
				KeyLen:      32,
			},
			wantErr: true,
		},
		{
			name: "invalid - zero memory",
			params: Argon2Params{
				Time:        3,
				Memory:      0,
				Parallelism: 4,
				KeyLen:      32,
			},
			wantErr: true,
		},
		{
			name: "invalid - zero parallelism",
			params: Argon2Params{
				Time:        3,
				Memory:      64 * 1024,
				Parallelism: 0,
				KeyLen:      32,
			},
			wantErr: true,
		},
		{
			name: "invalid - zero key length",
			params: Argon2Params{
				Time:        3,
				Memory:      64 * 1024,
				Parallelism: 4,
				KeyLen:      0,
			},
			wantErr: true,
		},
		{
			name: "invalid - key length too short",
			params: Argon2Params{
				Time:        3,
				Memory:      64 * 1024,
				Parallelism: 4,
				KeyLen:      15,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.params.Validate()

			if tt.wantErr && err == nil {
				t.Error("Validate() expected error, got nil")
			}

			if !tt.wantErr && err != nil {
				t.Errorf("Validate() unexpected error: %v", err)
			}
		})
	}
}

// Benchmark tests
func BenchmarkDeriveKey(b *testing.B) {
	password := "test-password-for-benchmark"
	salt := make([]byte, 32)
	params := DefaultArgon2Params()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := DeriveKey(password, salt, params)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDeriveKeyLowMemory(b *testing.B) {
	password := "test-password-for-benchmark"
	salt := make([]byte, 32)
	params := Argon2Params{
		Time:        1,
		Memory:      8 * 1024, // 8 MB
		Parallelism: 2,
		KeyLen:      32,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := DeriveKey(password, salt, params)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGenerateSalt(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := GenerateSalt()
		if err != nil {
			b.Fatal(err)
		}
	}
}
