package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/hugnun/json-server/internal"
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate a config file",
	Args:  cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		if _, err := internal.LoadConfig(args[0]); err != nil {
			return fmt.Errorf("invalid config: %w", err)
		}
		fmt.Println("Config is valid")
		return nil
	},
}

var rootCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the HTTP server from a config file",
	Args:  cobra.ExactArgs(1),
	RunE:  internal.Run,
}

func init() {
	rootCmd.Flags().IntP("port", "p", 0, "Port to listen on (overrides config)")
	rootCmd.AddCommand(validateCmd)
}

// Execute runs the root cobra command. It exits the process on error.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
