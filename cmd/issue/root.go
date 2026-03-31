package issue

import "github.com/spf13/cobra"

var IssueCmd = &cobra.Command{
	Use:   "issue",
	Short: "Manage YouTrack issues",
	Long:  `Commands for updating comments and fields on YouTrack issues.`,
}
