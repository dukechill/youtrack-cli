package issue

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"youtrack-cli/internal/config"
	"youtrack-cli/internal/youtrack"
)

var (
	issueSprintBoardFlag string
	issueSprintNameFlag  string
)

var issueSprintCmd = &cobra.Command{
	Use:   "sprint",
	Short: "Inspect or update issue sprint membership",
	Long:  `Commands for checking which sprints an issue belongs to and moving an issue to a sprint.`,
}

var issueSprintShowCmd = &cobra.Command{
	Use:   "show [issue-id]",
	Short: "Show the sprints assigned to an issue",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil {
			fmt.Printf("Error loading configuration: %v\n", err)
			return
		}

		issueID := args[0]
		sprints, err := youtrack.ListIssueSprints(cfg, issueID)
		if err != nil {
			fmt.Printf("Error listing sprints for %s: %v\n", issueID, err)
			return
		}

		youtrack.PrintIssueSprints(issueID, sprints)
	},
}

var issueSprintSetCmd = &cobra.Command{
	Use:   "set [issue-id...]",
	Short: "Assign one or more issues to a sprint",
	Long:  `Assigns one or more issues to an explicit sprint, or to the current sprint on the configured board when --sprint is omitted.`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil {
			fmt.Printf("Error loading configuration: %v\n", err)
			return
		}

		boardName := strings.TrimSpace(issueSprintBoardFlag)
		if boardName == "" {
			boardName = cfg.BoardName
		}
		if boardName == "" {
			fmt.Println("Error: Board name not specified. Please use the --board flag or set a default board using 'youtrack-cli config set board [board_name]'")
			return
		}

		sprintName := strings.TrimSpace(issueSprintNameFlag)
		if sprintName == "" {
			cfgForLookup := cfg
			cfgForLookup.BoardName = boardName

			sprintName, err = youtrack.DetermineCurrentSprint(cfgForLookup)
			if err != nil {
				fmt.Printf("Error determining current sprint for board '%s': %v\n", boardName, err)
				return
			}
		}

		if err := youtrack.SetIssuesSprint(cfg, args, boardName, sprintName); err != nil {
			fmt.Printf("Error updating sprint for %s: %v\n", strings.Join(args, ", "), err)
			return
		}

		fmt.Printf("Sprint updated for %s -> %s (%s).\n", strings.Join(args, ", "), sprintName, boardName)
	},
}

func init() {
	IssueCmd.AddCommand(issueSprintCmd)
	issueSprintCmd.AddCommand(issueSprintShowCmd)
	issueSprintCmd.AddCommand(issueSprintSetCmd)

	issueSprintSetCmd.Flags().StringVar(&issueSprintBoardFlag, "board", "", "Board name to resolve the current sprint from")
	issueSprintSetCmd.Flags().StringVar(&issueSprintNameFlag, "sprint", "", "Sprint name to assign. Defaults to the current sprint on the selected board.")
}
