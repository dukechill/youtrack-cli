package work

import (
	"github.com/spf13/cobra"
)

var WorkCmd = &cobra.Command{
	Use:   "work",
	Short: "Manage YouTrack work items",
	Long:  `Commands for adding and checking work items in YouTrack.`,
}

func init() {
	// This is where you would add WorkCmd to the main rootCmd
	// For now, it's handled in cmd/root.go or main.go's init()
	// but for a nested command structure, it's often added here.
}