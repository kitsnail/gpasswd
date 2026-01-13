package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	// Version will be set at build time
	Version = "0.1.0-dev"
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "gpasswd",
	Short: "A secure local password manager",
	Long: `gpasswd is a command-line password manager that stores your passwords
securely on your local machine using strong encryption (AES-256-GCM + Argon2id).

All data is stored locally - no cloud, no sync, full control.`,
	Version: Version,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Global flags can be defined here
}
