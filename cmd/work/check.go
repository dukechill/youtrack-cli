package work

import (
	"fmt"
	"youtrack-cli/internal/config"
	"youtrack-cli/internal/youtrack"

	"github.com/spf13/cobra"
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check for issues with no work logged today",
	Long:  `Checks for YouTrack issues assigned to you that have no work logged for the current day.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil {
			fmt.Printf("Error loading configuration: %v\n", err)
			return
		}

		issuesWithoutWork, err := youtrack.CheckWork(cfg)
		if err != nil {
			fmt.Printf("Error checking work: %v\n", err)
			return
		}

		if len(issuesWithoutWork) > 0 {
			fmt.Println("You have not logged work for the following issues today:")
			for _, issue := range issuesWithoutWork {
				fmt.Println("- ", issue)
			}
		} else {
			fmt.Println("All issues have work logged for today.")
		}
	},
}

func init() {
	WorkCmd.AddCommand(checkCmd) // WorkCmd is defined in cmd/work/root.go
}