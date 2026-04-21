package issue

import "github.com/spf13/cobra"

var IssueCmd = &cobra.Command{
	Use:   "issue",
	Short: "Manage YouTrack issues",
	Long:  `Commands for updating comments, fields, and sprint membership on YouTrack issues.`,
}
