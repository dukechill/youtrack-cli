package youtrack

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"time"
	"youtrack-cli/internal/config"
)

// DetermineSprint determines the sprint name to use based on flags, default config, or latest active sprint.
func DetermineSprint(cfg config.Config, flagSprintName string) (string, error) {
	// 1. If sprint name is provided via flag, use it directly
	if flagSprintName != "" {
		return flagSprintName, nil
	}

	// 2. If default sprint is configured, use it
	if cfg.DefaultSprint != "" {
		return cfg.DefaultSprint, nil
	}

	// 3. Otherwise, try to find the latest active sprint
	if cfg.BoardName == "" {
		return "", fmt.Errorf("board name is not configured, cannot determine latest sprint")
	}

	sprints, err := ListSprints(cfg, cfg.BoardName)
	if err != nil {
		return "", fmt.Errorf("failed to list sprints for board '%s': %w", cfg.BoardName, err)
	}

	if len(sprints) == 0 {
		return "", fmt.Errorf("no sprints found for board '%s'", cfg.BoardName)
	}

	// Heuristic 1: Sort by finish date (descending) if available
	// This assumes YouTrack API returns valid start/finish dates.
	sort.Slice(sprints, func(i, j int) bool {
		// Prioritize sprints with valid finish dates
		if sprints[i].Finish > 0 && sprints[j].Finish > 0 {
			return sprints[i].Finish > sprints[j].Finish // Newest finish date first
		} else if sprints[i].Finish > 0 { // i has finish date, j doesn't
			return true
		} else if sprints[j].Finish > 0 { // j has finish date, i doesn't
			return false
		}
		// Fallback to sorting by name with numbers if no valid finish dates
		numI := extractNumberFromName(sprints[i].Name)
		numJ := extractNumberFromName(sprints[j].Name)
		if numI != numJ {
			return numI > numJ // Sort by number descending
		}
		return sprints[i].Name > sprints[j].Name // Fallback to alphabetical if numbers are same
	})

	return sprints[0].Name, nil
}

// extractNumberFromName extracts a number from a sprint name for sorting.
func extractNumberFromName(name string) int {
	re := regexp.MustCompile(`(\d+)$`)
	matches := re.FindAllString(name, -1)
	if len(matches) > 0 {
		num, err := strconv.Atoi(matches[len(matches)-1]) // Take the last number found
		if err == nil {
			return num
		}
	}
	return 0 // Default if no number found or error
}

// Helper to convert Unix milliseconds to time.Time (if Start/Finish are Unix ms)
func unixMilliToTime(ms int64) time.Time {
	return time.Unix(ms/1000, (ms%1000)*int64(time.Millisecond))
}
