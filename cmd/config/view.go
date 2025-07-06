package config

import (
	"fmt"
	"youtrack-cli/internal/config"

	"github.com/spf13/cobra"
)

var viewCmd = &cobra.Command{
	Use:   "view",
	Short: "View current configuration",
	Long:  `Displays the raw content of your YouTrack CLI configuration file.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil {
			fmt.Printf("Error loading configuration: %v\n", err)
			return
		}
		config.PrintRaw(cfg)
	},
}

func init() {
	ConfigCmd.AddCommand(viewCmd) // ConfigCmd is defined in cmd/config/root.go
}