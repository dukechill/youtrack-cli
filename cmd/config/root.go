package config

import (
	"github.com/spf13/cobra"
)

var ConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage YouTrack CLI configuration",
	Long:  `Commands for managing your YouTrack CLI configuration, including setting values and viewing current settings.`,
}

func init() {
	// This is where you would add ConfigCmd to the main rootCmd
	// For now, it's handled in cmd/root.go or main.go's init()
	// but for a nested command structure, it's often added here.
	// However, per your structure, it's added in cmd/root.go's init()
	// or main.go's main() function.
}