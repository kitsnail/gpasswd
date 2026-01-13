package cli

import (
	"fmt"

	"github.com/kitsnail/gpasswd/internal/crypto"
	"github.com/spf13/cobra"
)

var (
	generateLength           int
	generateUseUppercase     bool
	generateUseLowercase     bool
	generateUseDigits        bool
	generateUseSymbols       bool
	generateExcludeAmbiguous bool
	generateShowStrength     bool
	generateCount            int
)

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate a secure random password",
	Long: `Generate a cryptographically secure random password with customizable options.

Examples:
  # Generate a 20-character password with all character types
  gpasswd generate

  # Generate a 32-character password
  gpasswd generate --length 32

  # Generate password without symbols
  gpasswd generate --no-symbols

  # Generate password excluding ambiguous characters (0, O, 1, l, I)
  gpasswd generate --exclude-ambiguous

  # Generate 5 passwords
  gpasswd generate --count 5

  # Show password strength analysis
  gpasswd generate --show-strength`,
	RunE: runGenerate,
}

func init() {
	rootCmd.AddCommand(generateCmd)

	// Define flags
	generateCmd.Flags().IntVarP(&generateLength, "length", "l", 20,
		"Length of the password (4-128)")
	generateCmd.Flags().BoolVar(&generateUseUppercase, "uppercase", true,
		"Include uppercase letters (A-Z)")
	generateCmd.Flags().BoolVar(&generateUseLowercase, "lowercase", true,
		"Include lowercase letters (a-z)")
	generateCmd.Flags().BoolVar(&generateUseDigits, "digits", true,
		"Include digits (0-9)")
	generateCmd.Flags().BoolVar(&generateUseSymbols, "symbols", true,
		"Include symbols (!@#$...)")
	generateCmd.Flags().BoolVar(&generateExcludeAmbiguous, "exclude-ambiguous", false,
		"Exclude ambiguous characters (0, O, 1, l, I)")
	generateCmd.Flags().BoolVarP(&generateShowStrength, "show-strength", "s", false,
		"Show password strength analysis")
	generateCmd.Flags().IntVarP(&generateCount, "count", "c", 1,
		"Number of passwords to generate (1-10)")

	// Add convenience flags
	generateCmd.Flags().BoolP("no-uppercase", "U", false, "Exclude uppercase letters")
	generateCmd.Flags().BoolP("no-lowercase", "L", false, "Exclude lowercase letters")
	generateCmd.Flags().BoolP("no-digits", "D", false, "Exclude digits")
	generateCmd.Flags().BoolP("no-symbols", "S", false, "Exclude symbols")
}

func runGenerate(cmd *cobra.Command, args []string) error {
	// Handle convenience "no-" flags
	if noUpper, _ := cmd.Flags().GetBool("no-uppercase"); noUpper {
		generateUseUppercase = false
	}
	if noLower, _ := cmd.Flags().GetBool("no-lowercase"); noLower {
		generateUseLowercase = false
	}
	if noDigits, _ := cmd.Flags().GetBool("no-digits"); noDigits {
		generateUseDigits = false
	}
	if noSymbols, _ := cmd.Flags().GetBool("no-symbols"); noSymbols {
		generateUseSymbols = false
	}

	// Validate count
	if generateCount < 1 || generateCount > 10 {
		return fmt.Errorf("count must be between 1 and 10")
	}

	// Build options
	options := crypto.GenerateOptions{
		UseUppercase:     generateUseUppercase,
		UseLowercase:     generateUseLowercase,
		UseDigits:        generateUseDigits,
		UseSymbols:       generateUseSymbols,
		ExcludeAmbiguous: generateExcludeAmbiguous,
	}

	// Check if at least one character type is selected
	if !options.UseUppercase && !options.UseLowercase &&
		!options.UseDigits && !options.UseSymbols {
		return fmt.Errorf("at least one character type must be enabled")
	}

	// Generate passwords
	for i := 0; i < generateCount; i++ {
		password, err := crypto.Generate(generateLength, options)
		if err != nil {
			return fmt.Errorf("failed to generate password: %w", err)
		}

		// Print password
		fmt.Println(password)

		// Show strength if requested
		if generateShowStrength {
			strength := crypto.CheckStrength(password)
			fmt.Printf("  Strength: %s (Score: %d/100)\n", strength.Level, strength.Score)
			if len(strength.Feedback) > 0 {
				fmt.Println("  Suggestions:")
				for _, feedback := range strength.Feedback {
					fmt.Printf("    - %s\n", feedback)
				}
			}
			if i < generateCount-1 {
				fmt.Println() // Empty line between passwords
			}
		}
	}

	return nil
}
