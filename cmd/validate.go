package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/hugnun/json-server/internal"
)

func init() {
	rootCmd.AddCommand(validateCmd)
}

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
