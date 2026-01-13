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

var editCmd = &cobra.Command{
	Use:   "edit <name>",
	Short: "Edit an existing password entry",
	Long: `Edit an existing password entry in the vault.

You can update any field: username, password, URL, notes, category, or tags.
Fields not specified will remain unchanged.

The master password is required to decrypt and re-encrypt the entry.

Examples:
  gpasswd edit github
  gpasswd edit github --username newuser@example.com
  gpasswd edit github --password newpass123
  gpasswd edit github --generate`,
	Aliases: []string{"update", "modify"},
	Args:    cobra.ExactArgs(1),
	RunE:    runEdit,
}

var (
	editUsername string
	editPassword string
	editURL      string
	editNotes    string
	editCategory string
	editTags     []string
	editGenerate bool
	editGenLen   int
	editSetTags  bool
)

func init() {
	rootCmd.AddCommand(editCmd)

	editCmd.Flags().StringVarP(&editUsername, "username", "u", "", "New username")
	editCmd.Flags().StringVarP(&editPassword, "password", "p", "", "New password")
	editCmd.Flags().StringVarP(&editURL, "url", "l", "", "New URL")
	editCmd.Flags().StringVarP(&editNotes, "notes", "n", "", "New notes")
	editCmd.Flags().StringVarP(&editCategory, "category", "c", "", "New category")
	editCmd.Flags().StringSliceVarP(&editTags, "tags", "t", []string{}, "New tags (comma-separated)")
	editCmd.Flags().BoolVarP(&editGenerate, "generate", "g", false, "Generate new password")
	editCmd.Flags().IntVar(&editGenLen, "gen-length", 20, "Length of generated password")
	editCmd.Flags().BoolVar(&editSetTags, "set-tags", false, "Replace tags (otherwise keep existing)")
}

func runEdit(cmd *cobra.Command, args []string) error {
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

	// Get existing entry
	entry, err := db.GetEntryByName(entryName, key)
	if err != nil {
		return fmt.Errorf("failed to get entry: %w", err)
	}

	fmt.Printf("\n📝 Editing entry: %s\n", entry.Name)

	// Check if any flags provided
	hasFlags := cmd.Flags().Changed("username") ||
		cmd.Flags().Changed("password") ||
		cmd.Flags().Changed("url") ||
		cmd.Flags().Changed("notes") ||
		cmd.Flags().Changed("category") ||
		cmd.Flags().Changed("tags") ||
		editGenerate

	if hasFlags {
		// Update from flags
		if cmd.Flags().Changed("username") {
			entry.Username = editUsername
		}

		if editGenerate {
			// Generate new password
			genOptions := crypto.GenerateOptions{
				UseUppercase:     true,
				UseLowercase:     true,
				UseDigits:        true,
				UseSymbols:       true,
				ExcludeAmbiguous: cfg.PasswordGenerator.ExcludeAmbiguous,
			}

			generated, err := crypto.Generate(editGenLen, genOptions)
			if err != nil {
				return fmt.Errorf("failed to generate password: %w", err)
			}

			entry.Password = generated
			fmt.Printf("✓ Generated new password: %s\n", generated)

			strength := crypto.CheckStrength(generated)
			fmt.Printf("  Strength: %s (Score: %d/100)\n", strength.Level.String(), strength.Score)
		} else if cmd.Flags().Changed("password") {
			entry.Password = editPassword
		}

		if cmd.Flags().Changed("url") {
			entry.URL = editURL
		}

		if cmd.Flags().Changed("notes") {
			entry.Notes = editNotes
		}

		if cmd.Flags().Changed("category") {
			entry.Category = editCategory
		}

		if editSetTags || cmd.Flags().Changed("tags") {
			entry.Tags = editTags
		}
	} else {
		// Interactive editing
		fmt.Println("\nLeave blank to keep current value.\n")

		// Username
		var newUsername string
		usernamePrompt := &survey.Input{
			Message: "Username:",
			Default: entry.Username,
		}
		if err := survey.AskOne(usernamePrompt, &newUsername); err == nil && newUsername != "" {
			entry.Username = newUsername
		}

		// Password choice
		var passwordChoice string
		passwordPrompt := &survey.Select{
			Message: "Password:",
			Options: []string{
				"Keep current password",
				"Generate new password",
				"Enter new password manually",
			},
		}
		if err := survey.AskOne(passwordPrompt, &passwordChoice); err != nil {
			return fmt.Errorf("password choice failed: %w", err)
		}

		if strings.HasPrefix(passwordChoice, "Generate") {
			genOptions := crypto.GenerateOptions{
				UseUppercase:     true,
				UseLowercase:     true,
				UseDigits:        true,
				UseSymbols:       true,
				ExcludeAmbiguous: cfg.PasswordGenerator.ExcludeAmbiguous,
			}

			generated, err := crypto.Generate(20, genOptions)
			if err != nil {
				return fmt.Errorf("failed to generate password: %w", err)
			}

			entry.Password = generated
			fmt.Printf("✓ Generated new password: %s\n", generated)

			strength := crypto.CheckStrength(generated)
			fmt.Printf("  Strength: %s (Score: %d/100)\n", strength.Level.String(), strength.Score)
		} else if strings.HasPrefix(passwordChoice, "Enter") {
			var newPassword string
			newPassPrompt := &survey.Password{
				Message: "New password:",
			}
			if err := survey.AskOne(newPassPrompt, &newPassword, survey.WithValidator(survey.Required)); err != nil {
				return fmt.Errorf("password prompt failed: %w", err)
			}

			entry.Password = newPassword

			strength := crypto.CheckStrength(newPassword)
			fmt.Printf("  Strength: %s (Score: %d/100)\n", strength.Level.String(), strength.Score)
		}

		// URL
		var newURL string
		urlPrompt := &survey.Input{
			Message: "URL:",
			Default: entry.URL,
		}
		if err := survey.AskOne(urlPrompt, &newURL); err == nil && newURL != "" {
			entry.URL = newURL
		}

		// Category
		var newCategory string
		categoryPrompt := &survey.Input{
			Message: "Category:",
			Default: entry.Category,
		}
		if err := survey.AskOne(categoryPrompt, &newCategory); err == nil && newCategory != "" {
			entry.Category = newCategory
		}

		// Tags
		var tagsInput string
		currentTags := strings.Join(entry.Tags, ", ")
		tagsPrompt := &survey.Input{
			Message: "Tags (comma-separated):",
			Default: currentTags,
		}
		if err := survey.AskOne(tagsPrompt, &tagsInput); err == nil && tagsInput != "" {
			entry.Tags = []string{}
			for _, tag := range strings.Split(tagsInput, ",") {
				trimmed := strings.TrimSpace(tag)
				if trimmed != "" {
					entry.Tags = append(entry.Tags, trimmed)
				}
			}
		}

		// Notes
		var newNotes string
		notesPrompt := &survey.Multiline{
			Message: "Notes (Ctrl+D when done):",
			Default: entry.Notes,
		}
		if err := survey.AskOne(notesPrompt, &newNotes); err == nil && newNotes != "" {
			entry.Notes = newNotes
		}
	}

	// Update entry in database
	fmt.Println("\n🔐 Encrypting and updating entry...")
	if err := db.UpdateEntry(entry, key); err != nil {
		return fmt.Errorf("failed to update entry: %w", err)
	}

	fmt.Println("\n✅ Entry updated successfully!")
	fmt.Printf("   Name: %s\n", entry.Name)
	fmt.Printf("   Category: %s\n", entry.Category)
	if entry.Username != "" {
		fmt.Printf("   Username: %s\n", entry.Username)
	}
	if entry.URL != "" {
		fmt.Printf("   URL: %s\n", entry.URL)
	}

	return nil
}
