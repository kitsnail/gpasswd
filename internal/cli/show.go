package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"

	"github.com/kitsnail/gpasswd/internal/crypto"
	"github.com/kitsnail/gpasswd/internal/storage"
	"github.com/kitsnail/gpasswd/pkg/config"
)

var showCmd = &cobra.Command{
	Use:   "show <name>",
	Short: "Show a password entry",
	Long: `Show details of a password entry including the password.

The master password is required to decrypt the entry.

By default, the password is hidden. Use --reveal to display it.

Examples:
  gpasswd show github
  gpasswd show "Gmail Work" --reveal`,
	Aliases: []string{"get", "view"},
	Args:    cobra.ExactArgs(1),
	RunE:    runShow,
}

var (
	showReveal bool
)

func init() {
	rootCmd.AddCommand(showCmd)

	showCmd.Flags().BoolVarP(&showReveal, "reveal", "r", false, "Reveal password in output")
}

func runShow(cmd *cobra.Command, args []string) error {
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

	// Display entry details
	fmt.Println("\n" + strings.Repeat("─", 60))
	fmt.Printf("📝 Entry: %s\n", entry.Name)
	fmt.Println(strings.Repeat("─", 60))

	fmt.Printf("Category:    %s\n", entry.Category)

	if entry.Username != "" {
		fmt.Printf("Username:    %s\n", entry.Username)
	}

	// Password display
	if showReveal {
		fmt.Printf("Password:    %s\n", entry.Password)

		// Show strength
		strength := crypto.CheckStrength(entry.Password)
		fmt.Printf("Strength:    %s (Score: %d/100)\n", strength.Level.String(), strength.Score)
	} else {
		fmt.Printf("Password:    %s\n", strings.Repeat("•", 12))
		fmt.Println("             (use --reveal to show)")
	}

	if entry.URL != "" {
		fmt.Printf("URL:         %s\n", entry.URL)
	}

	if len(entry.Tags) > 0 {
		fmt.Printf("Tags:        %s\n", strings.Join(entry.Tags, ", "))
	}

	if entry.Notes != "" {
		fmt.Println("\nNotes:")
		// Indent notes
		for _, line := range strings.Split(entry.Notes, "\n") {
			fmt.Printf("  %s\n", line)
		}
	}

	fmt.Println("\nTimestamps:")
	dateFormat := "2006-01-02 15:04:05"
	if cfg.Display.DateFormat != "" {
		dateFormat = cfg.Display.DateFormat
	}
	fmt.Printf("  Created:   %s\n", entry.CreatedAt.Format(dateFormat))
	fmt.Printf("  Updated:   %s\n", entry.UpdatedAt.Format(dateFormat))

	fmt.Printf("\nID:          %s\n", entry.ID)
	fmt.Println(strings.Repeat("─", 60))

	// Helpful actions
	fmt.Println("\n💡 Actions:")
	fmt.Printf("   • Copy password:  gpasswd copy %s\n", entry.Name)
	fmt.Printf("   • Edit entry:     gpasswd edit %s\n", entry.Name)
	fmt.Printf("   • Delete entry:   gpasswd delete %s\n", entry.Name)

	return nil
}
