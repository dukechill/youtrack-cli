package cmd

import (
	"fmt"
	"youtrack-cli/internal/config"
	"youtrack-cli/internal/youtrack"

	"github.com/spf13/cobra"
)

var sprintCmd = &cobra.Command{
	Use:   "sprint",
	Short: "Manage sprints",
	Long:  `Commands for managing YouTrack sprints.`,
}

var sprintCurrentCmd = &cobra.Command{
	Use:   "current",
	Short: "Show the current sprint for a board",
	Long:  `Shows the sprint selected by YouTrack's current-sprint marker or time window logic.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil {
			fmt.Printf("Error loading configuration: %v\n", err)
			return
		}

		boardName, _ := cmd.Flags().GetString("board")
		if boardName == "" {
			boardName = cfg.BoardName
		}

		if boardName == "" {
			fmt.Println("Error: Board name not specified. Please use the --board flag or set a default board using 'youtrack-cli config set board [board_name]'")
			return
		}

		cfg.BoardName = boardName

		sprintName, err := youtrack.DetermineCurrentSprint(cfg)
		if err != nil {
			fmt.Printf("Error determining current sprint for board '%s': %v\n", boardName, err)
			return
		}

		fmt.Printf("Current sprint for board '%s': %s\n", boardName, sprintName)
	},
}

var sprintListCmd = &cobra.Command{
	Use:   "list",
	Short: "List sprints for a specific board",
	Long:  `Lists all sprints for a specified YouTrack board. Uses the default board from config if not specified.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil {
			fmt.Printf("Error loading configuration: %v\n", err)
			return
		}

		boardName, _ := cmd.Flags().GetString("board")
		if boardName == "" {
			boardName = cfg.BoardName
		}

		if boardName == "" {
			fmt.Println("Error: Board name not specified. Please use the --board flag or set a default board using 'youtrack-cli config set board [board_name]'")
			return
		}

		sprints, err := youtrack.ListSprints(cfg, boardName)
		if err != nil {
			fmt.Printf("Error listing sprints for board '%s': %v\n", boardName, err)
			return
		}

		youtrack.PrintSprints(boardName, sprints)
	},
}

func init() {
	// rootCmd.AddCommand(sprintCmd) // REMOVED: Added in cmd/root.go
	sprintCmd.AddCommand(sprintCurrentCmd)
	sprintCmd.AddCommand(sprintListCmd)

	// Define flags for the sprint list command
	sprintCurrentCmd.Flags().StringP("board", "b", "", "Board name to resolve the current sprint from")
	sprintListCmd.Flags().StringP("board", "b", "", "Board name to list sprints from")
}
