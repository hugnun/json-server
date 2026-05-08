package cmd

import (
	"fmt"
	"os"

	"github.com/hugnun/json-server/internal"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate a config file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := os.ReadFile(args[0])
		if err != nil {
			return fmt.Errorf("failed to read config: %w", err)
		}
		var config internal.Config
		if err := yaml.Unmarshal(data, &config); err != nil {
			return fmt.Errorf("invalid YAML: %w", err)
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

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}