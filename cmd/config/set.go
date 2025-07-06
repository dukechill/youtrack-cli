package config

import (
	"fmt"
	"youtrack-cli/internal/config"

	"github.com/spf13/cobra"
)

var setCmd = &cobra.Command{
	Use:   "set [key] [value]",
	Short: "Set a configuration value (e.g., sprint, board)",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]
		value := args[1]

		err := config.SetValue(key, value)
		if err != nil {
			fmt.Printf("Error setting configuration value: %v\n", err)
			return
		}
		fmt.Printf("Configuration updated: %s = %s\n", key, value)
	},
}

func init() {
	ConfigCmd.AddCommand(setCmd) // ConfigCmd is defined in cmd/config/root.go
}