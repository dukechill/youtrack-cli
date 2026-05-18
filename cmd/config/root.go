package config

import (
	"github.com/spf13/cobra"
)

var ConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage YouTrack CLI configuration",
	Long:  `Commands for managing your YouTrack CLI configuration, including setting values and viewing current settings.`,
}
