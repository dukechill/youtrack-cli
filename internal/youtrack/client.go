package youtrack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"                      // Added missing import
	"youtrack-cli/internal/config" // Import config package

	"time" // For http.Client timeout
)

// Client represents a YouTrack API client.
type Client struct {
	BaseURL    string
	Token      string
	HTTPClient *http.Client
}

// NewClient creates a new YouTrack API client.
func NewClient(cfg config.Config) *Client {
	return &Client{
		BaseURL: cfg.URL,
		Token:   cfg.Token,
		HTTPClient: &http.Client{
			Timeout: time.Second * 10, // Set a timeout for HTTP requests
		},
	}
}

// get performs a GET request to the YouTrack API and decodes the response into v.
func (c *Client) get(path string, v interface{}) error {
	apiURL := fmt.Sprintf("%s%s", c.BaseURL, path)
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %s: %s", resp.Status, string(body))
	}

	if v != nil {
		if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}
	return nil
}

// post performs a POST request to the YouTrack API with a JSON body and decodes the response into v.
func (c *Client) post(path string, body interface{}, v interface{}) error {
	apiURL := fmt.Sprintf("%s%s", c.BaseURL, path)
	jsonData, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %s: %s", resp.Status, string(bodyBytes))
	}

	if v != nil {
		if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}
	return nil
}

// --- YouTrack API specific functions (moved from main.go) ---

// FetchIssues fetches YouTrack issues based on a query.
func FetchIssues(cfg config.Config, query string) ([]Issue, error) {
	client := NewClient(cfg)
	fields := "idReadable,summary,customFields(name,value(name,presentation))"
	encodedQuery := url.QueryEscape(query)
	path := fmt.Sprintf("/api/issues?fields=%s&query=%s", fields, encodedQuery)

	var issues []Issue
	if err := client.get(path, &issues); err != nil {
		return nil, err
	}

	// Fetch sprints for each issue
	for i := range issues {
		sprintsPath := fmt.Sprintf("/api/issues/%s/sprints?fields=id,name", issues[i].ID)
		var issueSprints []Sprint
		if err := client.get(sprintsPath, &issueSprints); err != nil {
			// Log warning but continue if sprint fetching fails for a single issue
			fmt.Printf("Warning: Could not fetch sprints for issue %s: %v\n", issues[i].ID, err)
		} else {
			issues[i].Sprints = issueSprints
		}
	}

	return issues, nil
}

// ListBoards fetches all agile boards.
func ListBoards(cfg config.Config) ([]AgileBoard, error) {
	client := NewClient(cfg)
	fields := "id,name"
	path := fmt.Sprintf("/api/agiles?fields=%s", fields)

	var boards []AgileBoard
	if err := client.get(path, &boards); err != nil {
		return nil, err
	}
	return boards, nil
}

// ListSprints fetches sprints for a given board name.
func ListSprints(cfg config.Config, boardName string) ([]Sprint, error) {
	client := NewClient(cfg)

	// First, get all boards to find the ID of the specified board
	boards, err := ListBoards(cfg) // Reuse ListBoards
	if err != nil {
		return nil, err
	}

	var boardID string
	for _, b := range boards {
		if b.Name == boardName {
			boardID = b.ID
			break
		}
	}

	if boardID == "" {
		return nil, fmt.Errorf("board '%s' not found", boardName)
	}

	// Now, get the sprints for that board
	fields := "id,name,isCurrent,start,finish"
	path := fmt.Sprintf("/api/agiles/%s/sprints?fields=%s", boardID, fields)

	var sprints []Sprint
	if err := client.get(path, &sprints); err != nil {
		return nil, err
	}
	return sprints, nil
}

// AddWorkItem adds a work item to a YouTrack issue.
func AddWorkItem(cfg config.Config, issueID, minutes, description string) error {
	client := NewClient(cfg)
	path := fmt.Sprintf("/api/issues/%s/timeTracking/workItems?fields=date,duration(minutes),author(login),text", issueID)

	workItem := map[string]interface{}{
		"duration": map[string]string{"presentation": minutes + "m"},
		"text":     description,
	}

	return client.post(path, workItem, nil)
}

// CheckWork checks for issues with no work logged today.
func CheckWork(cfg config.Config) ([]string, error) {
	client := NewClient(cfg)
	path := "/api/issues?fields=idReadable,summary,updated&query=for:me"

	var issues []struct {
		ID      string `json:"idReadable"`
		Summary string `json:"summary"`
		Updated int64  `json:"updated"`
	}
	if err := client.get(path, &issues); err != nil {
		return nil, err
	}

	today := time.Now().Truncate(24 * time.Hour)
	var issuesWithoutWork []string

	for _, issue := range issues {
		workItemsPath := fmt.Sprintf("/api/issues/%s/timeTracking/workItems?fields=date", issue.ID)
		var workItems []WorkItem
		if err := client.get(workItemsPath, &workItems); err != nil {
			// Log warning but continue if fetching work items fails for a single issue
			fmt.Printf("Warning: Could not fetch work items for issue %s: %v\n", issue.ID, err)
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
	return issuesWithoutWork, nil
}

// BuildQuery constructs the YouTrack query string.
// sprintName 可為 ""；assigneeName 建議支援 "me" / "unassigned" / 指定使用者。
// boardName 必須有值才能使用 sprint 過濾。
func BuildQuery(sprintName, assigneeName, boardName string) string {
	var parts []string

	// 1) 預設只顯示自己的票
	if assigneeName == "" {
		parts = append(parts, "for:me")
	} else if assigneeName == "me" {
		parts = append(parts, "for:me")
	} else if assigneeName == "unassigned" {
		parts = append(parts, "assignee: unassigned")
	} else {
		parts = append(parts, fmt.Sprintf("for: %s", assigneeName))
	}

	// 2) Sprint 過濾（沒有任何雙引號）
	if sprintName != "" {
		if boardName == "" {
			fmt.Println("Error: Board name is not configured. Use `youtrack-cli config set board ...`")
			return strings.Join(parts, " ")
		}
		boardPart := fmt.Sprintf("Board %s:", boardName) // ← 刪掉雙引號
		sprintPart := fmt.Sprintf("{%s}", sprintName)    // ← 也不要雙引號
		parts = append(parts, boardPart+" "+sprintPart)
	}

	return strings.Join(parts, " ")
}

// PrintIssues prints YouTrack issues in a formatted table.
func PrintIssues(issues []Issue) {
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

// PrintBoards prints agile boards in a formatted table.
func PrintBoards(boards []AgileBoard) {
	fmt.Printf("%-30s\t%s\n", "BOARD NAME", "ID")
	for _, board := range boards {
		fmt.Printf("%-30s\t%s\n", board.Name, board.ID)
	}
}

// PrintSprints prints sprints for a given board in a formatted list.
func PrintSprints(boardName string, sprints []Sprint) {
	fmt.Printf("Sprints in board '%s':\n", boardName)
	for _, sprint := range sprints {
		fmt.Println(sprint.Name)
	}
}

