package cli

import (
	"fmt"
	"os"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"

	"github.com/kitsnail/gpasswd/internal/clipboard"
	"github.com/kitsnail/gpasswd/internal/crypto"
	"github.com/kitsnail/gpasswd/internal/storage"
	"github.com/kitsnail/gpasswd/pkg/config"
)

var copyCmd = &cobra.Command{
	Use:   "copy <name>",
	Short: "Copy a password to clipboard",
	Long: `Copy a password entry to the system clipboard.

The password will be automatically cleared from the clipboard after a timeout
(default: 30 seconds, configurable in config.yaml).

The master password is required to decrypt the entry.

Examples:
  gpasswd copy github
  gpasswd copy "Gmail Work"`,
	Aliases: []string{"cp"},
	Args:    cobra.ExactArgs(1),
	RunE:    runCopy,
}

var (
	copyNoClear bool
	copyTimeout int
)

func init() {
	rootCmd.AddCommand(copyCmd)

	copyCmd.Flags().BoolVar(&copyNoClear, "no-clear", false, "Don't auto-clear clipboard")
	copyCmd.Flags().IntVarP(&copyTimeout, "timeout", "t", 0, "Clipboard clear timeout in seconds (0 = use config default)")
}

func runCopy(cmd *cobra.Command, args []string) error {
	entryName := args[0]

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Determine database path
	dbPath := cfg.Database.Path
	if dbPath == "" {
		dbPath = config.GetVaultPath()
	}

	// Check if vault exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return fmt.Errorf("vault not initialized. Run 'gpasswd init' first")
	}

	// Open database
	db, err := storage.InitDB(dbPath)
	if err != nil {
		return fmt.Errorf("failed to open vault: %w", err)
	}
	defer db.Close()

	// Prompt for master password
	var masterPassword string
	masterPrompt := &survey.Password{
		Message: "Master password:",
	}
	if err := survey.AskOne(masterPrompt, &masterPassword, survey.WithValidator(survey.Required)); err != nil {
		return fmt.Errorf("master password prompt failed: %w", err)
	}

	// Get salt and params
	salt, err := db.GetSalt()
	if err != nil {
		return fmt.Errorf("failed to get salt: %w", err)
	}

	params, err := db.GetArgon2Params()
	if err != nil {
		return fmt.Errorf("failed to get Argon2 parameters: %w", err)
	}

	// Derive encryption key
	fmt.Println("🔓 Unlocking vault...")
	key, err := crypto.DeriveKey(masterPassword, salt, params)
	if err != nil {
		return fmt.Errorf("failed to derive encryption key: %w", err)
	}

	// Get entry by name
	entry, err := db.GetEntryByName(entryName, key)
	if err != nil {
		return fmt.Errorf("failed to get entry: %w", err)
	}

	// Copy password to clipboard
	if err := clipboard.Copy(entry.Password); err != nil {
		return fmt.Errorf("failed to copy to clipboard: %w", err)
	}

	fmt.Printf("✅ Password for '%s' copied to clipboard\n", entry.Name)

	// Auto-clear clipboard after timeout
	if !copyNoClear {
		timeout := copyTimeout
		if timeout == 0 {
			timeout = cfg.Clipboard.ClearTimeout
			if timeout == 0 {
				timeout = 30 // Default 30 seconds
			}
		}

		fmt.Printf("⏱️  Clipboard will be cleared in %d seconds\n", timeout)
		fmt.Println("   (Press Ctrl+C to cancel and keep in clipboard)")

		done, err := clipboard.CopyWithAutoClear(entry.Password, time.Duration(timeout)*time.Second)
		if err != nil {
			return fmt.Errorf("failed to setup auto-clear: %w", err)
		}

		// Wait for auto-clear or interrupt
		<-done
		fmt.Println("\n🧹 Clipboard cleared")
	} else {
		fmt.Println("⚠️  Clipboard will NOT be auto-cleared (--no-clear flag)")
	}

	return nil
}
