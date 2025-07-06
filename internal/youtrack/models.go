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
	ID   string `json:"id"`
	Name string `json:"name"`
	// Add other relevant sprint fields if needed for sorting/filtering
	Start      int64 `json:"start"`
	Finish     int64 `json:"finish"`
	IsArchived bool  `json:"archived"`
	IsCurrent  bool  `json:"isCurrent"` // YouTrack API might have this
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
