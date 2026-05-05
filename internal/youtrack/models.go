package youtrack

// Config struct is now in internal/config/file.go
// type Config struct { ... }

type Issue struct {
	ID           string        `json:"idReadable"`
	Summary      string        `json:"summary"`
	CustomFields []CustomField `json:"customFields"`
	Sprints      []Sprint      `json:"sprints,omitempty"` // Populated by separate API call
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
	ID        string `json:"id"`
	Name      string `json:"name"`
	IsCurrent bool   `json:"isCurrent"`
	// Add other relevant sprint fields if needed for sorting/filtering
	Start  int64 `json:"start"`  // 新增：Sprint 開始時間 (Unix timestamp in milliseconds)
	Finish int64 `json:"finish"` // 新增：Sprint 結束時間 (Unix timestamp in milliseconds)
	// IsArchived bool `json:"archived"`
	// IsCurrent  bool `json:"isCurrent"` // YouTrack API might have this
}

type IssueInspect struct {
	Issue     Issue
	Comments  []IssueComment
	WorkItems []WorkItem
	Sprints   []Sprint
	Links     []IssueLink
	Warnings  []string
}

type IssueComment struct {
	ID      string `json:"id"`
	Created int64  `json:"created"`
	Updated int64  `json:"updated"`
	Text    string `json:"text"`
	Author  Author `json:"author"`
}

type IssueLink struct {
	LinkType IssueLinkType `json:"linkType"`
	Issues   []Issue       `json:"issues"`
}

type IssueLinkType struct {
	Name           string `json:"name"`
	LocalizedName  string `json:"localizedName"`
	SourceToTarget string `json:"sourceToTarget"`
	TargetToSource string `json:"targetToSource"`
	Directed       bool   `json:"directed"`
}

type WorkItem struct {
	Date     int64    `json:"date"`
	Duration Duration `json:"duration"`
	Author   Author   `json:"author"`
	Text     string   `json:"text"`
}

type Duration struct {
	Minutes      int    `json:"minutes"`
	Presentation string `json:"presentation"`
}

type Author struct {
	Login    string `json:"login"`
	FullName string `json:"fullName"`
}
