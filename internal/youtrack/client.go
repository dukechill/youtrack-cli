package youtrack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"  // 新增：用於解析估時字串
	"strconv" // 新增：用於解析估時字串
	"strings"
	"youtrack-cli/internal/config"

	"time"
)

type DailySyncOptions struct {
	Minutes         int
	Comment         string
	State           string
	WorkDescription string
}

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

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
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

// --- YouTrack API specific functions ---

// FetchIssues fetches YouTrack issues based on a query.
func FetchIssues(cfg config.Config, query string) ([]Issue, error) {
	client := NewClient(cfg)
	fields := "idReadable,summary,customFields(name,value(login,fullName,presentation,name)),assignee(fullName,login)"
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

	fields := "id,name,isCurrent,start,finish"
	const pageSize = 100
	var allSprints []Sprint

	for skip := 0; ; skip += pageSize {
		path := fmt.Sprintf(
			"/api/agiles/%s/sprints?fields=%s&$top=%d&$skip=%d",
			boardID,
			fields,
			pageSize,
			skip,
		)

		var page []Sprint
		if err := client.get(path, &page); err != nil {
			return nil, err
		}

		allSprints = append(allSprints, page...)
		if len(page) < pageSize {
			break
		}
	}

	return allSprints, nil
}

// ListIssueSprints fetches the sprints associated with a single issue.
func ListIssueSprints(cfg config.Config, issueID string) ([]Sprint, error) {
	client := NewClient(cfg)
	path := fmt.Sprintf("/api/issues/%s/sprints?fields=id,name,isCurrent,start,finish,archived", issueID)

	var sprints []Sprint
	if err := client.get(path, &sprints); err != nil {
		return nil, err
	}

	return sprints, nil
}

// InspectIssue fetches one issue and related read-only evidence for work inventory.
func InspectIssue(cfg config.Config, issueID string) (IssueInspect, error) {
	client := NewClient(cfg)
	result := IssueInspect{}

	escapedIssueID := url.PathEscape(issueID)
	issueFields := strings.Join([]string{
		"idReadable",
		"summary",
		"customFields(name,value(login,fullName,presentation,name,minutes))",
	}, ",")
	issuePath := fmt.Sprintf("/api/issues/%s?fields=%s", escapedIssueID, issueFields)
	if err := client.get(issuePath, &result.Issue); err != nil {
		return result, err
	}

	commentsPath := fmt.Sprintf(
		"/api/issues/%s/comments?fields=id,created,updated,text,author(login,fullName)",
		escapedIssueID,
	)
	if err := client.get(commentsPath, &result.Comments); err != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf("comments: %v", err))
	}

	workItemsPath := fmt.Sprintf(
		"/api/issues/%s/timeTracking/workItems?fields=date,duration(minutes,presentation),author(login,fullName),text",
		escapedIssueID,
	)
	if err := client.get(workItemsPath, &result.WorkItems); err != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf("work items: %v", err))
	}

	sprints, err := ListIssueSprints(cfg, issueID)
	if err != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf("sprints: %v", err))
	} else {
		result.Sprints = sprints
	}

	linksPath := fmt.Sprintf(
		"/api/issues/%s/links?fields=linkType(name,localizedName,sourceToTarget,targetToSource,directed),issues(idReadable,summary)",
		escapedIssueID,
	)
	if err := client.get(linksPath, &result.Links); err != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf("links: %v", err))
	}

	return result, nil
}

// ApplyCommand applies a YouTrack command to one or more issues.
func ApplyCommand(cfg config.Config, query string, issueIDs []string) error {
	client := NewClient(cfg)

	issues := make([]map[string]string, 0, len(issueIDs))
	for _, issueID := range issueIDs {
		issues = append(issues, map[string]string{"idReadable": issueID})
	}

	body := map[string]any{
		"query":  query,
		"issues": issues,
	}

	return client.post("/api/commands", body, nil)
}

// SetIssueSprint assigns an issue to a specific sprint on a board.
func SetIssueSprint(cfg config.Config, issueID, boardName, sprintName string) error {
	return SetIssuesSprint(cfg, []string{issueID}, boardName, sprintName)
}

// SetIssuesSprint assigns one or more issues to a specific sprint on a board.
func SetIssuesSprint(cfg config.Config, issueIDs []string, boardName, sprintName string) error {
	if strings.TrimSpace(boardName) == "" {
		return fmt.Errorf("board name is required")
	}
	if strings.TrimSpace(sprintName) == "" {
		return fmt.Errorf("sprint name is required")
	}
	if len(issueIDs) == 0 {
		return fmt.Errorf("at least one issue id is required")
	}

	query := fmt.Sprintf("board %s %s", strings.TrimSpace(boardName), strings.TrimSpace(sprintName))
	return ApplyCommand(cfg, query, issueIDs)
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

// AddComment adds a comment to a YouTrack issue.
func AddComment(cfg config.Config, issueID, text string) error {
	client := NewClient(cfg)
	path := fmt.Sprintf("/api/issues/%s/comments?fields=id,text", issueID)

	comment := map[string]string{
		"text": text,
	}

	return client.post(path, comment, nil)
}

// UpdateState updates the State field on a YouTrack issue.
func UpdateState(cfg config.Config, issueID, state string) error {
	client := NewClient(cfg)
	path := fmt.Sprintf("/api/issues/%s?fields=idReadable,customFields(name,value(name))", issueID)

	issueUpdate := map[string]interface{}{
		"customFields": []map[string]interface{}{
			{
				"name":  "State",
				"$type": "StateIssueCustomField",
				"value": map[string]string{
					"name": state,
				},
			},
		},
	}

	return client.post(path, issueUpdate, nil)
}

// UpdateEstimation updates the Estimation field on a YouTrack issue.
func UpdateEstimation(cfg config.Config, issueID string, minutes int) error {
	client := NewClient(cfg)
	path := fmt.Sprintf("/api/issues/%s?fields=idReadable,customFields(name,value(presentation))", issueID)

	issueUpdate := map[string]interface{}{
		"customFields": []map[string]interface{}{
			{
				"name":  "Estimation",
				"$type": "PeriodIssueCustomField",
				"value": map[string]string{
					"presentation": fmt.Sprintf("%dm", minutes),
					"$type":        "PeriodValue",
				},
			},
		},
	}

	return client.post(path, issueUpdate, nil)
}

// DailySync logs work, adds a comment, and/or updates state for a single issue.
func DailySync(cfg config.Config, issueID string, opts DailySyncOptions) ([]string, error) {
	var actions []string

	if opts.Minutes <= 0 && opts.Comment == "" && opts.State == "" {
		return nil, fmt.Errorf("nothing to do: provide --minutes, --comment, or --state")
	}

	if opts.Minutes > 0 {
		workDescription := opts.WorkDescription
		if workDescription == "" {
			workDescription = opts.Comment
		}
		if workDescription == "" {
			return nil, fmt.Errorf("work log requires --work or --comment when --minutes is set")
		}

		if err := AddWorkItem(cfg, issueID, strconv.Itoa(opts.Minutes), workDescription); err != nil {
			return actions, fmt.Errorf("log work item: %w", err)
		}
		actions = append(actions, fmt.Sprintf("logged %d minutes", opts.Minutes))
	}

	if opts.Comment != "" {
		if err := AddComment(cfg, issueID, opts.Comment); err != nil {
			return actions, fmt.Errorf("add comment: %w", err)
		}
		actions = append(actions, "added comment")
	}

	if opts.State != "" {
		if err := UpdateState(cfg, issueID, opts.State); err != nil {
			return actions, fmt.Errorf("update state: %w", err)
		}
		actions = append(actions, fmt.Sprintf("set state to %s", opts.State))
	}

	return actions, nil
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
// 新增：issueType 參數
func BuildQuery(sprintName, assigneeName, issueType, boardName string) string {
	var parts []string

	// 1) 處理指派人過濾
	if assigneeName == "" { // 如果沒有指定指派人，預設為 for:me
		parts = append(parts, "for:me")
	} else if assigneeName == "me" {
		parts = append(parts, "for:me")
	} else if assigneeName == "unassigned" {
		parts = append(parts, "assignee: unassigned")
	} else {
		parts = append(parts, fmt.Sprintf("for: %s", assigneeName))
	}

	// 2) 處理 Type 過濾
	if issueType != "" {
		parts = append(parts, fmt.Sprintf("Type: %s", issueType))
	}

	// 3) 處理 Sprint 過濾
	if sprintName != "" {
		if boardName == "" {
			fmt.Println("Error: Board name is not configured. Use `youtrack-cli config set board ...`")
			return strings.Join(parts, " ") // 如果沒有 boardName，則不進行 sprint 過濾
		}
		// YouTrack 查詢語法中，Board 和 Sprint 名稱如果包含空格，需要用雙引號包起來
		// 但在 BuildQuery 內部，我們只構建語法，不進行 URL 編碼
		boardPart := fmt.Sprintf("Board %s:", boardName) // 退回變更：移除雙引號
		sprintPart := fmt.Sprintf("{%s}", sprintName)    // 退回變更：移除雙引號
		parts = append(parts, boardPart+" "+sprintPart)
	}

	query := strings.Join(parts, " ")
	fmt.Println("Debug Query:", query) // 保留除錯輸出

	return query
}

// PrintIssues prints YouTrack issues in a formatted table.
func PrintIssues(issues []Issue) {
	header := "%-15s\t%-10s\t%-15s\t%-12s\t%-12s\t%-15s\t%-20s\t%s\n"
	row := "%-15s\t%-10s\t%-15s\t%-12s\t%-12s\t%-15s\t%-20s\t%s\n"

	fmt.Printf(header, "ID", "Type", "Status", "Estimation", "Spent Time", "Sprint", "Assignee", "Title")

	for _, iss := range issues {

		// ---------- 1. 先抓 Assignee ----------
		assignee := "unassigned"
		// 這裡的 iss.Assignee 是 Issue 結構體中的 Assignee 欄位
		// 根據 models.go，Issue 結構體沒有 Assignee 欄位
		// 而是透過 customFields 或直接在 FetchIssues 中處理
		for _, cf := range iss.CustomFields {
			if cf.Name == "Assignee" || cf.Name == "Assignee(s)" {
				if names := extractAssigneeNames(cf.Value); len(names) > 0 {
					assignee = strings.Join(names, ", ")
				}
			}
		}

		// ---------- 2. 解析其他欄位 ----------
		data := map[string]string{
			"Type":       "N/A",
			"Status":     "N/A",
			"Estimation": "N/A",
			"Spent Time": "N/A",
		}
		for _, cf := range iss.CustomFields {
			val := presentation(cf.Value)
			switch cf.Name {
			case "Type":
				data["Type"] = val
			case "State":
				data["Status"] = val
			case "Estimation":
				data["Estimation"] = val
			case "Spent time":
				data["Spent Time"] = val
			}
		}

		// ---------- 3. Sprint 名稱串起來 ----------
		sprint := "N/A"
		if len(iss.Sprints) > 0 {
			var ss []string
			for _, s := range iss.Sprints {
				ss = append(ss, s.Name)
			}
			sprint = strings.Join(ss, ", ")
		}

		// ---------- 4. 輸出 ----------
		fmt.Printf(row, iss.ID, data["Type"], data["Status"], data["Estimation"],
			data["Spent Time"], sprint, assignee, iss.Summary)
	}
}

/* --- 小工具 ---------------------------------------------------- */

// 把 CustomField.Value 轉成可閱讀字串
func presentation(v interface{}) string {
	if v == nil {
		return ""
	}
	if m, ok := v.(map[string]interface{}); ok {
		if p, ok := m["presentation"].(string); ok && p != "" {
			return p
		}
		if n, ok := m["name"].(string); ok {
			return n
		}
	}
	return fmt.Sprintf("%v", v)
}

// 從 Assignee custom field 提取人名 (支援單人 / 多人陣列)
func extractAssigneeNames(v interface{}) []string {
	var names []string
	switch val := v.(type) {
	case map[string]interface{}:
		if fn, ok := val["fullName"].(string); ok && fn != "" {
			names = append(names, fn)
		}
	case []interface{}:
		for _, item := range val {
			if m, ok := item.(map[string]interface{}); ok {
				if fn, ok := m["fullName"].(string); ok && fn != "" {
					names = append(names, fn)
				}
			}
		}
	}
	return names
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
		marker := " "
		if sprint.IsCurrent {
			marker = "*"
		}
		if sprint.IsCurrent {
			fmt.Printf("%s %s [current]\n", marker, sprint.Name)
			continue
		}
		fmt.Printf("%s %s\n", marker, sprint.Name)
	}
}

// PrintIssueSprints prints the sprint membership for a single issue.
func PrintIssueSprints(issueID string, sprints []Sprint) {
	fmt.Printf("Sprints for issue '%s':\n", issueID)
	if len(sprints) == 0 {
		fmt.Println("  (none)")
		return
	}

	for _, sprint := range sprints {
		marker := " "
		if sprint.IsCurrent {
			marker = "*"
		}
		if sprint.IsCurrent {
			fmt.Printf("%s %s [current]\n", marker, sprint.Name)
			continue
		}
		fmt.Printf("%s %s\n", marker, sprint.Name)
	}
}

// PrintIssueInspect prints issue details and recent evidence for progress review.
func PrintIssueInspect(result IssueInspect) {
	issue := result.Issue
	fmt.Printf("%s: %s\n", issue.ID, issue.Summary)
	fmt.Println()

	fmt.Println("Fields:")
	printField("Type", customFieldPresentation(issue.CustomFields, "Type"))
	printField("State", customFieldPresentation(issue.CustomFields, "State"))
	printField("Priority", customFieldPresentation(issue.CustomFields, "Priority"))
	printField("Assignee", customFieldPresentation(issue.CustomFields, "Assignee", "Assignee(s)"))
	printField("Estimation", customFieldPresentation(issue.CustomFields, "Estimation"))
	printField("Spent time", customFieldPresentation(issue.CustomFields, "Spent time"))
	fmt.Println()

	fmt.Println("Sprints:")
	if len(result.Sprints) == 0 {
		fmt.Println("  (none)")
	} else {
		for _, sprint := range result.Sprints {
			suffix := ""
			if sprint.IsCurrent {
				suffix = " [current]"
			}
			fmt.Printf("  - %s%s\n", sprint.Name, suffix)
		}
	}
	fmt.Println()

	fmt.Println("Recent comments:")
	if len(result.Comments) == 0 {
		fmt.Println("  (none)")
	} else {
		for _, comment := range lastIssueComments(result.Comments, 5) {
			fmt.Printf("  - %s %s: %s\n", formatMillis(comment.Created), authorName(comment.Author), compactText(comment.Text, 180))
		}
	}
	fmt.Println()

	fmt.Println("Recent work items:")
	if len(result.WorkItems) == 0 {
		fmt.Println("  (none)")
	} else {
		for _, item := range lastWorkItems(result.WorkItems, 5) {
			duration := item.Duration.Presentation
			if duration == "" && item.Duration.Minutes > 0 {
				duration = HumanizeDuration(time.Duration(item.Duration.Minutes) * time.Minute)
			}
			fmt.Printf("  - %s %s %s: %s\n", formatMillis(item.Date), duration, authorName(item.Author), compactText(item.Text, 180))
		}
	}
	fmt.Println()

	fmt.Println("Links:")
	if len(result.Links) == 0 {
		fmt.Println("  (none)")
	} else {
		printedLinks := 0
		for _, link := range result.Links {
			label := link.LinkType.LocalizedName
			if label == "" {
				label = link.LinkType.Name
			}
			if label == "" {
				label = link.LinkType.TargetToSource
			}
			if label == "" {
				label = "linked"
			}
			for _, linkedIssue := range link.Issues {
				fmt.Printf("  - %s: %s %s\n", label, linkedIssue.ID, linkedIssue.Summary)
				printedLinks++
			}
		}
		if printedLinks == 0 {
			fmt.Println("  (none)")
		}
	}

	if len(result.Warnings) > 0 {
		fmt.Println()
		fmt.Println("Warnings:")
		for _, warning := range result.Warnings {
			fmt.Printf("  - %s\n", warning)
		}
	}
}

func printField(name, value string) {
	if value == "" {
		value = "N/A"
	}
	fmt.Printf("  %-12s %s\n", name+":", value)
}

func customFieldPresentation(fields []CustomField, names ...string) string {
	for _, wanted := range names {
		for _, cf := range fields {
			if cf.Name == wanted {
				if names := extractAssigneeNames(cf.Value); len(names) > 0 {
					return strings.Join(names, ", ")
				}
				return presentation(cf.Value)
			}
		}
	}
	return ""
}

func authorName(author Author) string {
	if author.FullName != "" {
		return author.FullName
	}
	if author.Login != "" {
		return author.Login
	}
	return "unknown"
}

func formatMillis(ms int64) string {
	if ms <= 0 {
		return "unknown-date"
	}
	return time.UnixMilli(ms).Format("2006-01-02 15:04")
}

func compactText(text string, maxLen int) string {
	cleaned := strings.Join(strings.Fields(text), " ")
	if len(cleaned) <= maxLen {
		return cleaned
	}
	if maxLen <= 3 {
		return cleaned[:maxLen]
	}
	return cleaned[:maxLen-3] + "..."
}

func lastIssueComments(comments []IssueComment, limit int) []IssueComment {
	if len(comments) <= limit {
		return comments
	}
	return comments[len(comments)-limit:]
}

func lastWorkItems(items []WorkItem, limit int) []WorkItem {
	if len(items) <= limit {
		return items
	}
	return items[len(items)-limit:]
}

// parseEstimation parses a YouTrack estimation string (e.g., "3h", "2d 4h", "45m") into a time.Duration.
// Assumes 1d = 6h.
func parseEstimation(str string) time.Duration {
	var totalDuration time.Duration

	// Regex to find days, hours, and minutes
	re := regexp.MustCompile(`(\d+)(d|h|m)`)
	matches := re.FindAllStringSubmatch(str, -1)

	for _, match := range matches {
		value, err := strconv.Atoi(match[1])
		if err != nil {
			continue // Skip if value is not a valid number
		}
		unit := match[2]

		switch unit {
		case "d":
			totalDuration += time.Duration(value*6) * time.Hour // 1 day = 6 hours
		case "h":
			totalDuration += time.Duration(value) * time.Hour
		case "m":
			totalDuration += time.Duration(value) * time.Minute
		}
	}
	return totalDuration
}

// SumEstimation calculates the total estimation from a slice of Issues.
func SumEstimation(issues []Issue) time.Duration {
	var total time.Duration
	for _, issue := range issues {
		for _, cf := range issue.CustomFields {
			if cf.Name == "Estimation" {
				if cf.Value != nil {
					// YouTrack API returns PeriodValue as a map with "presentation" key
					if valMap, ok := cf.Value.(map[string]interface{}); ok {
						if presentation, ok := valMap["presentation"].(string); ok {
							total += parseEstimation(presentation)
						}
					}
				}
				break // Found Estimation, move to next issue
			}
		}
	}
	return total
}

// HumanizeDuration converts a time.Duration into a human-readable string (e.g., "1d 3h 45m").
// Assumes 1d = 6h for display purposes.
func HumanizeDuration(d time.Duration) string {
	if d == 0 {
		return "0m"
	}

	var parts []string

	// Calculate days (1d = 6h)
	totalHours := int(d.Hours())
	days := totalHours / 6
	if days > 0 {
		parts = append(parts, fmt.Sprintf("%dd", days))
		totalHours %= 6
	}

	// Calculate remaining hours
	hours := totalHours
	if hours > 0 {
		parts = append(parts, fmt.Sprintf("%dh", hours))
	}

	// Calculate remaining minutes
	minutes := int(d.Minutes()) % 60
	if minutes > 0 || (len(parts) == 0 && minutes == 0 && d > 0) { // If no larger units, show minutes even if 0, but only if duration is not 0
		parts = append(parts, fmt.Sprintf("%dm", minutes))
	}

	return strings.Join(parts, " ")
}
