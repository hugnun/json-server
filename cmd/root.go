package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "json-server",
	Short: "A CLI tool for serving JSON data over HTTP",
	Long:  `It serves a JSON file over HTTP, with optional CORS and basic auth support.`,
}

// Execute runs the root cobra command. It exits the process on error.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
