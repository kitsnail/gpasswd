package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"

	"github.com/kitsnail/gpasswd/internal/storage"
	"github.com/kitsnail/gpasswd/pkg/config"
)

var deleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete a password entry",
	Long: `Delete a password entry from the vault.

This operation requires confirmation (unless --force is used).
The entry will be permanently removed from the database.

Note: Master password is NOT required for deletion (only metadata is accessed).

Examples:
  gpasswd delete github
  gpasswd delete "Gmail Work"
  gpasswd delete github --force`,
	Aliases: []string{"rm", "remove"},
	Args:    cobra.ExactArgs(1),
	RunE:    runDelete,
}

var (
	deleteForce bool
)

func init() {
	rootCmd.AddCommand(deleteCmd)

	deleteCmd.Flags().BoolVarP(&deleteForce, "force", "f", false, "Skip confirmation prompt")
}

func runDelete(cmd *cobra.Command, args []string) error {
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

	// Get entries to find the one matching the name
	entries, err := db.ListEntries()
	if err != nil {
		return fmt.Errorf("failed to list entries: %w", err)
	}

	// Find entry by name (case-insensitive)
	var targetEntry *struct {
		ID       string
		Name     string
		Category string
		Username string
	}

	for _, entry := range entries {
		if strings.EqualFold(entry.Name, entryName) {
			targetEntry = &struct {
				ID       string
				Name     string
				Category string
				Username string
			}{
				ID:       entry.ID,
				Name:     entry.Name,
				Category: entry.Category,
				Username: entry.Username,
			}
			break
		}
	}

	if targetEntry == nil {
		return fmt.Errorf("entry not found: %s", entryName)
	}

	// Display entry details
	fmt.Println("\n" + strings.Repeat("─", 60))
	fmt.Printf("🗑️  Entry to delete: %s\n", targetEntry.Name)
	fmt.Println(strings.Repeat("─", 60))
	fmt.Printf("Category:    %s\n", targetEntry.Category)
	if targetEntry.Username != "" {
		fmt.Printf("Username:    %s\n", targetEntry.Username)
	}
	fmt.Printf("ID:          %s\n", targetEntry.ID)
	fmt.Println(strings.Repeat("─", 60))

	// Confirmation prompt (unless --force)
	if !deleteForce {
		fmt.Println("\n⚠️  WARNING: This operation cannot be undone!")

		var confirmed bool
		confirmPrompt := &survey.Confirm{
			Message: fmt.Sprintf("Are you sure you want to delete '%s'?", targetEntry.Name),
			Default: false,
		}

		if err := survey.AskOne(confirmPrompt, &confirmed); err != nil {
			return fmt.Errorf("confirmation prompt failed: %w", err)
		}

		if !confirmed {
			fmt.Println("\n❌ Deletion cancelled")
			return nil
		}
	}

	// Delete entry
	if err := db.DeleteEntry(targetEntry.ID); err != nil {
		return fmt.Errorf("failed to delete entry: %w", err)
	}

	fmt.Printf("\n✅ Entry '%s' deleted successfully\n", targetEntry.Name)

	return nil
}
