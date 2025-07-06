package youtrack

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"time"
	"youtrack-cli/internal/config"
)

// DetermineSprint 回傳要使用的 Sprint 名稱。
// 優先序：CLI flag > config.DefaultSprint > 自動偵測。
func DetermineSprint(cfg config.Config, flagSprintName string) (string, error) {
	// 1) CLI flag 最高
	if flagSprintName != "" {
		return flagSprintName, nil
	}
	// 2) ~/.youtrack-cli.yaml 內的預設
	if cfg.DefaultSprint != "" {
		return cfg.DefaultSprint, nil
	}
	// 3) 自動偵測須先知道 Board 名稱
	if cfg.BoardName == "" {
		return "", fmt.Errorf("board name is not configured, cannot auto-detect sprint")
	}

	sprints, err := ListSprints(cfg, cfg.BoardName) // 需帶 isCurrent/start/finish
	if err != nil {
		return "", err
	}
	if len(sprints) == 0 {
		return "", fmt.Errorf("no sprint found on board %q", cfg.BoardName)
	}

	// ────────────────────────────────
	// Step 1: isCurrent == true
	for _, sp := range sprints {
		if sp.IsCurrent {
			return sp.Name, nil
		}
	}

	now := time.Now()

	// Step 2: 日期區間涵蓋今天
	for _, sp := range sprints {
		if inRange(sp, now) {
			return sp.Name, nil
		}
	}

	// Step 3: 最近結束的 Sprint
	var past *Sprint
	pastDiff := time.Duration(1<<63 - 1)
	for i := range sprints {
		sp := &sprints[i]
		finish := unixMilliToTime(sp.Finish)
		if finish.IsZero() || now.Before(finish) {
			continue
		}
		if d := now.Sub(finish); d < pastDiff {
			pastDiff = d
			past = sp
		}
	}
	if past != nil {
		return past.Name, nil
	}

	// Step 4: 即將開始的 Sprint
	var future *Sprint
	futureDiff := time.Duration(1<<63 - 1)
	for i := range sprints {
		sp := &sprints[i]
		start := unixMilliToTime(sp.Start)
		if start.IsZero() || now.After(start) {
			continue
		}
		if d := start.Sub(now); d < futureDiff {
			futureDiff = d
			future = sp
		}
	}
	if future != nil {
		return future.Name, nil
	}

	// Step 5: 名稱尾數字最大
	sort.Slice(sprints, func(i, j int) bool {
		return extractNumber(sprints[i].Name) > extractNumber(sprints[j].Name)
	})
	return sprints[0].Name, nil
}

/* ───────────────────────────────
   Helper functions
   ─────────────────────────────── */

func inRange(sp Sprint, t time.Time) bool {
	if sp.Start == 0 || sp.Finish == 0 {
		return false
	}
	start := unixMilliToTime(sp.Start)
	finish := unixMilliToTime(sp.Finish)
	return !t.Before(start) && !t.After(finish)
}

func unixMilliToTime(ms int64) time.Time {
	if ms == 0 {
		return time.Time{}
	}
	return time.UnixMilli(ms)
}

func extractNumber(name string) int {
	re := regexp.MustCompile(`\d+`)
	m := re.FindAllString(name, -1)
	if len(m) == 0 {
		return 0
	}
	n, _ := strconv.Atoi(m[len(m)-1])
	return n
}
