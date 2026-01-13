package cli

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/kitsnail/gpasswd/internal/models"
	"github.com/kitsnail/gpasswd/internal/storage"
	"github.com/kitsnail/gpasswd/pkg/config"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all password entries",
	Long: `List all password entries in the vault.

Displays entry metadata without decrypting passwords (no master password required).
Shows: Name, Category, Username, and creation date.

You can filter by category using the --category flag.

Examples:
  gpasswd list
  gpasswd list --category work
  gpasswd list -c email`,
	Aliases: []string{"ls"},
	RunE:    runList,
}

var (
	listCategory string
	listVerbose  bool
)

func init() {
	rootCmd.AddCommand(listCmd)

	listCmd.Flags().StringVarP(&listCategory, "category", "c", "", "Filter by category")
	listCmd.Flags().BoolVarP(&listVerbose, "verbose", "v", false, "Show additional details")
}

func runList(cmd *cobra.Command, args []string) error {
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

	// Get entries
	var entries []*models.Entry
	if listCategory != "" {
		entries, err = db.ListEntriesByCategory(listCategory)
		if err != nil {
			return fmt.Errorf("failed to list entries: %w", err)
		}
	} else {
		entries, err = db.ListEntries()
		if err != nil {
			return fmt.Errorf("failed to list entries: %w", err)
		}
	}

	// Check if empty
	if len(entries) == 0 {
		if listCategory != "" {
			fmt.Printf("No entries found in category '%s'\n", listCategory)
		} else {
			fmt.Println("No entries in vault")
			fmt.Println("\n💡 Add your first entry:")
			fmt.Println("   gpasswd add")
		}
		return nil
	}

	// Display header
	if listCategory != "" {
		fmt.Printf("📋 Entries in category '%s': %d\n\n", listCategory, len(entries))
	} else {
		fmt.Printf("📋 Total entries: %d\n\n", len(entries))
	}

	// Create table writer
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)

	// Print header
	if listVerbose {
		fmt.Fprintln(w, "NAME\tCATEGORY\tUSERNAME\tCREATED\tUPDATED\tID")
		fmt.Fprintln(w, "----\t--------\t--------\t-------\t-------\t--")
	} else {
		fmt.Fprintln(w, "NAME\tCATEGORY\tUSERNAME\tCREATED")
		fmt.Fprintln(w, "----\t--------\t--------\t-------")
	}

	// Print entries
	dateFormat := "2006-01-02 15:04"
	if cfg.Display.DateFormat != "" {
		dateFormat = cfg.Display.DateFormat
	}

	for _, entry := range entries {
		name := entry.Name
		category := entry.Category
		username := entry.Username
		if username == "" {
			username = "-"
		}

		created := entry.CreatedAt.Format(dateFormat)

		if listVerbose {
			updated := entry.UpdatedAt.Format(dateFormat)
			id := entry.ID
			if len(id) > 8 {
				id = id[:8] + "..."
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
				name, category, username, created, updated, id)
		} else {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
				name, category, username, created)
		}
	}

	w.Flush()

	// Summary footer
	fmt.Println()
	if !listVerbose {
		fmt.Println("💡 Tip: Use --verbose (-v) to show more details")
	}
	fmt.Println("💡 Use 'gpasswd copy <name>' to copy a password")

	return nil
}
