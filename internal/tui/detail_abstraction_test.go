package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/janekbaraniewski/openusage/internal/core"
)

func TestDetailTabs_SingleAllTab(t *testing.T) {
	snap := core.UsageSnapshot{
		ProviderID: "test",
		AccountID:  "test",
		Timestamp:  time.Now(),
		Metrics:    map[string]core.Metric{"rpm": {Used: core.Float64Ptr(10), Unit: "req"}},
		ModelUsage: []core.ModelUsageRecord{
			{RawModelID: "gpt-4", CostUSD: core.Float64Ptr(5.0)},
		},
		Attributes: map[string]string{"plan": "pro"},
	}

	tabs := DetailTabs(snap)
	if len(tabs) != 1 || tabs[0] != "All" {
		t.Fatalf("expected single All tab, got %v", tabs)
	}
}

func TestBuildDetailTrendsSection_IncludesBreakdownCharts(t *testing.T) {
	snap := core.UsageSnapshot{
		ProviderID: "claude_code",
		AccountID:  "test",
		Timestamp:  time.Now(),
		Metrics: map[string]core.Metric{
			"mcp_github_total": {Used: core.Float64Ptr(18)},
			"mcp_gopls_total":  {Used: core.Float64Ptr(10)},
		},
		DailySeries: map[string][]core.TimePoint{
			"cost": {
				{Date: "2026-02-18", Value: 5},
				{Date: "2026-02-19", Value: 8},
				{Date: "2026-02-20", Value: 12},
			},
			"usage_model_claude-opus-4-1": {
				{Date: "2026-02-18", Value: 3},
				{Date: "2026-02-19", Value: 5},
				{Date: "2026-02-20", Value: 7},
			},
			"usage_model_claude-haiku-4-5": {
				{Date: "2026-02-18", Value: 2},
				{Date: "2026-02-19", Value: 4},
				{Date: "2026-02-20", Value: 6},
			},
			"tokens_client_webapp": {
				{Date: "2026-02-18", Value: 1200},
				{Date: "2026-02-19", Value: 1500},
				{Date: "2026-02-20", Value: 1700},
			},
			"tokens_client_api_server": {
				{Date: "2026-02-18", Value: 900},
				{Date: "2026-02-19", Value: 1100},
				{Date: "2026-02-20", Value: 1400},
			},
			"usage_project_openusage": {
				{Date: "2026-02-18", Value: 6},
				{Date: "2026-02-19", Value: 9},
				{Date: "2026-02-20", Value: 11},
			},
			"usage_mcp_github": {
				{Date: "2026-02-18", Value: 4},
				{Date: "2026-02-19", Value: 6},
				{Date: "2026-02-20", Value: 8},
			},
			"usage_mcp_gopls": {
				{Date: "2026-02-18", Value: 2},
				{Date: "2026-02-19", Value: 3},
				{Date: "2026-02-20", Value: 5},
			},
		},
	}

	widget := core.DefaultDashboardWidget()
	widget.ShowClientComposition = true

	lines := buildDetailTrendsSection(snap, widget, 96, core.TimeWindowAll)
	out := stripANSI(strings.Join(lines, "\n"))

	for _, title := range []string{"Model Breakdown", "Client Breakdown", "Project Breakdown", "MCP Usage"} {
		if !strings.Contains(out, title) {
			t.Fatalf("expected %q chart in trends output, got:\n%s", title, out)
		}
	}
}

func TestRenderInfoSection_SplitsAttributesDiagnosticsRaw(t *testing.T) {
	snap := core.UsageSnapshot{
		ProviderID:  "test",
		AccountID:   "test",
		Timestamp:   time.Now(),
		Attributes:  map[string]string{"plan": "pro", "email": "test@example.com"},
		Diagnostics: map[string]string{"stale_data": "true"},
		Raw:         map[string]string{"api_version": "v2"},
	}

	var sb strings.Builder
	renderInfoSection(&sb, snap, core.DefaultDashboardWidget(), 80)
	out := sb.String()

	if !strings.Contains(out, "Attributes") {
		t.Error("expected Attributes section header")
	}
	if !strings.Contains(out, "Diagnostics") {
		t.Error("expected Diagnostics section header")
	}
	if !strings.Contains(out, "Raw Data") {
		t.Error("expected Raw Data section header")
	}
}

func TestRenderInfoSection_OnlyRaw(t *testing.T) {
	snap := core.UsageSnapshot{
		ProviderID: "test",
		AccountID:  "test",
		Timestamp:  time.Now(),
		Raw:        map[string]string{"key": "value"},
	}

	var sb strings.Builder
	renderInfoSection(&sb, snap, core.DefaultDashboardWidget(), 80)
	out := sb.String()

	if strings.Contains(out, "Attributes") {
		t.Error("should not show Attributes when empty")
	}
	if strings.Contains(out, "Diagnostics") {
		t.Error("should not show Diagnostics when empty")
	}
	if !strings.Contains(out, "Raw Data") {
		t.Error("expected Raw Data section")
	}
}

func TestRenderDetailContent_AtVariousWidths(t *testing.T) {
	snap := core.UsageSnapshot{
		ProviderID: "claude_code",
		AccountID:  "test",
		Timestamp:  time.Now(),
		Status:     core.StatusOK,
		Metrics: map[string]core.Metric{
			"plan_percent_used": {Used: core.Float64Ptr(60), Limit: core.Float64Ptr(100), Unit: "%"},
			"spend_limit":       {Used: core.Float64Ptr(45), Limit: core.Float64Ptr(100), Unit: "USD"},
		},
		ModelUsage: []core.ModelUsageRecord{
			{RawModelID: "opus-4", Canonical: "claude-opus-4", CostUSD: core.Float64Ptr(100), InputTokens: core.Float64Ptr(50000), OutputTokens: core.Float64Ptr(25000)},
		},
		DailySeries: map[string][]core.TimePoint{
			"cost": {
				{Date: "2026-02-18", Value: 5},
				{Date: "2026-02-19", Value: 8},
				{Date: "2026-02-20", Value: 12},
			},
		},
		Attributes: map[string]string{"plan": "pro"},
		Raw:        map[string]string{"version": "1.0"},
	}

	for _, width := range []int{40, 60, 80, 120} {
		out := RenderDetailContent(snap, time.Now(), width, 0.3, 0.1, 0, core.TimeWindowAll, false)
		if len(out) == 0 {
			t.Errorf("empty output at width %d", width)
		}
	}
}

func TestBuildDetailUsageSection_ProjectionRowsAlignWithinCard(t *testing.T) {
	now := time.Date(2026, 7, 2, 12, 0, 0, 0, time.UTC)
	snap := core.UsageSnapshot{
		ProviderID: "codex",
		Metrics: map[string]core.Metric{
			"rate_limit_primary":   makeUsageMetric(50, 100, "5h"),
			"rate_limit_secondary": makeUsageMetric(40, 100, "7d"),
		},
		Resets: map[string]time.Time{
			"rate_limit_primary":   now.Add(2*time.Hour + 30*time.Minute),
			"rate_limit_secondary": now.Add(3 * 24 * time.Hour),
		},
	}
	widget := core.DefaultDashboardWidget()
	widget.GaugePriority = []string{"rate_limit_primary", "rate_limit_secondary"}

	lines := buildDetailUsageSection(snap, widget, 80, 0.30, 0.15, false, now)
	if len(lines) != 4 {
		t.Fatalf("expected two gauges and two annotation rows, got %d: %v", len(lines), lines)
	}
	for i, line := range lines {
		if strings.Contains(line, "\n") {
			t.Fatalf("line %d contains an embedded newline: %q", i, line)
		}
	}

	var card strings.Builder
	renderDetailCard(&card, detailSection{title: "Usage", lines: lines}, 88)
	renderedCard := strings.TrimPrefix(strings.TrimSuffix(card.String(), "\n"), "\n")
	rendered := strings.Split(renderedCard, "\n")
	wantWidth := lipgloss.Width(rendered[0])
	for i, line := range rendered[1:] {
		if got := lipgloss.Width(line); got != wantWidth {
			t.Errorf("rendered card line %d width = %d, want %d: %q", i+1, got, wantWidth, line)
		}
	}
}
