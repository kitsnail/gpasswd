package crypto

import (
	"testing"
)

func TestGenerate(t *testing.T) {
	tests := []struct {
		name    string
		length  int
		options GenerateOptions
		wantErr bool
	}{
		{
			name:   "default options - 20 characters",
			length: 20,
			options: GenerateOptions{
				UseUppercase: true,
				UseLowercase: true,
				UseDigits:    true,
				UseSymbols:   true,
			},
			wantErr: false,
		},
		{
			name:   "only lowercase",
			length: 16,
			options: GenerateOptions{
				UseLowercase: true,
			},
			wantErr: false,
		},
		{
			name:   "uppercase and digits",
			length: 12,
			options: GenerateOptions{
				UseUppercase: true,
				UseDigits:    true,
			},
			wantErr: false,
		},
		{
			name:   "all character types",
			length: 32,
			options: GenerateOptions{
				UseUppercase: true,
				UseLowercase: true,
				UseDigits:    true,
				UseSymbols:   true,
			},
			wantErr: false,
		},
		{
			name:   "exclude ambiguous characters",
			length: 20,
			options: GenerateOptions{
				UseUppercase:     true,
				UseLowercase:     true,
				UseDigits:        true,
				UseSymbols:       true,
				ExcludeAmbiguous: true,
			},
			wantErr: false,
		},
		{
			name:    "invalid - length too short",
			length:  3,
			options: GenerateOptions{UseUppercase: true},
			wantErr: true,
		},
		{
			name:    "invalid - length too long",
			length:  129,
			options: GenerateOptions{UseUppercase: true},
			wantErr: true,
		},
		{
			name:    "invalid - no character types selected",
			length:  16,
			options: GenerateOptions{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			password, err := Generate(tt.length, tt.options)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Generate() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Generate() unexpected error: %v", err)
				return
			}

			// Check password length
			if len(password) != tt.length {
				t.Errorf("Generate() password length = %d, want %d", len(password), tt.length)
			}

			// Check password is not empty
			if password == "" {
				t.Error("Generate() returned empty password")
			}

			// Verify character types are present (basic check)
			if tt.options.UseUppercase && !containsAny(password, uppercaseChars) {
				t.Error("Generate() should contain uppercase characters")
			}

			if tt.options.UseLowercase && !containsAny(password, lowercaseChars) {
				t.Error("Generate() should contain lowercase characters")
			}

			if tt.options.UseDigits && !containsAny(password, digitChars) {
				t.Error("Generate() should contain digits")
			}

			if tt.options.UseSymbols && !containsAny(password, symbolChars) {
				t.Error("Generate() should contain symbols")
			}

			// Check for ambiguous characters exclusion
			if tt.options.ExcludeAmbiguous {
				ambiguous := "0O1lI"
				if containsAny(password, ambiguous) {
					t.Error("Generate() should not contain ambiguous characters when excluded")
				}
			}
		})
	}
}

func TestGenerateRandomness(t *testing.T) {
	// Generate multiple passwords and ensure they're different
	options := GenerateOptions{
		UseUppercase: true,
		UseLowercase: true,
		UseDigits:    true,
		UseSymbols:   true,
	}

	passwords := make(map[string]bool)
	iterations := 100

	for i := 0; i < iterations; i++ {
		password, err := Generate(20, options)
		if err != nil {
			t.Fatalf("Generate() error: %v", err)
		}

		if passwords[password] {
			t.Errorf("Generate() produced duplicate password: %s", password)
		}
		passwords[password] = true
	}

	if len(passwords) != iterations {
		t.Errorf("Generate() uniqueness = %d, want %d", len(passwords), iterations)
	}
}

func TestCheckStrength(t *testing.T) {
	tests := []struct {
		name     string
		password string
		want     StrengthLevel
	}{
		{
			name:     "very weak - too short",
			password: "abc",
			want:     VeryWeak,
		},
		{
			name:     "weak - only lowercase (8 chars)",
			password: "abcdefgh",
			want:     Weak,
		},
		{
			name:     "weak - lowercase and digits",
			password: "abc12345",
			want:     Weak,
		},
		{
			name:     "fair - mixed case and digits",
			password: "Abc12345",
			want:     Fair,
		},
		{
			name:     "strong - mixed with symbols",
			password: "Abc123!@#",
			want:     Strong,
		},
		{
			name:     "very strong - long and complex",
			password: "Xk9$mP2@vL4#nR8&qT3!",
			want:     VeryStrong,
		},
		{
			name:     "very strong - passphrase",
			password: "correct-horse-battery-staple-2024!",
			want:     VeryStrong,
		},
		{
			name:     "weak - common password",
			password: "password",
			want:     VeryWeak,
		},
		{
			name:     "weak - common password with number",
			password: "password123",
			want:     VeryWeak,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CheckStrength(tt.password)

			if result.Level != tt.want {
				t.Errorf("CheckStrength() level = %v, want %v (score: %d)",
					result.Level, tt.want, result.Score)
			}

			// Verify score is within valid range
			if result.Score < 0 || result.Score > 100 {
				t.Errorf("CheckStrength() score = %d, want 0-100", result.Score)
			}

			// Verify feedback is provided for weak passwords
			if result.Level <= Fair && len(result.Feedback) == 0 {
				t.Error("CheckStrength() should provide feedback for weak passwords")
			}
		})
	}
}

func TestCheckStrengthEntropy(t *testing.T) {
	// Test that longer passwords generally have higher scores
	passwords := []string{
		"abc",                  // Very short
		"abcdefgh",             // Short
		"Abcdef12",             // Medium
		"Abc123!@#def",         // Good
		"Xk9$mP2@vL4#nR8&qT3!", // Very strong
	}

	var previousScore int
	for i, password := range passwords {
		result := CheckStrength(password)

		// Each password should generally have a higher score than the previous
		// (with some tolerance for edge cases)
		if i > 0 && result.Score < previousScore-5 {
			t.Errorf("CheckStrength() score regression: password %d (score %d) should be >= previous (score %d)",
				i, result.Score, previousScore)
		}
		previousScore = result.Score
	}
}

// Benchmark tests
func BenchmarkGenerate(b *testing.B) {
	options := GenerateOptions{
		UseUppercase: true,
		UseLowercase: true,
		UseDigits:    true,
		UseSymbols:   true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Generate(20, options)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCheckStrength(b *testing.B) {
	password := "Xk9$mP2@vL4#nR8&qT3!"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = CheckStrength(password)
	}
}
