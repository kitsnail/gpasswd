package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"

	"github.com/kitsnail/gpasswd/internal/crypto"
	"github.com/kitsnail/gpasswd/internal/models"
	"github.com/kitsnail/gpasswd/internal/storage"
	"github.com/kitsnail/gpasswd/pkg/config"
)

var addCmd = &cobra.Command{
	Use:   "add [name]",
	Short: "Add a new password entry",
	Long: `Add a new password entry to the vault.

You can optionally specify the entry name as an argument, or you will be
prompted for all information interactively.

You can choose to:
- Enter a password manually
- Generate a strong password automatically

Example:
  gpasswd add github
  gpasswd add "Gmail Work"
  gpasswd add`,
	RunE: runAdd,
}

var (
	addUsername  string
	addPassword  string
	addURL       string
	addNotes     string
	addCategory  string
	addTags      []string
	addGenerate  bool
	addGenLength int
)

func init() {
	rootCmd.AddCommand(addCmd)

	addCmd.Flags().StringVarP(&addUsername, "username", "u", "", "Username or email")
	addCmd.Flags().StringVarP(&addPassword, "password", "p", "", "Password (if not provided, will prompt or generate)")
	addCmd.Flags().StringVarP(&addURL, "url", "l", "", "Website URL")
	addCmd.Flags().StringVarP(&addNotes, "notes", "n", "", "Additional notes")
	addCmd.Flags().StringVarP(&addCategory, "category", "c", "general", "Category (e.g., email, social, banking)")
	addCmd.Flags().StringSliceVarP(&addTags, "tags", "t", []string{}, "Comma-separated tags")
	addCmd.Flags().BoolVarP(&addGenerate, "generate", "g", false, "Generate a strong password")
	addCmd.Flags().IntVar(&addGenLength, "gen-length", 20, "Length of generated password")
}

func runAdd(cmd *cobra.Command, args []string) error {
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

	// Create entry
	entry := &models.Entry{
		Category: addCategory,
	}

	// Get entry name
	if len(args) > 0 {
		entry.Name = args[0]
	} else {
		namePrompt := &survey.Input{
			Message: "Entry name (e.g., 'GitHub', 'Gmail Work'):",
		}
		if err := survey.AskOne(namePrompt, &entry.Name, survey.WithValidator(survey.Required)); err != nil {
			return fmt.Errorf("name prompt failed: %w", err)
		}
	}

	// Get username (interactive if not provided via flag)
	if addUsername == "" {
		usernamePrompt := &survey.Input{
			Message: "Username or email (optional):",
		}
		survey.AskOne(usernamePrompt, &entry.Username)
	} else {
		entry.Username = addUsername
	}

	// Get password
	if addPassword != "" {
		// Password provided via flag
		entry.Password = addPassword
	} else if addGenerate {
		// Generate password
		genOptions := crypto.GenerateOptions{
			UseUppercase:     cfg.PasswordGenerator.UseUppercase,
			UseLowercase:     cfg.PasswordGenerator.UseLowercase,
			UseDigits:        cfg.PasswordGenerator.UseDigits,
			UseSymbols:       cfg.PasswordGenerator.UseSymbols,
			ExcludeAmbiguous: cfg.PasswordGenerator.ExcludeAmbiguous,
		}

		length := addGenLength
		if length == 20 && cfg.PasswordGenerator.Length > 0 {
			length = cfg.PasswordGenerator.Length
		}

		generated, err := crypto.Generate(length, genOptions)
		if err != nil {
			return fmt.Errorf("failed to generate password: %w", err)
		}

		entry.Password = generated
		fmt.Printf("✓ Generated password: %s\n", generated)

		// Show strength
		strength := crypto.CheckStrength(generated)
		fmt.Printf("  Strength: %s (Score: %d/100)\n", strength.Level.String(), strength.Score)
	} else {
		// Prompt for password choice
		var choice string
		choicePrompt := &survey.Select{
			Message: "Password:",
			Options: []string{
				"Generate a strong password (recommended)",
				"Enter password manually",
			},
		}
		if err := survey.AskOne(choicePrompt, &choice); err != nil {
			return fmt.Errorf("password choice failed: %w", err)
		}

		if strings.HasPrefix(choice, "Generate") {
			// Generate password
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
			fmt.Printf("✓ Generated password: %s\n", generated)

			strength := crypto.CheckStrength(generated)
			fmt.Printf("  Strength: %s (Score: %d/100)\n", strength.Level.String(), strength.Score)
		} else {
			// Manual password entry
			passwordPrompt := &survey.Password{
				Message: "Enter password:",
			}
			if err := survey.AskOne(passwordPrompt, &entry.Password, survey.WithValidator(survey.Required)); err != nil {
				return fmt.Errorf("password prompt failed: %w", err)
			}

			// Check strength
			strength := crypto.CheckStrength(entry.Password)
			fmt.Printf("  Strength: %s (Score: %d/100)\n", strength.Level.String(), strength.Score)

			if strength.Level < crypto.Fair {
				fmt.Println("  ⚠️  Weak password. Consider using a generated password.")
			}
		}
	}

	// Get URL (interactive if not provided)
	if addURL == "" {
		urlPrompt := &survey.Input{
			Message: "Website URL (optional):",
		}
		survey.AskOne(urlPrompt, &entry.URL)
	} else {
		entry.URL = addURL
	}

	// Get category (already set from flag or default)
	if addCategory == "general" {
		categoryPrompt := &survey.Input{
			Message: "Category (optional, default: general):",
			Default: "general",
		}
		survey.AskOne(categoryPrompt, &entry.Category)
	}

	// Get tags
	if len(addTags) == 0 {
		var tagsInput string
		tagsPrompt := &survey.Input{
			Message: "Tags (comma-separated, optional):",
		}
		survey.AskOne(tagsPrompt, &tagsInput)

		if tagsInput != "" {
			for _, tag := range strings.Split(tagsInput, ",") {
				trimmed := strings.TrimSpace(tag)
				if trimmed != "" {
					entry.Tags = append(entry.Tags, trimmed)
				}
			}
		}
	} else {
		entry.Tags = addTags
	}

	// Get notes
	if addNotes == "" {
		notesPrompt := &survey.Multiline{
			Message: "Notes (optional, press Ctrl+D when done):",
		}
		survey.AskOne(notesPrompt, &entry.Notes)
	} else {
		entry.Notes = addNotes
	}

	fmt.Println("\n🔐 Encrypting and storing entry...")

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
	key, err := crypto.DeriveKey(masterPassword, salt, params)
	if err != nil {
		return fmt.Errorf("failed to derive encryption key: %w", err)
	}

	// Create entry in database
	if err := db.CreateEntry(entry, key); err != nil {
		return fmt.Errorf("failed to create entry: %w", err)
	}

	fmt.Println("\n✅ Entry added successfully!")
	fmt.Printf("   Name: %s\n", entry.Name)
	fmt.Printf("   Category: %s\n", entry.Category)
	if entry.Username != "" {
		fmt.Printf("   Username: %s\n", entry.Username)
	}
	if entry.URL != "" {
		fmt.Printf("   URL: %s\n", entry.URL)
	}
	if len(entry.Tags) > 0 {
		fmt.Printf("   Tags: %s\n", strings.Join(entry.Tags, ", "))
	}
	fmt.Printf("   ID: %s\n", entry.ID)

	fmt.Println("\n💡 Next steps:")
	fmt.Println("   • View all entries: gpasswd list")
	fmt.Println("   • Copy password: gpasswd copy " + entry.Name)
	fmt.Println("   • View entry details: gpasswd show " + entry.Name)

	return nil
}
