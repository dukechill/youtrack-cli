package issue

import (
	"fmt"
	"strconv"
	"youtrack-cli/internal/config"
	"youtrack-cli/internal/youtrack"

	"github.com/spf13/cobra"
)

var setEstimationCmd = &cobra.Command{
	Use:   "set-estimation [issue-id] [minutes]",
	Short: "Update the Estimation field on a YouTrack issue",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil {
			fmt.Printf("Error loading configuration: %v\n", err)
			return
		}

		issueID := args[0]
		minutes, err := strconv.Atoi(args[1])
		if err != nil || minutes <= 0 {
			fmt.Printf("Error: minutes must be a positive integer, got %q\n", args[1])
			return
		}

		if err := youtrack.UpdateEstimation(cfg, issueID, minutes); err != nil {
			fmt.Printf("Error updating estimation: %v\n", err)
			return
		}

		fmt.Printf("Estimation updated for %s -> %d minutes.\n", issueID, minutes)
	},
}

func init() {
	IssueCmd.AddCommand(setEstimationCmd)
}
