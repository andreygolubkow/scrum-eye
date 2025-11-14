package teamcity

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"dev-digest/common"
)

// BuildSummary holds minimal details about a TeamCity build configuration result
type BuildSummary struct {
	BuildTypeID string
	Branch      string
	Status      string
	TestsTotal  int
	TestsPassed int
	TestsFailed int
}

type Module struct{}

func (Module) Name() string                    { return "TeamCity" }
func (Module) Enabled(cfg *common.Config) bool { return cfg.TeamCity != nil }

func init() { common.Register(Module{}) }

func (Module) Run(ctx context.Context, cfg *common.Config) (*common.Report, error) {
	tc := cfg.TeamCity
	if tc == nil {
		return nil, fmt.Errorf("teamcity config missing")
	}
	if tc.Token == "" {
		return &common.Report{Title: "TeamCity", Summary: "token not provided; skipping API calls"}, nil
	}

	client := &http.Client{}

	var rows [][]string
	summaries := []BuildSummary{}

	for _, bt := range tc.Builds {
		sum, err := fetchLatestBuild(ctx, client, tc.BaseURL, tc.Token, bt, tc.Branch)
		if err != nil {
			rows = append(rows, []string{bt, tc.Branch, fmt.Sprintf("error: %v", err), "-", "-", "-"})
			continue
		}
		summaries = append(summaries, *sum)
		rows = append(rows, []string{sum.BuildTypeID, sum.Branch, sum.Status,
			fmt.Sprintf("%d", sum.TestsTotal), fmt.Sprintf("%d", sum.TestsPassed), fmt.Sprintf("%d", sum.TestsFailed)})
	}

	table := &common.Table{Headers: []string{"BuildType", "Branch", "Status", "Tests", "Passed", "Failed"}, Rows: rows}
	rep := &common.Report{Title: "TeamCity", Summary: "Latest builds and tests", Sections: []common.Section{{Header: "Builds", Table: table}}}
	return rep, nil
}

// TeamCity minimal responses
type teamcityBuild struct {
	ID         int    `json:"id"`
	Status     string `json:"status"`
	BranchName string `json:"branchName"`
}
type testOccurrences struct {
	Count int `json:"count"`
}

func fetchLatestBuild(ctx context.Context, httpClient *http.Client, baseURL, token, buildTypeID, branch string) (*BuildSummary, error) {
	// API: GET /app/rest/builds?locator=buildType:<id>,branch:<branch>,status:any,count:1
	q := url.QueryEscape(fmt.Sprintf("buildType:%s,branch:%s,status:any,count:1", buildTypeID, branch))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimRight(baseURL, "/")+"/app/rest/builds?locator="+q, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("http %d", resp.StatusCode)
	}

	var payload struct {
		Count int             `json:"count"`
		Build []teamcityBuild `json:"build"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}
	if len(payload.Build) == 0 {
		return nil, fmt.Errorf("no builds found")
	}
	b := payload.Build[0]

	// Fetch test stats for the build: /app/rest/testOccurrences?locator=build:(id:<id>),status:SUCCESS
	tot, pass, fail := 0, 0, 0
	// total
	c, err := fetchTestCount(ctx, httpClient, baseURL, token, b.ID, "ANY")
	if err == nil {
		tot = c
	}
	c, err = fetchTestCount(ctx, httpClient, baseURL, token, b.ID, "SUCCESS")
	if err == nil {
		pass = c
	}
	c, err = fetchTestCount(ctx, httpClient, baseURL, token, b.ID, "FAILURE")
	if err == nil {
		fail = c
	}

	return &BuildSummary{BuildTypeID: buildTypeID, Branch: b.BranchName, Status: b.Status, TestsTotal: tot, TestsPassed: pass, TestsFailed: fail}, nil
}

func fetchTestCount(ctx context.Context, httpClient *http.Client, baseURL, token string, buildID int, status string) (int, error) {
	loc := fmt.Sprintf("build:(id:%d)", buildID)
	if status != "ANY" {
		loc += ",status:" + status
	}
	q := url.QueryEscape(loc)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimRight(baseURL, "/")+"/app/rest/testOccurrences?locator="+q, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return 0, fmt.Errorf("http %d", resp.StatusCode)
	}
	var t testOccurrences
	if err := json.NewDecoder(resp.Body).Decode(&t); err != nil {
		return 0, err
	}
	return t.Count, nil
}
