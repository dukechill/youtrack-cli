package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

type Config struct {
	URL   string `yaml:"url"`
	Token string `yaml:"token"`
}

type Issue struct {
	ID      string `json:"idReadable"`
	Summary string `json:"summary"`
}

type WorkItem struct {
	Date     int64    `json:"date"`
	Duration Duration `json:"duration"`
	Author   Author   `json:"author"`
	Text     string   `json:"text"`
}

type Duration struct {
	Minutes int `json:"minutes"`
}

type Author struct {
	Login string `json:"login"`
}

var rootCmd = &cobra.Command{
	Use:   "youtrack-cli",
	Short: "A CLI for interacting with YouTrack",
}

var configCmd = &cobra.Command{
	Use:   "configure",
	Short: "Configure YouTrack URL and Token",
	Run:   configure,
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List your YouTrack issues",
	Run:   listIssues,
}

var addWorkCmd = &cobra.Command{
	Use:   "add-work [issue-id] [minutes] [description]",
	Short: "Add a work item to a YouTrack issue",
	Args:  cobra.ExactArgs(3),
	Run:   addWorkItem,
}

var checkWorkCmd = &cobra.Command{
	Use:   "check-work",
	Short: "Check for issues with no work logged today",
	Run:   checkWork,
}

func main() {
	rootCmd.AddCommand(configCmd, listCmd, addWorkCmd, checkWorkCmd)
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func configure(cmd *cobra.Command, args []string) {
	var url, token string
	fmt.Print("Enter YouTrack URL: ")
	fmt.Scanln(&url)
	fmt.Print("Enter YouTrack API Token: ")
	fmt.Scanln(&token)

	config := Config{URL: url, Token: token}
	data, err := yaml.Marshal(&config)
	if err != nil {
		fmt.Println("Error creating config file:", err)
		return
	}

	home, _ := os.UserHomeDir()
	configPath := filepath.Join(home, ".youtrack-cli.yaml")
	if err := os.WriteFile(configPath, data, 0600); err != nil {
		fmt.Println("Error saving config file:", err)
		return
	}
	fmt.Println("Configuration saved to", configPath)
}

func loadConfig() (Config, error) {
	var config Config
	home, _ := os.UserHomeDir()
	configPath := filepath.Join(home, ".youtrack-cli.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return config, fmt.Errorf("config file not found, please run 'youtrack-cli configure'")
	}
	err = yaml.Unmarshal(data, &config)
	return config, err
}

func listIssues(cmd *cobra.Command, args []string) {
	config, err := loadConfig()
	if err != nil {
		fmt.Println(err)
		return
	}

	client := &http.Client{}
	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/api/issues?fields=idReadable,summary&query=for:me", config.URL), nil)
	req.Header.Set("Authorization", "Bearer "+config.Token)
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error fetching issues:", err)
		return
	}
	defer resp.Body.Close()

	var issues []Issue
	if err := json.NewDecoder(resp.Body).Decode(&issues); err != nil {
		fmt.Println("Error decoding issues:", err)
		return
	}

	jsonOutput, _ := cmd.Flags().GetBool("json")
	if jsonOutput {
		json.NewEncoder(os.Stdout).Encode(issues)
	} else {
		for _, issue := range issues {
			fmt.Printf("任務 ID: %s, 標題: %s\n", issue.ID, issue.Summary)
		}
	}
}

func addWorkItem(cmd *cobra.Command, args []string) {
	config, err := loadConfig()
	if err != nil {
		fmt.Println(err)
		return
	}

	issueID := args[0]
	minutes := args[1]
	description := args[2]

	url := fmt.Sprintf("%s/api/issues/%s/timeTracking/workItems?fields=date,duration(minutes),author(login),text", config.URL, issueID)

	workItem := map[string]interface{}{
		"duration": map[string]string{"presentation": minutes + "m"},
		"text":     description,
	}
	jsonData, _ := json.Marshal(workItem)

	client := &http.Client{}
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", "Bearer "+config.Token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error adding work item:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		fmt.Println("Work item added successfully.")
	} else {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Error: %s - %s\n", resp.Status, string(body))
	}
}

func checkWork(cmd *cobra.Command, args []string) {
	config, err := loadConfig()
	if err != nil {
		fmt.Println(err)
		return
	}

	client := &http.Client{}
	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/api/issues?fields=idReadable,summary,updated&query=for:me", config.URL), nil)
	req.Header.Set("Authorization", "Bearer "+config.Token)
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error fetching issues:", err)
		return
	}
	defer resp.Body.Close()

	var issues []struct {
		ID      string `json:"idReadable"`
		Summary string `json:"summary"`
		Updated int64  `json:"updated"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&issues); err != nil {
		fmt.Println("Error decoding issues:", err)
		return
	}

	today := time.Now().Truncate(24 * time.Hour)
	var issuesWithoutWork []string

	for _, issue := range issues {
		workReq, _ := http.NewRequest("GET", fmt.Sprintf("%s/api/issues/%s/timeTracking/workItems?fields=date", config.URL, issue.ID), nil)
		workReq.Header.Set("Authorization", "Bearer "+config.Token)
		workReq.Header.Set("Accept", "application/json")

		workResp, err := client.Do(workReq)
		if err != nil {
			continue
		}
		defer workResp.Body.Close()

		var workItems []WorkItem
		if err := json.NewDecoder(workResp.Body).Decode(&workItems); err != nil {
			continue
		}

		hasWorkToday := false
		for _, item := range workItems {
			itemDate := time.Unix(item.Date/1000, 0)
			if itemDate.Truncate(24*time.Hour).Equal(today) {
				hasWorkToday = true
				break
			}
		}

		if !hasWorkToday {
			issuesWithoutWork = append(issuesWithoutWork, fmt.Sprintf("%s: %s", issue.ID, issue.Summary))
		}
	}

	if len(issuesWithoutWork) > 0 {
		fmt.Println("You have not logged work for the following issues today:")
		for _, issue := range issuesWithoutWork {
			fmt.Println("- ", issue)
		}
	} else {
		fmt.Println("All issues have work logged for today.")
	}
}

func init() {
	listCmd.Flags().BoolP("json", "j", false, "Output in JSON format")
}
