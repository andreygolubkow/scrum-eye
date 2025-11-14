package azureboards

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"dev-digest/common"
)

type Module struct{}

func (Module) Name() string                    { return "Azure Boards" }
func (Module) Enabled(cfg *common.Config) bool { return cfg.Azure != nil && cfg.Azure.Boards != nil }

func init() { common.Register(Module{}) }

func (Module) Run(ctx context.Context, cfg *common.Config) (*common.Report, error) {
	az := cfg.Azure
	bcfg := az.Boards
	if az.PAT == "" {
		return &common.Report{Title: "Azure Boards", Summary: "PAT not provided; skipping API calls"}, nil
	}

	client := &http.Client{Timeout: 15 * time.Second}
	org := strings.TrimRight(az.Organization, "/")

	// 1) Current iteration
	iter, err := getCurrentIteration(ctx, client, org, bcfg.Project, bcfg.Team, az.PAT)
	if err != nil {
		return nil, fmt.Errorf("get current iteration: %w", err)
	}

	// 2) Query Feature work items in iteration using WIQL
	ids, err := queryFeatureIDs(ctx, client, org, bcfg.Project, bcfg.Team, az.PAT, iter.Path)
	if err != nil {
		return nil, fmt.Errorf("wiql query: %w", err)
	}

	// 3) Batch get details
	items, err := getWorkItemDetails(ctx, client, org, bcfg.Project, az.PAT, ids)
	if err != nil {
		return nil, fmt.Errorf("get work items: %w", err)
	}

	// Put IDs into shared Facts
	facts := common.GetFacts(ctx)
	facts.FeatureIDs = append(facts.FeatureIDs, ids...)

	// Build table
	rows := [][]string{}
	done, total := 0.0, 0.0
	for _, it := range items {
		title := it.Fields.Title
		state := it.Fields.State
		rem := it.Fields.RemainingWork
		comp := it.Fields.CompletedWork
		est := it.Fields.OriginalEstimate
		// Progress: if estimate exists, use (completed/estimate)
		var prog float64
		if est > 0 {
			prog = (comp / est) * 100.0
			total += est
			done += comp
		}
		rows = append(rows, []string{
			fmt.Sprintf("%d", it.ID), title, state,
			humanHours(est), humanHours(comp), humanHours(rem),
			fmt.Sprintf("%.0f%%", prog),
		})
	}

	summary := fmt.Sprintf("Sprint: %s (%s – %s). Features: %d. Progress: %.0f%%",
		iter.Name, iter.Attributes.StartDate.Format("2006-01-02"), iter.Attributes.FinishDate.Format("2006-01-02"), len(items), percent(done, total))

	table := &common.Table{Headers: []string{"ID", "Title", "State", "Est", "Done", "Rem", "Prog"}, Rows: rows}
	rep := &common.Report{Title: "Azure Boards", Summary: summary, Sections: []common.Section{{Header: "Features in Current Sprint", Table: table}}}
	if rep.Meta == nil {
		rep.Meta = map[string]any{}
	}
	rep.Meta["feature_ids"] = ids
	rep.Meta["iteration_path"] = iter.Path
	rep.Meta["iteration_name"] = iter.Name
	return rep, nil
}

type iteration struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Path       string `json:"path"`
	Attributes struct {
		StartDate  time.Time `json:"startDate"`
		FinishDate time.Time `json:"finishDate"`
		TimeFrame  string    `json:"timeFrame"`
	} `json:"attributes"`
}

type wiRef struct {
	ID int `json:"id"`
}

type workItem struct {
	ID     int `json:"id"`
	Fields struct {
		Title            string  `json:"System.Title"`
		State            string  `json:"System.State"`
		RemainingWork    float64 `json:"Microsoft.VSTS.Scheduling.RemainingWork"`
		CompletedWork    float64 `json:"Microsoft.VSTS.Scheduling.CompletedWork"`
		OriginalEstimate float64 `json:"Microsoft.VSTS.Scheduling.OriginalEstimate"`
	} `json:"fields"`
}

func getCurrentIteration(ctx context.Context, c *http.Client, org, project, team, pat string) (*iteration, error) {
	u := fmt.Sprintf("%s/%s/%s/_apis/work/teamsettings/iterations?timeframe=current&api-version=7.0", org, url.PathEscape(project), url.PathEscape(team))
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	addPAT(req, pat)
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("http %d", resp.StatusCode)
	}
	var payload struct {
		Value []iteration `json:"value"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}
	if len(payload.Value) == 0 {
		return nil, fmt.Errorf("no current iteration")
	}
	it := payload.Value[0]
	return &it, nil
}

func queryFeatureIDs(ctx context.Context, c *http.Client, org, project, team, pat, iterationPath string) ([]int, error) {
	wiql := fmt.Sprintf("Select [System.Id] From WorkItems Where [System.TeamProject]='%s' and [System.WorkItemType] in ('Feature') and [System.IterationPath] Under '%s' order by [System.Id]", project, iterationPath)
	u := fmt.Sprintf("%s/%s/%s/_apis/wit/wiql?api-version=7.0", org, url.PathEscape(project), url.PathEscape(team))
	body, _ := json.Marshal(map[string]string{"query": wiql})
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(body))
	addPAT(req, pat)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("http %d", resp.StatusCode)
	}
	var payload struct {
		WorkItems []wiRef `json:"workItems"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}
	ids := make([]int, 0, len(payload.WorkItems))
	for _, w := range payload.WorkItems {
		ids = append(ids, w.ID)
	}
	return ids, nil
}

func getWorkItemDetails(ctx context.Context, c *http.Client, org, project, pat string, ids []int) ([]workItem, error) {
	if len(ids) == 0 {
		return []workItem{}, nil
	}
	// Batch fetch
	parts := make([]string, len(ids))
	for i, id := range ids {
		parts[i] = fmt.Sprintf("%d", id)
	}
	u := fmt.Sprintf("%s/%s/_apis/wit/workitems?ids=%s&fields=%s&api-version=7.0", org, url.PathEscape(project), strings.Join(parts, ","), url.QueryEscape(strings.Join([]string{
		"System.Title", "System.State", "Microsoft.VSTS.Scheduling.RemainingWork", "Microsoft.VSTS.Scheduling.CompletedWork", "Microsoft.VSTS.Scheduling.OriginalEstimate",
	}, ",")))
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	addPAT(req, pat)
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("http %d", resp.StatusCode)
	}
	var payload struct {
		Value []workItem `json:"value"`
	}
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&payload); err != nil {
		// Some API versions return an array directly
		var arr []workItem
		if err2 := dec.Decode(&arr); err2 == nil {
			return arr, nil
		}
		return nil, err
	}
	return payload.Value, nil
}

func addPAT(req *http.Request, pat string) {
	tok := base64.StdEncoding.EncodeToString([]byte(":" + pat))
	req.Header.Set("Authorization", "Basic "+tok)
	req.Header.Set("Accept", "application/json")
}

func humanHours(v float64) string {
	if v == 0 {
		return "-"
	}
	return fmt.Sprintf("%.1fh", v)
}

func percent(done, total float64) float64 {
	if total <= 0 {
		return 0
	}
	return (done / total) * 100.0
}
