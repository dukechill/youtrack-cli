package config

import (
	"fmt"
	"youtrack-cli/internal/config"

	"github.com/spf13/cobra"
)

var showCmd = &cobra.Command{
	Use:	"show",
	Short:	"Show current configuration (hiding sensitive parts)",
	Long:	`Displays your current YouTrack CLI configuration, masking sensitive information like the API token.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil {
			fmt.Printf("Error loading configuration: %v\n", err)
			return
		}
		config.PrintMasked(cfg)
	},
}

func init() {
	ConfigCmd.AddCommand(showCmd) // ConfigCmd is defined in cmd/config/root.go
}