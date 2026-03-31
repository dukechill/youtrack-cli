package issue

import (
	"fmt"
	"strings"
	"youtrack-cli/internal/config"
	"youtrack-cli/internal/youtrack"

	"github.com/spf13/cobra"
)

var setStateCmd = &cobra.Command{
	Use:   "set-state [issue-id] [state...]",
	Short: "Update the State field on a YouTrack issue",
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil {
			fmt.Printf("Error loading configuration: %v\n", err)
			return
		}

		issueID := args[0]
		state := strings.Join(args[1:], " ")

		if err := youtrack.UpdateState(cfg, issueID, state); err != nil {
			fmt.Printf("Error updating state: %v\n", err)
			return
		}

		fmt.Printf("State updated for %s -> %s.\n", issueID, state)
	},
}

func init() {
	IssueCmd.AddCommand(setStateCmd)
}
