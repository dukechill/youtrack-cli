package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

type Config struct {
	URL           string `yaml:"url"`
	Token         string `yaml:"token"`
	DefaultSprint string `yaml:"default_sprint,omitempty"`
	BoardName     string `yaml:"board_name,omitempty"`
}

type Issue struct {
	ID           string        `json:"idReadable"`
	Summary      string        `json:"summary"`
	CustomFields []CustomField `json:"customFields"`
	Sprints      []Sprint      `json:"sprints,omitempty"`
}

type CustomField struct {
	Name  string      `json:"name"`
	Value interface{} `json:"value"`
}

type AgileBoard struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Sprint struct {
	ID   string `json:"id"`
	Name string `json:"name"`
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
	Use:   "config",
	Short: "Configure YouTrack settings",
}

var configSetCmd = &cobra.Command{
	Use:   "set [key] [value]",
	Short: "Set a configuration value (e.g., sprint, board)",
	Args:  cobra.ExactArgs(2),
	Run:   setConfigValue,
}

var configViewCmd = &cobra.Command{
	Use:   "view",
	Short: "View current configuration",
	Run:   viewConfig,
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration (hiding sensitive parts)",
	Run:   showConfig,
}

var boardCmd = &cobra.Command{
	Use:   "board",
	Short: "Manage agile boards",
}

var boardListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available agile boards",
	Run:   listBoards,
}

var configListBoardsCmd = &cobra.Command{
	Use:   "list-boards",
	Short: "List available agile boards",
	Run:   listBoards,
}

var sprintCmd = &cobra.Command{
	Use:   "sprint",
	Short: "Manage sprints",
}

var sprintListCmd = &cobra.Command{
	Use:   "list",
	Short: "List sprints for a specific board",
	Run:   listSprints,
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
	configCmd.AddCommand(configSetCmd, configViewCmd, configShowCmd, configListBoardsCmd)
	boardCmd.AddCommand(boardListCmd)
	sprintCmd.AddCommand(sprintListCmd)
	rootCmd.AddCommand(configCmd, boardCmd, sprintCmd, listCmd, addWorkCmd, checkWorkCmd)
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

func setConfigValue(cmd *cobra.Command, args []string) {
	config, err := loadConfig()
	if err != nil {
		// If config doesn't exist, create a new one
		if os.IsNotExist(err) {
			config = Config{}
		} else {
			fmt.Println("Error loading config:", err)
			return
		}
	}

	key := args[0]
	value := args[1]

	switch key {
	case "sprint":
		config.DefaultSprint = value
	case "board":
		config.BoardName = value
	default:
		fmt.Printf("Unknown config key: %s\n", key)
		return
	}

	data, err := yaml.Marshal(&config)
	if err != nil {
		fmt.Println("Error saving config:", err)
		return
	}

	home, _ := os.UserHomeDir()
	configPath := filepath.Join(home, ".youtrack-cli.yaml")
	if err := os.WriteFile(configPath, data, 0600); err != nil {
		fmt.Println("Error saving config file:", err)
		return
	}
	fmt.Printf("Configuration updated: %s = %s\n", key, value)
}

func viewConfig(cmd *cobra.Command, args []string) {
	config, err := loadConfig()
	if err != nil {
		fmt.Println(err)
		return
	}
	data, err := yaml.Marshal(&config)
	if err != nil {
		fmt.Println("Error formatting config:", err)
		return
	}
	fmt.Println(string(data))
}

func showConfig(cmd *cobra.Command, args []string) {
	config, err := loadConfig()
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("YouTrack URL: %s\n", config.URL)
	// Mask the token for security
	if len(config.Token) > 8 {
		fmt.Printf("API Token: %s...%s\n", config.Token[:4], config.Token[len(config.Token)-4:])
	} else {
		fmt.Printf("API Token: %s\n", config.Token)
	}
	fmt.Printf("Default Board: %s\n", config.BoardName)
	fmt.Printf("Default Sprint: %s\n", config.DefaultSprint)
}

func listBoards(cmd *cobra.Command, args []string) {
	config, err := loadConfig()
	if err != nil {
		fmt.Println(err)
		return
	}

	fields := "id,name"
	apiURL := fmt.Sprintf("%s/api/agiles?fields=%s", config.URL, fields)

	client := &http.Client{}
	req, _ := http.NewRequest("GET", apiURL, nil)
	req.Header.Set("Authorization", "Bearer "+config.Token)
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error fetching boards:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Error fetching boards: %s - %s\n", resp.Status, string(body))
		return
	}

	var boards []AgileBoard
	if err := json.NewDecoder(resp.Body).Decode(&boards); err != nil {
		fmt.Println("Error decoding boards:", err)
		return
	}

	fmt.Printf("%-30s\t%s\n", "BOARD NAME", "ID")
	for _, board := range boards {
		fmt.Printf("%-30s\t%s\n", board.Name, board.ID)
	}
}

func listSprints(cmd *cobra.Command, args []string) {
	config, err := loadConfig()
	if err != nil {
		fmt.Println(err)
		return
	}

	boardName, _ := cmd.Flags().GetString("board")
	if boardName == "" {
		boardName = config.BoardName
	}

	if boardName == "" {
		fmt.Println("Error: Board name not specified. Please use the --board flag or set a default board using 'youtrack-cli config set board [board_name]'")
		return
	}

	// First, get all boards to find the ID of the specified board
	boardsURL := fmt.Sprintf("%s/api/agiles?fields=id,name", config.URL)
	client := &http.Client{}
	req, _ := http.NewRequest("GET", boardsURL, nil)
	req.Header.Set("Authorization", "Bearer "+config.Token)
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error fetching boards:", err)
		return
	}
	defer resp.Body.Close()

	var boards []AgileBoard
	if err := json.NewDecoder(resp.Body).Decode(&boards); err != nil {
		fmt.Println("Error decoding boards:", err)
		return
	}

	var boardID string
	for _, b := range boards {
		if b.Name == boardName {
			boardID = b.ID
			break
		}
	}

	if boardID == "" {
		fmt.Printf("Error: Board '%s' not found\n", boardName)
		return
	}

	// Now, get the sprints for that board
	sprintsURL := fmt.Sprintf("%s/api/agiles/%s/sprints?fields=id,name", config.URL, boardID)
	req, _ = http.NewRequest("GET", sprintsURL, nil)
	req.Header.Set("Authorization", "Bearer "+config.Token)
	req.Header.Set("Accept", "application/json")

	resp, err = client.Do(req)
	if err != nil {
		fmt.Println("Error fetching sprints:", err)
		return
	}
	defer resp.Body.Close()

	var sprints []Sprint
	if err := json.NewDecoder(resp.Body).Decode(&sprints); err != nil {
		fmt.Println("Error decoding sprints:", err)
		return
	}

	fmt.Printf("Sprints in board '%s':\n", boardName)
	for _, sprint := range sprints {
		fmt.Println(sprint.Name)
	}
}

func loadConfig() (Config, error) {
	var config Config
	home, _ := os.UserHomeDir()
	configPath := filepath.Join(home, ".youtrack-cli.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return config, err
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

	sprintName, _ := cmd.Flags().GetString("sprint")
	if sprintName == "" {
		sprintName = config.DefaultSprint
	}

	query := ""
	if sprintName == "" {
		query = "for:me"
	}

	if sprintName != "" {
		if config.BoardName == "" {
			fmt.Println("Error: Board name is not configured. Please set it using 'youtrack-cli config set board [board_name]'")
			return
		}
		// ------------------- FIX START -------------------
		// Sprint 名稱含空白也只需 {Sprint Name}，不要加雙引號
		boardPart := fmt.Sprintf("Board %s:", config.BoardName)
		sprintPart := fmt.Sprintf("{%s}", sprintName)
		query += fmt.Sprintf("%s %s", boardPart, sprintPart)
		// -------------------  FIX END  -------------------
	}

	fields := "idReadable,summary,customFields(name,value(name,presentation))"
	apiURL := fmt.Sprintf("%s/api/issues?fields=%s&query=%s", config.URL, fields, url.QueryEscape(query))
	fmt.Println("Debug API URL:", apiURL)

	client := &http.Client{}
	req, _ := http.NewRequest("GET", apiURL, nil)
	req.Header.Set("Authorization", "Bearer "+config.Token)
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error fetching issues:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Error fetching issues: %s - %s\n", resp.Status, string(body))
		return
	}

	var issues []Issue
	if err := json.NewDecoder(resp.Body).Decode(&issues); err != nil {
		fmt.Println("Error decoding issues:", err)
		return
	}

	// Fetch sprints for each issue
	for i := range issues {
		sprintsAPIURL := fmt.Sprintf("%s/api/issues/%s/sprints?fields=id,name", config.URL, issues[i].ID)
		sprintReq, _ := http.NewRequest("GET", sprintsAPIURL, nil)
		sprintReq.Header.Set("Authorization", "Bearer "+config.Token)
		sprintReq.Header.Set("Accept", "application/json")

		sprintResp, err := client.Do(sprintReq)
		if err != nil {
			fmt.Printf("Error fetching sprints for issue %s: %v\n", issues[i].ID, err)
			continue
		}
		defer sprintResp.Body.Close()

		if sprintResp.StatusCode == http.StatusOK {
			var issueSprints []Sprint
			if err := json.NewDecoder(sprintResp.Body).Decode(&issueSprints); err != nil {
				fmt.Printf("Error decoding sprints for issue %s: %v\n", issues[i].ID, err)
			} else {
				issues[i].Sprints = issueSprints
			}
		} else {
			body, _ := io.ReadAll(sprintResp.Body)
			fmt.Printf("Error fetching sprints for issue %s: %s - %s\n", issues[i].ID, sprintResp.Status, string(body))
		}
	}

	// Print header
	fmt.Printf("%-15s\t%-10s\t%-15s\t%-12s\t%-12s\t%-15s\t%s\n", "ID", "Type", "Status", "Estimation", "Spent Time", "Sprint", "Title")

	for _, issue := range issues {
		issueData := map[string]string{
			"ID":         issue.ID,
			"Title":      issue.Summary,
			"Type":       "N/A",
			"Status":     "N/A",
			"Estimation": "N/A",
			"Spent Time": "N/A",
			"Sprint":     "N/A",
		}

		for _, cf := range issue.CustomFields {
			var value string
			if cf.Value != nil {
				if valMap, ok := cf.Value.(map[string]interface{}); ok {
					if name, ok := valMap["name"].(string); ok {
						value = name
					} else if presentation, ok := valMap["presentation"].(string); ok {
						value = presentation
					}
				}
			}

			switch cf.Name {
			case "Type":
				issueData["Type"] = value
			case "State":
				issueData["Status"] = value
			case "Estimation":
				issueData["Estimation"] = value
			case "Spent time":
				issueData["Spent Time"] = value
			}
		}

		// Populate Sprint from the fetched sprints
		if len(issue.Sprints) > 0 {
			var sprintNames []string
			for _, s := range issue.Sprints {
				sprintNames = append(sprintNames, s.Name)
			}
			issueData["Sprint"] = strings.Join(sprintNames, ", ")
		}

		fmt.Printf("%-15s\t%-10s\t%-15s\t%-12s\t%-12s\t%-15s\t%s\n",
			issueData["ID"],
			issueData["Type"],
			issueData["Status"],
			issueData["Estimation"],
			issueData["Spent Time"],
			issueData["Sprint"],
			issueData["Title"],
		)
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

	apiURL := fmt.Sprintf("%s/api/issues/%s/timeTracking/workItems?fields=date,duration(minutes),author(login),text", config.URL, issueID)

	workItem := map[string]interface{}{
		"duration": map[string]string{"presentation": minutes + "m"},
		"text":     description,
	}
	jsonData, _ := json.Marshal(workItem)

	client := &http.Client{}
	req, _ := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
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
	apiURL := fmt.Sprintf("%s/api/issues?fields=idReadable,summary,updated&query=for:me", config.URL)
	req, _ := http.NewRequest("GET", apiURL, nil)
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
			if itemDate.Truncate(24 * time.Hour).Equal(today) {
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
	listCmd.Flags().StringP("sprint", "s", "", "Specify the sprint to list issues from")
	sprintListCmd.Flags().StringP("board", "b", "", "Board name to list sprints from")
}
