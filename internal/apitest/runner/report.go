package runner

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

type Report struct {
	PanelURL   string    `json:"panel_url"`
	Timestamp  string    `json:"timestamp"`
	Summary    Summary   `json:"summary"`
	ByCategory []CategorySummary `json:"by_category"`
	Results    []Result  `json:"results"`
	CleanupResults []Result `json:"cleanup_results,omitempty"`
}

type Summary struct {
	TotalEndpoints   int     `json:"total_endpoints"`
	Called          int     `json:"called"`
	Successful      int     `json:"successful"`
	Failed          int     `json:"failed"`
	Skipped         int     `json:"skipped"`
	AgentDependent  int     `json:"agent_dependent"`
	AvgResponseMS   int64   `json:"avg_response_ms"`
	Coverage        float64 `json:"coverage_percent"`
}

type CategorySummary struct {
	Category   string `json:"category"`
	Total      int    `json:"total"`
	Successful int    `json:"successful"`
	Failed     int    `json:"failed"`
	Skipped    int    `json:"skipped"`
}

func GenerateReport(baseURL string, results []Result, cleanupResults []Result) *Report {
	var totalDuration int64
	successful := 0
	failed := 0
	skipped := 0
	agentDep := 0

	categoryMap := make(map[string]CategorySummary)

	for _, r := range results {
		totalDuration += r.DurationMS
		if r.ErrorMessage == "skipped (non-destructive mode)" {
			skipped++
		} else if r.Success {
			successful++
		} else {
			failed++
		}
		if r.AgentDependent {
			agentDep++
		}

		cat := r.Category
		if cat == "" {
			cat = "misc"
		}
		cs := categoryMap[cat]
		cs.Category = cat
		cs.Total++
		if r.Success {
			cs.Successful++
		} else {
			cs.Failed++
		}
		categoryMap[cat] = cs
	}

	avgMS := int64(0)
	if len(results) > 0 {
		avgMS = totalDuration / int64(len(results))
	}

	coverage := 0.0
	if len(results) > 0 {
		coverage = float64(successful) / float64(len(results)) * 100
	}

	var categories []CategorySummary
	for _, cs := range categoryMap {
		categories = append(categories, cs)
	}

	return &Report{
		PanelURL:   baseURL,
		Timestamp:  time.Now().Format(time.RFC3339),
		Summary: Summary{
			TotalEndpoints: len(Endpoints),
			Called:         len(results),
			Successful:     successful,
			Failed:         failed,
			Skipped:        skipped,
			AgentDependent: agentDep,
			AvgResponseMS:  avgMS,
			Coverage:       coverage,
		},
		ByCategory:   categories,
		Results:      results,
		CleanupResults: cleanupResults,
	}
}

func (r *Report) SaveJSON(filename string) error {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal report: %w", err)
	}
	return os.WriteFile(filename, data, 0644)
}

func (r *Report) SaveMarkdown(filename string) error {
	var sb strings.Builder

	sb.WriteString("# API Test Report\n\n")
	sb.WriteString(fmt.Sprintf("**Panel URL:** %s\n", r.PanelURL))
	sb.WriteString(fmt.Sprintf("**Timestamp:** %s\n\n", r.Timestamp))

	sb.WriteString("## Summary\n\n")
	sb.WriteString(fmt.Sprintf("| Metric | Value |\n"))
	sb.WriteString(fmt.Sprintf("|--------|-------|\n"))
	sb.WriteString(fmt.Sprintf("| Total Endpoints | %d |\n", r.Summary.TotalEndpoints))
	sb.WriteString(fmt.Sprintf("| Called | %d |\n", r.Summary.Called))
	sb.WriteString(fmt.Sprintf("| Successful | %d |\n", r.Summary.Successful))
	sb.WriteString(fmt.Sprintf("| Failed | %d |\n", r.Summary.Failed))
	sb.WriteString(fmt.Sprintf("| Skipped | %d |\n", r.Summary.Skipped))
	sb.WriteString(fmt.Sprintf("| Agent Dependent | %d |\n", r.Summary.AgentDependent))
	sb.WriteString(fmt.Sprintf("| Avg Response Time | %d ms |\n", r.Summary.AvgResponseMS))
	sb.WriteString(fmt.Sprintf("| Success Rate | %.1f%% |\n\n", r.Summary.Coverage))

	sb.WriteString("## By Category\n\n")
	sb.WriteString(fmt.Sprintf("| Category | Total | Successful | Failed | Skipped |\n"))
	sb.WriteString(fmt.Sprintf("|----------|-------|------------|--------|--------|\n"))
	for _, cs := range r.ByCategory {
		sb.WriteString(fmt.Sprintf("| %s | %d | %d | %d | %d |\n",
			cs.Category, cs.Total, cs.Successful, cs.Failed, cs.Skipped))
	}
	sb.WriteString("\n")

	sb.WriteString("## Results\n\n")
	sb.WriteString(fmt.Sprintf("| Method | Path | Status | Duration (ms) | Category | Success | Error |\n"))
	sb.WriteString(fmt.Sprintf("|--------|------|--------|---------------|----------|---------|-------|\n"))

	for _, result := range r.Results {
		statusStr := fmt.Sprintf("%d", result.Status)
		successStr := "✅"
		if !result.Success {
			successStr = "❌"
		}
		errStr := ""
		if result.ErrorMessage != "" {
			errStr = result.ErrorMessage
		} else if !result.Success {
			errStr = "Non-2xx status"
		}
		if len(errStr) > 50 {
			errStr = errStr[:50] + "..."
		}
		sb.WriteString(fmt.Sprintf("| %s | %s | %s | %d | %s | %s | %s |\n",
			result.Method, result.Path, statusStr, result.DurationMS, result.Category, successStr, errStr))
	}

	sb.WriteString("\n## Failed Endpoints\n\n")
	hasFailures := false
	for _, result := range r.Results {
		if !result.Success && result.ErrorMessage != "skipped (non-destructive mode)" {
			hasFailures = true
			break
		}
	}

	if !hasFailures {
		sb.WriteString("*No failed endpoints*\n\n")
	} else {
		sb.WriteString(fmt.Sprintf("| Method | Path | Status | Response |\n"))
		sb.WriteString(fmt.Sprintf("|--------|------|--------|----------|\n"))
		for _, result := range r.Results {
			if !result.Success && result.ErrorMessage != "skipped (non-destructive mode)" {
				bodyPreview := result.Body
				if len(bodyPreview) > 200 {
					bodyPreview = bodyPreview[:200] + "..."
				}
				bodyPreview = strings.ReplaceAll(bodyPreview, "\n", " ")
				sb.WriteString(fmt.Sprintf("| %s | %s | %d | %s |\n",
					result.Method, result.Path, result.Status, bodyPreview))
			}
		}
		sb.WriteString("\n")
	}

	if len(r.CleanupResults) > 0 {
		sb.WriteString("## Cleanup Results\n\n")
		cleanupFailed := 0
		for _, cr := range r.CleanupResults {
			if !cr.Success {
				cleanupFailed++
			}
		}
		sb.WriteString(fmt.Sprintf("| Metric | Value |\n"))
		sb.WriteString(fmt.Sprintf("|--------|-------|\n"))
		sb.WriteString(fmt.Sprintf("| Total Cleanup Ops | %d |\n", len(r.CleanupResults)))
		sb.WriteString(fmt.Sprintf("| Failed | %d |\n\n", cleanupFailed))
	}

	return os.WriteFile(filename, []byte(sb.String()), 0644)
}
