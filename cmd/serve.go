package cmd

import (
	"github.com/spf13/cobra"

	"github.com/hugnun/json-server/internal"
)

func init() {
	serveCmd.Flags().IntP("port", "p", 0, "Port to listen on (overrides config)")
	rootCmd.AddCommand(serveCmd)
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the HTTP server from a config file",
	Args:  cobra.ExactArgs(1),
	RunE:  internal.Run,
}
