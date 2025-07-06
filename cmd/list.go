package cmd

import (
	"fmt"
	"youtrack-cli/internal/config"
	"youtrack-cli/internal/youtrack"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List your YouTrack issues",
	Long:  `List YouTrack issues based on various filters like sprint, assignee, etc.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil {
			fmt.Printf("Error loading configuration: %v\n", err)
			return
		}

		sprintName, _ := cmd.Flags().GetString("sprint")
		assigneeName, _ := cmd.Flags().GetString("assignee")

		// Determine sprint name (flag > default config > latest sprint)
		determinedSprint, err := youtrack.DetermineSprint(cfg, sprintName)
		if err != nil {
			fmt.Printf("Warning: Could not determine sprint: %v. Listing issues without sprint filter.\n", err)
			determinedSprint = "" // Proceed without sprint filter if determination fails
		}

		// Build YouTrack query string
		query := youtrack.BuildQuery(determinedSprint, assigneeName, cfg.BoardName)

		// Fetch issues from YouTrack API
		issues, err := youtrack.FetchIssues(cfg, query)
		if err != nil {
			fmt.Printf("Error fetching issues: %v\n", err)
			return
		}

		// Print issues in a formatted table
		youtrack.PrintIssues(issues)
	},
}

func init() {
	// rootCmd.AddCommand(listCmd) // REMOVED: Added in cmd/root.go

	// Define flags for the list command
	listCmd.Flags().StringP("sprint", "s", "", "Specify the sprint to list issues from")
	listCmd.Flags().StringP("assignee", "a", "", "Specify the assignee to list issues for (e.g., 'me', 'unassigned', or a username)")
}