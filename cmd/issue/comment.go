package issue

import (
	"fmt"
	"strings"
	"youtrack-cli/internal/config"
	"youtrack-cli/internal/youtrack"

	"github.com/spf13/cobra"
)

var commentCmd = &cobra.Command{
	Use:   "comment [issue-id] [text...]",
	Short: "Add a comment to a YouTrack issue",
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil {
			fmt.Printf("Error loading configuration: %v\n", err)
			return
		}

		issueID := args[0]
		commentText := strings.Join(args[1:], " ")

		if err := youtrack.AddComment(cfg, issueID, commentText); err != nil {
			fmt.Printf("Error adding comment: %v\n", err)
			return
		}

		fmt.Printf("Comment added to %s.\n", issueID)
	},
}

func init() {
	IssueCmd.AddCommand(commentCmd)
}
