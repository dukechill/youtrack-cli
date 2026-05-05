package issue

import (
	"fmt"
	"youtrack-cli/internal/config"
	"youtrack-cli/internal/youtrack"

	"github.com/spf13/cobra"
)

var inspectCmd = &cobra.Command{
	Use:   "inspect [issue-id]",
	Short: "Inspect one issue with recent comments, work items, sprints, and links",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil {
			fmt.Printf("Error loading configuration: %v\n", err)
			return
		}

		issueID := args[0]
		inspect, err := youtrack.InspectIssue(cfg, issueID)
		if err != nil {
			fmt.Printf("Error inspecting issue %s: %v\n", issueID, err)
			return
		}

		youtrack.PrintIssueInspect(inspect)
	},
}

func init() {
	IssueCmd.AddCommand(inspectCmd)
}
