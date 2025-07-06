package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"youtrack-cli/cmd/config" // Import config package
	"youtrack-cli/cmd/work"   // Import work package
)

var rootCmd = &cobra.Command{
	Use:   "youtrack-cli",
	Short: "A CLI for interacting with YouTrack",
	Long: `youtrack-cli is a command-line interface tool designed to interact with
YouTrack, allowing you to manage issues, sprints, and configurations directly from your terminal.`,
	// Uncomment the following line if your root command has its own action
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	// Add top-level commands
	rootCmd.AddCommand(config.ConfigCmd) // Add the config root command
	rootCmd.AddCommand(boardCmd)
	rootCmd.AddCommand(sprintCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(work.WorkCmd) // Add the work root command

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.youtrack-cli.yaml)")

	// Cobra also supports local flags, which will only run when this command
	// is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
