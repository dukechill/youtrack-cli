package work

import (
	"fmt"
	"youtrack-cli/internal/config"
	"youtrack-cli/internal/youtrack"

	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add [issue-id] [minutes] [description]",
	Short: "Add a work item to a YouTrack issue",
	Args:  cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil {
			fmt.Printf("Error loading configuration: %v\n", err)
			return
		}

		issueID := args[0]
		minutes := args[1]
		description := args[2]

		err = youtrack.AddWorkItem(cfg, issueID, minutes, description)
		if err != nil {
			fmt.Printf("Error adding work item: %v\n", err)
			return
		}
		fmt.Println("Work item added successfully.")
	},
}

func init() {
	WorkCmd.AddCommand(addCmd) // WorkCmd is defined in cmd/work/root.go
}