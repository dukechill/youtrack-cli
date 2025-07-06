package cmd

import (
	"fmt"
	"youtrack-cli/internal/config"
	"youtrack-cli/internal/youtrack"

	"github.com/spf13/cobra"
)

var boardCmd = &cobra.Command{
	Use:   "board",
	Short: "Manage agile boards",
	Long:  `Commands for managing YouTrack agile boards.`,
}

var boardListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available agile boards",
	Long:  `Lists all agile boards configured in your YouTrack instance.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil {
			fmt.Printf("Error loading configuration: %v\n", err)
			return
		}

		boards, err := youtrack.ListBoards(cfg)
		if err != nil {
			fmt.Printf("Error listing boards: %v\n", err)
			return
		}

		youtrack.PrintBoards(boards)
	},
}

func init() {
	rootCmd.AddCommand(boardCmd)
	boardCmd.AddCommand(boardListCmd)
}
