package issue

import (
	"fmt"
	"youtrack-cli/internal/config"
	"youtrack-cli/internal/youtrack"

	"github.com/spf13/cobra"
)

var (
	dailySyncMinutes int
	dailySyncComment string
	dailySyncState   string
	dailySyncWork    string
)

var dailySyncCmd = &cobra.Command{
	Use:   "daily-sync [issue-id]",
	Short: "Log work and apply a concise end-of-day issue update",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil {
			fmt.Printf("Error loading configuration: %v\n", err)
			return
		}

		issueID := args[0]
		actions, err := youtrack.DailySync(cfg, issueID, youtrack.DailySyncOptions{
			Minutes:         dailySyncMinutes,
			Comment:         dailySyncComment,
			State:           dailySyncState,
			WorkDescription: dailySyncWork,
		})
		if err != nil {
			fmt.Printf("Error running daily sync: %v\n", err)
			return
		}

		fmt.Printf("Daily sync completed for %s.\n", issueID)
		for _, action := range actions {
			fmt.Printf("- %s\n", action)
		}
	},
}

func init() {
	dailySyncCmd.Flags().IntVar(&dailySyncMinutes, "minutes", 0, "Minutes to log as a work item")
	dailySyncCmd.Flags().StringVar(&dailySyncComment, "comment", "", "Progress note to add as an issue comment")
	dailySyncCmd.Flags().StringVar(&dailySyncState, "state", "", "Target State value to apply")
	dailySyncCmd.Flags().StringVar(&dailySyncWork, "work", "", "Optional work item description; defaults to --comment")

	IssueCmd.AddCommand(dailySyncCmd)
}
