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
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

/* ───────────────────────────────
   型別定義
   ─────────────────────────────── */

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

type sprintInfo struct {
	Name      string `json:"name"`
	IsCurrent bool   `json:"isCurrent"`
	Start     int64  `json:"start"`
	Finish    int64  `json:"finish"`
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

/* ───────────────────────────────
   Cobra CLI 初始化
   ─────────────────────────────── */

var (
	rootCmd             = &cobra.Command{Use: "youtrack-cli", Short: "A CLI for interacting with YouTrack"}
	configCmd           = &cobra.Command{Use: "config", Short: "Configure YouTrack settings"}
	configSetCmd        = &cobra.Command{Use: "set [key] [value]", Short: "Set a configuration value (e.g., sprint, board)", Args: cobra.ExactArgs(2), Run: setConfigValue}
	configViewCmd       = &cobra.Command{Use: "view", Short: "View current configuration", Run: viewConfig}
	configShowCmd       = &cobra.Command{Use: "show", Short: "Show current configuration (hiding sensitive parts)", Run: showConfig}
	configListBoardsCmd = &cobra.Command{Use: "list-boards", Short: "List available agile boards", Run: listBoards}

	boardCmd     = &cobra.Command{Use: "board", Short: "Manage agile boards"}
	boardListCmd = &cobra.Command{Use: "list", Short: "List available agile boards", Run: listBoards}

	sprintCmd     = &cobra.Command{Use: "sprint", Short: "Manage sprints"}
	sprintListCmd = &cobra.Command{Use: "list", Short: "List sprints for a specific board", Run: listSprints}

	listCmd      = &cobra.Command{Use: "list", Short: "List your YouTrack issues", Run: listIssues}
	addWorkCmd   = &cobra.Command{Use: "add-work [issue-id] [minutes] [description]", Short: "Add a work item to a YouTrack issue", Args: cobra.ExactArgs(3), Run: addWorkItem}
	checkWorkCmd = &cobra.Command{Use: "check-work", Short: "Check for issues with no work logged today", Run: checkWork}
)

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

/* ───────────────────────────────
   Config & Helper
   ─────────────────────────────── */

func loadConfig() (Config, error) {
	var cfg Config
	home, _ := os.UserHomeDir()
	data, err := os.ReadFile(filepath.Join(home, ".youtrack-cli.yaml"))
	if err != nil {
		return cfg, err
	}
	return cfg, yaml.Unmarshal(data, &cfg)
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

func getJSON(client *http.Client, url, token string, v interface{}) error {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(v)
}

/* ───────────────────────────────
   listIssues 核心邏輯
   ─────────────────────────────── */

func listIssues(cmd *cobra.Command, args []string) {
	config, err := loadConfig()
	if err != nil {
		fmt.Println(err)
		return
	}

	/* 1. 永遠帶 for:me */
	queryParts := []string{"for:me"}

	/* 2. sprint 名稱優先順序：flag > default > {current sprint} */
	flagSprint, _ := cmd.Flags().GetString("sprint")
	sprintName := firstNonEmpty(flagSprint, config.DefaultSprint, "{current sprint}")

	/* 3. 若為 {current sprint}，解析看板找到實際名稱 */
	if sprintName == "{current sprint}" {
		name, err := resolveCurrentSprint(config)
		if err != nil {
			fmt.Println("Warning:", err) // fallback → 查全部
			sprintName = ""
		} else {
			sprintName = name
		}
	}

	if sprintName == "" {
		fmt.Println("Debug Sprint Name: <none>")
	} else {
		fmt.Println("Debug Sprint Name:", sprintName)
	}

	/* 4. 組 Board + Sprint 條件 */
	if config.BoardName != "" && sprintName != "" {
		boardPart := fmt.Sprintf("Board %s:", config.BoardName)
		if !strings.HasPrefix(sprintName, "{") {
			sprintName = "{" + sprintName + "}"
		}
		queryParts = append(queryParts, boardPart+" "+sprintName)
	}

	query := strings.Join(queryParts, " ")

	/* 5. 呼叫 Issue API */
	fields := "idReadable,summary,customFields(name,value(name,presentation))"
	apiURL := fmt.Sprintf("%s/api/issues?fields=%s&query=%s",
		config.URL, fields, url.QueryEscape(query))
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

	/********************************************************
	 * 6. 逐票補上 Sprint、估時等欄位（沿用你原本的列印邏輯）
	 ********************************************************/
	for i := range issues {
		sprintsAPIURL := fmt.Sprintf("%s/api/issues/%s/sprints?fields=id,name",
			config.URL, issues[i].ID)
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
			fmt.Printf("Error fetching sprints for issue %s: %s - %s\n",
				issues[i].ID, sprintResp.Status, string(body))
		}
	}

	// ----------- 印出結果（保持原有格式） -----------
	fmt.Printf("%-15s\t%-10s\t%-15s\t%-12s\t%-12s\t%-15s\t%s\n",
		"ID", "Type", "Status", "Estimation", "Spent Time", "Sprint", "Title")

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

		if len(issue.Sprints) > 0 {
			var names []string
			for _, s := range issue.Sprints {
				names = append(names, s.Name)
			}
			issueData["Sprint"] = strings.Join(names, ", ")
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

/* ---------- 解析 current sprint ---------- */

func resolveCurrentSprint(cfg Config) (string, error) {
	if cfg.BoardName == "" {
		return "", fmt.Errorf("Board name is required to resolve {current sprint}")
	}

	client := &http.Client{}

	// 1) 取得 boardID
	var boards []AgileBoard
	if err := getJSON(client, fmt.Sprintf("%s/api/agiles?fields=id,name", cfg.URL), cfg.Token, &boards); err != nil {
		return "", err
	}
	var boardID string
	for _, b := range boards {
		if b.Name == cfg.BoardName {
			boardID = b.ID
			break
		}
	}
	if boardID == "" {
		return "", fmt.Errorf("Board '%s' not found", cfg.BoardName)
	}

	// 2) 抓 sprints
	var sprints []sprintInfo
	url := fmt.Sprintf("%s/api/agiles/%s/sprints?fields=name,isCurrent,start,finish", cfg.URL, boardID)
	if err := getJSON(client, url, cfg.Token, &sprints); err != nil {
		return "", err
	}

	today := time.Now()
	reNum := regexp.MustCompile(`(\d+)$`)
	maxNum := -1
	fallback := ""

	for _, sp := range sprints {
		if sp.IsCurrent {
			return sp.Name, nil
		}
		if sp.Start > 0 && sp.Finish > 0 {
			start := time.UnixMilli(sp.Start)
			finish := time.UnixMilli(sp.Finish)
			if !today.Before(start) && !today.After(finish) {
				return sp.Name, nil
			}
		}
		if m := reNum.FindStringSubmatch(sp.Name); len(m) == 2 {
			if n, _ := strconv.Atoi(m[1]); n > maxNum {
				maxNum, fallback = n, sp.Name
			}
		}
	}
	if fallback != "" {
		return fallback, nil
	}
	return "", fmt.Errorf("no current sprint found")
}

/* ───────────────────────────────
   其他指令（listBoards、listSprints、addWorkItem、
   checkWork、config*）保持原樣
   ─────────────────────────────── */

func listBoards(cmd *cobra.Command, args []string) {
	cfg, err := loadConfig()
	if err != nil {
		fmt.Println(err)
		return
	}

	fields := "id,name"
	apiURL := fmt.Sprintf("%s/api/agiles?fields=%s", cfg.URL, fields)

	client := &http.Client{}
	req, _ := http.NewRequest("GET", apiURL, nil)
	req.Header.Set("Authorization", "Bearer "+cfg.Token)
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
	cfg, err := loadConfig()
	if err != nil {
		fmt.Println(err)
		return
	}

	boardName, _ := cmd.Flags().GetString("board")
	if boardName == "" {
		boardName = cfg.BoardName
	}
	if boardName == "" {
		fmt.Println("Error: Board name not specified.")
		return
	}

	// 找 boardID
	var boards []AgileBoard
	if err := getJSON(&http.Client{}, fmt.Sprintf("%s/api/agiles?fields=id,name", cfg.URL), cfg.Token, &boards); err != nil {
		fmt.Println("Error fetching boards:", err)
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

	var sprints []Sprint
	url := fmt.Sprintf("%s/api/agiles/%s/sprints?fields=id,name", cfg.URL, boardID)
	if err := getJSON(&http.Client{}, url, cfg.Token, &sprints); err != nil {
		fmt.Println("Error fetching sprints:", err)
		return
	}

	fmt.Printf("Sprints in board '%s':\n", boardName)
	for _, sp := range sprints {
		fmt.Println(sp.Name)
	}
}

func setConfigValue(cmd *cobra.Command, args []string) {
	cfg, _ := loadConfig() // ignore error, will be zero value if file 不存在
	key, val := args[0], args[1]

	switch key {
	case "sprint":
		cfg.DefaultSprint = val
	case "board":
		cfg.BoardName = val
	default:
		fmt.Printf("Unknown config key: %s\n", key)
		return
	}

	data, _ := yaml.Marshal(&cfg)
	home, _ := os.UserHomeDir()
	if err := os.WriteFile(filepath.Join(home, ".youtrack-cli.yaml"), data, 0600); err != nil {
		fmt.Println("Error saving config:", err)
		return
	}
	fmt.Printf("Configuration updated: %s = %s\n", key, val)
}

func viewConfig(cmd *cobra.Command, args []string) {
	cfg, err := loadConfig()
	if err != nil {
		fmt.Println(err)
		return
	}
	b, _ := yaml.Marshal(&cfg)
	fmt.Println(string(b))
}

func showConfig(cmd *cobra.Command, args []string) {
	cfg, err := loadConfig()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("YouTrack URL: %s\n", cfg.URL)
	if len(cfg.Token) > 8 {
		fmt.Printf("API Token: %s...%s\n", cfg.Token[:4], cfg.Token[len(cfg.Token)-4:])
	} else {
		fmt.Printf("API Token: %s\n", cfg.Token)
	}
	fmt.Printf("Default Board: %s\n", cfg.BoardName)
	fmt.Printf("Default Sprint: %s\n", cfg.DefaultSprint)
}

func addWorkItem(cmd *cobra.Command, args []string) {
	cfg, err := loadConfig()
	if err != nil {
		fmt.Println(err)
		return
	}

	issueID, minutes, desc := args[0], args[1], args[2]
	apiURL := fmt.Sprintf("%s/api/issues/%s/timeTracking/workItems?fields=date,duration(minutes),author(login),text", cfg.URL, issueID)

	payload := map[string]interface{}{
		"duration": map[string]string{"presentation": minutes + "m"},
		"text":     desc,
	}
	data, _ := json.Marshal(payload)

	client := &http.Client{}
	req, _ := http.NewRequest("POST", apiURL, bytes.NewBuffer(data))
	req.Header.Set("Authorization", "Bearer "+cfg.Token)
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
	cfg, err := loadConfig()
	if err != nil {
		fmt.Println(err)
		return
	}

	client := &http.Client{}
	apiURL := fmt.Sprintf("%s/api/issues?fields=idReadable,summary,updated&query=for:me", cfg.URL)
	req, _ := http.NewRequest("GET", apiURL, nil)
	req.Header.Set("Authorization", "Bearer "+cfg.Token)
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
	var noWork []string

	for _, iss := range issues {
		u := fmt.Sprintf("%s/api/issues/%s/timeTracking/workItems?fields=date", cfg.URL, iss.ID)
		if err := getJSON(client, u, cfg.Token, &[]WorkItem{}); err != nil { // quick check
			continue
		}
		workReq, _ := http.NewRequest("GET", u, nil)
		workReq.Header.Set("Authorization", "Bearer "+cfg.Token)
		workReq.Header.Set("Accept", "application/json")
		workResp, err := client.Do(workReq)
		if err != nil {
			continue
		}
		defer workResp.Body.Close()

		var items []WorkItem
		if err := json.NewDecoder(workResp.Body).Decode(&items); err != nil {
			continue
		}
		hasToday := false
		for _, it := range items {
			if time.Unix(it.Date/1000, 0).Truncate(24 * time.Hour).Equal(today) {
				hasToday = true
				break
			}
		}
		if !hasToday {
			noWork = append(noWork, fmt.Sprintf("%s: %s", iss.ID, iss.Summary))
		}
	}

	if len(noWork) > 0 {
		fmt.Println("You have not logged work for the following issues today:")
		for _, l := range noWork {
			fmt.Println("-", l)
		}
	} else {
		fmt.Println("All issues have work logged for today.")
	}
}

func init() {
	listCmd.Flags().StringP("sprint", "s", "", "Specify the sprint to list issues from")
	sprintListCmd.Flags().StringP("board", "b", "", "Board name to list sprints from")
}
