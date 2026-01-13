package crypto

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"unicode"
)

// Character sets for password generation
const (
	uppercaseChars          = "ABCDEFGHJKLMNPQRSTUVWXYZ"  // Excluded: I, O (ambiguous)
	lowercaseChars          = "abcdefghijkmnopqrstuvwxyz" // Excluded: l (ambiguous)
	digitChars              = "23456789"                  // Excluded: 0, 1 (ambiguous)
	symbolChars             = "!@#$%^&*()-_=+[]{}|;:,.<>?"
	uppercaseCharsAmbiguous = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	lowercaseCharsAmbiguous = "abcdefghijklmnopqrstuvwxyz"
	digitCharsAmbiguous     = "0123456789"
)

// Password length constraints
const (
	MinPasswordLength = 4
	MaxPasswordLength = 128
)

// GenerateOptions configures password generation
type GenerateOptions struct {
	UseUppercase     bool
	UseLowercase     bool
	UseDigits        bool
	UseSymbols       bool
	ExcludeAmbiguous bool
}

// StrengthLevel represents password strength
type StrengthLevel int

const (
	VeryWeak StrengthLevel = iota
	Weak
	Fair
	Strong
	VeryStrong
)

func (s StrengthLevel) String() string {
	switch s {
	case VeryWeak:
		return "Very Weak"
	case Weak:
		return "Weak"
	case Fair:
		return "Fair"
	case Strong:
		return "Strong"
	case VeryStrong:
		return "Very Strong"
	default:
		return "Unknown"
	}
}

// StrengthResult contains password strength analysis
type StrengthResult struct {
	Level    StrengthLevel
	Score    int      // 0-100
	Feedback []string // Suggestions for improvement
}

// Common weak passwords to check against
var commonPasswords = map[string]bool{
	"password":    true,
	"password1":   true,
	"password123": true,
	"12345678":    true,
	"123456789":   true,
	"qwerty":      true,
	"abc123":      true,
	"monkey":      true,
	"1234567":     true,
	"letmein":     true,
	"trustno1":    true,
	"dragon":      true,
	"baseball":    true,
	"111111":      true,
	"iloveyou":    true,
	"master":      true,
	"sunshine":    true,
	"ashley":      true,
	"bailey":      true,
	"passw0rd":    true,
	"shadow":      true,
	"123123":      true,
	"654321":      true,
	"superman":    true,
	"qazwsx":      true,
}

// Generate creates a random password with specified options
func Generate(length int, options GenerateOptions) (string, error) {
	return generateWithRetries(length, options, 0)
}

// generateWithRetries generates password with retry limit to prevent infinite recursion
func generateWithRetries(length int, options GenerateOptions, retryCount int) (string, error) {
	const maxRetries = 10

	// Validate length
	if length < MinPasswordLength {
		return "", fmt.Errorf("password length must be at least %d", MinPasswordLength)
	}
	if length > MaxPasswordLength {
		return "", fmt.Errorf("password length must not exceed %d", MaxPasswordLength)
	}

	// Build character set
	charset := buildCharset(options)
	if charset == "" {
		return "", errors.New("at least one character type must be enabled")
	}

	// Generate password
	password := make([]byte, length)
	charsetLen := big.NewInt(int64(len(charset)))

	for i := 0; i < length; i++ {
		// Use crypto/rand for cryptographically secure randomness
		randomIndex, err := rand.Int(rand.Reader, charsetLen)
		if err != nil {
			return "", fmt.Errorf("failed to generate random number: %w", err)
		}
		password[i] = charset[randomIndex.Int64()]
	}

	result := string(password)

	// Ensure at least one character from each enabled type is present
	if !meetsRequirements(result, options) {
		// Retry generation if requirements not met, up to max retries
		if retryCount < maxRetries {
			return generateWithRetries(length, options, retryCount+1)
		}
		// If max retries reached, force at least one character of each type
		return forceRequirements(password, options), nil
	}

	return result, nil
}

// buildCharset constructs the character set based on options
func buildCharset(options GenerateOptions) string {
	var charset strings.Builder

	if options.UseUppercase {
		if options.ExcludeAmbiguous {
			charset.WriteString(uppercaseChars)
		} else {
			charset.WriteString(uppercaseCharsAmbiguous)
		}
	}

	if options.UseLowercase {
		if options.ExcludeAmbiguous {
			charset.WriteString(lowercaseChars)
		} else {
			charset.WriteString(lowercaseCharsAmbiguous)
		}
	}

	if options.UseDigits {
		if options.ExcludeAmbiguous {
			charset.WriteString(digitChars)
		} else {
			charset.WriteString(digitCharsAmbiguous)
		}
	}

	if options.UseSymbols {
		charset.WriteString(symbolChars)
	}

	return charset.String()
}

// meetsRequirements checks if password contains at least one character from each enabled type
func meetsRequirements(password string, options GenerateOptions) bool {
	if options.UseUppercase && !containsAny(password, uppercaseCharsAmbiguous) {
		return false
	}
	if options.UseLowercase && !containsAny(password, lowercaseCharsAmbiguous) {
		return false
	}
	if options.UseDigits && !containsAny(password, digitCharsAmbiguous) {
		return false
	}
	if options.UseSymbols && !containsAny(password, symbolChars) {
		return false
	}
	return true
}

// containsAny checks if string contains any character from charset
func containsAny(s, charset string) bool {
	for _, c := range charset {
		if strings.ContainsRune(s, c) {
			return true
		}
	}
	return false
}

// forceRequirements ensures password contains at least one character of each required type
func forceRequirements(password []byte, options GenerateOptions) string {
	idx := 0

	if options.UseUppercase && !containsAny(string(password), uppercaseCharsAmbiguous) {
		if options.ExcludeAmbiguous {
			password[idx] = uppercaseChars[0]
		} else {
			password[idx] = uppercaseCharsAmbiguous[0]
		}
		idx++
	}

	if options.UseLowercase && !containsAny(string(password), lowercaseCharsAmbiguous) {
		if options.ExcludeAmbiguous {
			password[idx] = lowercaseChars[0]
		} else {
			password[idx] = lowercaseCharsAmbiguous[0]
		}
		idx++
	}

	if options.UseDigits && !containsAny(string(password), digitCharsAmbiguous) {
		if options.ExcludeAmbiguous {
			password[idx] = digitChars[0]
		} else {
			password[idx] = digitCharsAmbiguous[0]
		}
		idx++
	}

	if options.UseSymbols && !containsAny(string(password), symbolChars) {
		password[idx] = symbolChars[0]
	}

	return string(password)
}

// CheckStrength analyzes password strength
func CheckStrength(password string) StrengthResult {
	result := StrengthResult{
		Feedback: make([]string, 0),
	}

	// Check if it's a common password
	if commonPasswords[strings.ToLower(password)] {
		result.Level = VeryWeak
		result.Score = 0
		result.Feedback = append(result.Feedback, "This is a commonly used password")
		return result
	}

	score := 0

	// Length scoring (0-30 points)
	length := len(password)
	switch {
	case length < 6:
		score += length * 2
		result.Feedback = append(result.Feedback, "Password is too short (minimum 12 characters recommended)")
	case length < 8:
		score += length * 2
		result.Feedback = append(result.Feedback, "Password is too short (minimum 8 characters recommended)")
	case length < 12:
		score += 16 + (length-8)*2
	case length < 16:
		score += 24 + (length - 12)
	default:
		score += 30
	}

	// Character variety scoring (0-40 points)
	var hasUpper, hasLower, hasDigit, hasSymbol bool
	for _, c := range password {
		switch {
		case unicode.IsUpper(c):
			hasUpper = true
		case unicode.IsLower(c):
			hasLower = true
		case unicode.IsDigit(c):
			hasDigit = true
		case unicode.IsPunct(c) || unicode.IsSymbol(c):
			hasSymbol = true
		}
	}

	variety := 0
	if hasUpper {
		score += 10
		variety++
	} else {
		result.Feedback = append(result.Feedback, "Add uppercase letters")
	}

	if hasLower {
		score += 10
		variety++
	} else {
		result.Feedback = append(result.Feedback, "Add lowercase letters")
	}

	if hasDigit {
		score += 10
		variety++
	} else {
		result.Feedback = append(result.Feedback, "Add numbers")
	}

	if hasSymbol {
		score += 10
		variety++
	} else {
		result.Feedback = append(result.Feedback, "Add special characters")
	}

	// Bonus for using all character types
	if variety == 4 {
		score += 10
	}

	// Entropy estimation (0-20 points)
	entropy := calculateEntropy(password)
	entropyScore := int(entropy / 5) // Rough scaling
	if entropyScore > 20 {
		entropyScore = 20
	}
	score += entropyScore

	// Penalty for patterns (0-10 points deduction)
	if hasSequentialChars(password) {
		score -= 5
		result.Feedback = append(result.Feedback, "Avoid sequential characters (e.g., abc, 123)")
	}

	if hasRepeatedChars(password) {
		score -= 5
		result.Feedback = append(result.Feedback, "Avoid repeated characters")
	}

	// Ensure score is in valid range
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}

	result.Score = score

	// Determine strength level
	switch {
	case score < 20:
		result.Level = VeryWeak
	case score < 40:
		result.Level = Weak
	case score < 60:
		result.Level = Fair
	case score < 80:
		result.Level = Strong
	default:
		result.Level = VeryStrong
		result.Feedback = nil // No feedback needed for very strong passwords
	}

	return result
}

// calculateEntropy estimates password entropy
func calculateEntropy(password string) float64 {
	if len(password) == 0 {
		return 0
	}

	// Determine character space
	charSpace := 0
	var hasUpper, hasLower, hasDigit, hasSymbol bool

	for _, c := range password {
		switch {
		case unicode.IsUpper(c):
			hasUpper = true
		case unicode.IsLower(c):
			hasLower = true
		case unicode.IsDigit(c):
			hasDigit = true
		case unicode.IsPunct(c) || unicode.IsSymbol(c):
			hasSymbol = true
		}
	}

	if hasUpper {
		charSpace += 26
	}
	if hasLower {
		charSpace += 26
	}
	if hasDigit {
		charSpace += 10
	}
	if hasSymbol {
		charSpace += 32 // Approximate
	}

	if charSpace == 0 {
		return 0
	}

	// Entropy = log2(charSpace^length)
	// Simplified: length * log2(charSpace)
	// Using bit shifting approximation: log2(x) ≈ length(binary(x))
	// For more accuracy, we'd need math.Log2, but let's use a simple approximation
	var log2CharSpace float64
	switch {
	case charSpace >= 94:
		log2CharSpace = 6.5 // log2(94) ≈ 6.5
	case charSpace >= 62:
		log2CharSpace = 6.0 // log2(62) ≈ 6.0
	case charSpace >= 36:
		log2CharSpace = 5.2 // log2(36) ≈ 5.2
	case charSpace >= 26:
		log2CharSpace = 4.7 // log2(26) ≈ 4.7
	default:
		log2CharSpace = 3.3 // log2(10) ≈ 3.3
	}

	return float64(len(password)) * log2CharSpace
}

// hasSequentialChars checks for sequential character patterns
func hasSequentialChars(password string) bool {
	if len(password) < 3 {
		return false
	}

	for i := 0; i < len(password)-2; i++ {
		// Check for ascending sequence
		if password[i]+1 == password[i+1] && password[i+1]+1 == password[i+2] {
			return true
		}
		// Check for descending sequence
		if password[i]-1 == password[i+1] && password[i+1]-1 == password[i+2] {
			return true
		}
	}

	return false
}

// hasRepeatedChars checks for repeated character patterns
func hasRepeatedChars(password string) bool {
	if len(password) < 3 {
		return false
	}

	for i := 0; i < len(password)-2; i++ {
		if password[i] == password[i+1] && password[i+1] == password[i+2] {
			return true
		}
	}

	return false
}
