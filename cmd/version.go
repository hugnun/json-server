package cmd

import (
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Run: func(cmd *cobra.Command, _ []string) {
		cmd.Println("v1.0.0")
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
