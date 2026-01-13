package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"

	"github.com/kitsnail/gpasswd/internal/crypto"
	"github.com/kitsnail/gpasswd/internal/storage"
	"github.com/kitsnail/gpasswd/pkg/config"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new password vault",
	Long: `Initialize a new password vault with a master password.

This command will:
1. Prompt for a master password (with confirmation)
2. Check password strength
3. Generate a cryptographic salt
4. Initialize the encrypted database
5. Store Argon2 parameters

The vault will be created at: ~/.gpasswd/vault.db

Example:
  gpasswd init`,
	RunE: runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Determine database path
	dbPath := cfg.Database.Path
	if dbPath == "" {
		// Default to ~/.gpasswd/vault.db
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		dbPath = filepath.Join(homeDir, ".gpasswd", "vault.db")
	}

	// Check if vault already exists
	if _, err := os.Stat(dbPath); err == nil {
		fmt.Fprintf(os.Stderr, "⚠️  Vault already exists at: %s\n", dbPath)

		var overwrite bool
		prompt := &survey.Confirm{
			Message: "Do you want to overwrite the existing vault? (ALL DATA WILL BE LOST)",
			Default: false,
		}
		if err := survey.AskOne(prompt, &overwrite); err != nil {
			return fmt.Errorf("prompt failed: %w", err)
		}

		if !overwrite {
			fmt.Println("✓ Initialization cancelled")
			return nil
		}

		// Remove existing vault
		if err := os.RemoveAll(filepath.Dir(dbPath)); err != nil {
			return fmt.Errorf("failed to remove existing vault: %w", err)
		}
	}

	// Prompt for master password
	var masterPassword string
	passwordPrompt := &survey.Password{
		Message: "Enter master password:",
	}
	if err := survey.AskOne(passwordPrompt, &masterPassword, survey.WithValidator(survey.Required)); err != nil {
		return fmt.Errorf("password prompt failed: %w", err)
	}

	// Check password strength
	strength := crypto.CheckStrength(masterPassword)
	fmt.Printf("\n🔐 Password Strength: %s (Score: %d/100)\n", strength.Level.String(), strength.Score)

	if strength.Level < crypto.Fair {
		fmt.Println("\n⚠️  Your password is weak. Consider:")
		for _, feedback := range strength.Feedback {
			fmt.Printf("   • %s\n", feedback)
		}

		var continueWeak bool
		confirmPrompt := &survey.Confirm{
			Message: "Continue with this weak password?",
			Default: false,
		}
		if err := survey.AskOne(confirmPrompt, &continueWeak); err != nil {
			return fmt.Errorf("confirmation failed: %w", err)
		}

		if !continueWeak {
			fmt.Println("✓ Initialization cancelled. Please choose a stronger password.")
			return nil
		}
	}

	// Confirm password
	var confirmPassword string
	confirmPrompt := &survey.Password{
		Message: "Confirm master password:",
	}
	if err := survey.AskOne(confirmPrompt, &confirmPassword, survey.WithValidator(survey.Required)); err != nil {
		return fmt.Errorf("confirmation prompt failed: %w", err)
	}

	if masterPassword != confirmPassword {
		return fmt.Errorf("passwords do not match")
	}

	fmt.Println("\n🔧 Initializing vault...")

	// Generate cryptographic salt
	fmt.Println("   • Generating cryptographic salt...")
	salt, err := crypto.GenerateSalt()
	if err != nil {
		return fmt.Errorf("failed to generate salt: %w", err)
	}

	// Get Argon2 parameters from config or use defaults
	var argon2Params crypto.Argon2Params
	if cfg.Security.Argon2.Time > 0 {
		argon2Params = crypto.Argon2Params{
			Time:        cfg.Security.Argon2.Time,
			Memory:      cfg.Security.Argon2.Memory,
			Parallelism: cfg.Security.Argon2.Parallelism,
			KeyLen:      cfg.Security.Argon2.KeyLength,
		}
	} else {
		argon2Params = crypto.DefaultArgon2Params()
	}

	// Validate parameters
	if err := argon2Params.Validate(); err != nil {
		return fmt.Errorf("invalid Argon2 parameters: %w", err)
	}

	// Test key derivation (to verify password works)
	fmt.Println("   • Deriving encryption key (this may take a moment)...")
	_, err = crypto.DeriveKey(masterPassword, salt, argon2Params)
	if err != nil {
		return fmt.Errorf("failed to derive key: %w", err)
	}

	// Initialize database
	fmt.Printf("   • Creating database at: %s\n", dbPath)
	db, err := storage.InitDB(dbPath)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer db.Close()

	// Store salt
	fmt.Println("   • Storing cryptographic salt...")
	if err := db.SetSalt(salt); err != nil {
		return fmt.Errorf("failed to store salt: %w", err)
	}

	// Store Argon2 parameters
	fmt.Println("   • Storing key derivation parameters...")
	if err := db.SetArgon2Params(argon2Params); err != nil {
		return fmt.Errorf("failed to store Argon2 parameters: %w", err)
	}

	// Store metadata
	if err := db.SetMetadata("version", Version); err != nil {
		return fmt.Errorf("failed to store version: %w", err)
	}

	if err := db.SetMetadata("created_at", fmt.Sprintf("%d", os.Getpid())); err != nil {
		// Non-critical, just log
		fmt.Fprintf(os.Stderr, "Warning: failed to store created_at: %v\n", err)
	}

	// Success!
	fmt.Println("\n✅ Vault initialized successfully!")
	fmt.Printf("   Location: %s\n", dbPath)
	fmt.Printf("   Encryption: AES-256-GCM\n")
	fmt.Printf("   Key Derivation: Argon2id (Time=%d, Memory=%dMB, Threads=%d)\n",
		argon2Params.Time, argon2Params.Memory/1024, argon2Params.Parallelism)
	fmt.Println("\n💡 Next steps:")
	fmt.Println("   • Add your first password: gpasswd add")
	fmt.Println("   • Generate a strong password: gpasswd generate")
	fmt.Println("   • List all entries: gpasswd list")
	fmt.Println("\n⚠️  IMPORTANT: Remember your master password!")
	fmt.Println("   There is NO way to recover it if you forget.")

	return nil
}
