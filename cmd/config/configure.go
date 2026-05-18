package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"youtrack-cli/internal/config"

	"github.com/spf13/cobra"
)

var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Interactively configure YouTrack CLI settings",
	Long:  `This command guides you through setting up your YouTrack CLI by prompting for necessary credentials and default values.`,
	Run: func(cmd *cobra.Command, args []string) {
		reader := bufio.NewReader(os.Stdin)

		fmt.Print("Enter YouTrack Base URL (e.g., https://your-instance.youtrack.cloud): ")
		baseURL, _ := reader.ReadString('\n')
		baseURL = strings.TrimSpace(baseURL)
		if err := config.SetValue("url", baseURL); err != nil {
			fmt.Printf("Error setting YouTrack Base URL: %v\n", err)
			return
		}

		fmt.Print("Enter YouTrack Permanent Token: ")
		token, _ := reader.ReadString('\n')
		token = strings.TrimSpace(token)
		if err := config.SetValue("token", token); err != nil {
			fmt.Printf("Error setting YouTrack Permanent Token: %v\n", err)
			return
		}

		fmt.Print("Enter default Board name (optional, press Enter to skip): ")
		boardName, _ := reader.ReadString('\n')
		boardName = strings.TrimSpace(boardName)
		if boardName != "" {
			if err := config.SetValue("board", boardName); err != nil {
				fmt.Printf("Error setting default Board name: %v\n", err)
				return
			}
		}

		fmt.Print("Enter default Sprint Name (optional, press Enter to skip): ")
		sprintName, _ := reader.ReadString('\n')
		sprintName = strings.TrimSpace(sprintName)
		if sprintName != "" {
			if err := config.SetValue("sprint", sprintName); err != nil {
				fmt.Printf("Error setting default Sprint Name: %v\n", err)
				return
			}
		}

		fmt.Println("\nConfiguration complete! You can now use 'youtrack-cli config show' to view your settings.")
	},
}

func init() {
	ConfigCmd.AddCommand(configureCmd)
}
