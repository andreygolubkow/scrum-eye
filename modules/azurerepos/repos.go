package azurerepos

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"dev-digest/common"
)

type Module struct{}

func (Module) Name() string                    { return "Azure Repos" }
func (Module) Enabled(cfg *common.Config) bool { return cfg.Azure != nil && cfg.Azure.Repos != nil }

func init() { common.Register(Module{}) }

func (Module) Run(ctx context.Context, cfg *common.Config) (*common.Report, error) {
	az := cfg.Azure
	if az.PAT == "" {
		return &common.Report{Title: "Azure Repos", Summary: "PAT not provided; skipping API calls"}, nil
	}

	client := &http.Client{Timeout: 15 * time.Second}
	org := strings.TrimRight(az.Organization, "/")

	// Get feature IDs from Facts (populated by Boards)
	facts := common.GetFacts(ctx)
	featureIDs := facts.FeatureIDs

	// 1) list repos
	repos, err := listRepos(ctx, client, org, az.PAT)
	if err != nil {
		return nil, err
	}

	// 2) find branches per repo matching feature IDs
	rows := [][]string{}
	matches := 0
	for _, r := range repos {
		if len(featureIDs) == 0 {
			// Just list default branch
			rows = append(rows, []string{r.Name, r.DefaultBranch, "-"})
			continue
		}
		branches, err := listBranches(ctx, client, org, r.Project.Name, r.ID, az.PAT)
		if err != nil {
			rows = append(rows, []string{r.Name, "error", err.Error()})
			continue
		}
		// prepare map
		bnames := make([]string, len(branches))
		for i, b := range branches {
			bnames[i] = b.Name
		}
		sort.Strings(bnames)
		found := []string{}
		for _, id := range featureIDs {
			needle := fmt.Sprintf("%d", id)
			for _, br := range bnames {
				if strings.Contains(strings.ToLower(br), strings.ToLower(needle)) {
					found = append(found, br)
				}
			}
		}
		if len(found) == 0 {
			rows = append(rows, []string{r.Name, r.DefaultBranch, "no feature branches"})
		} else {
			matches += len(found)
			rows = append(rows, []string{r.Name, r.DefaultBranch, strings.Join(found, ", ")})
		}
	}

	summary := fmt.Sprintf("Repos: %d. Feature branch matches: %d", len(repos), matches)
	table := &common.Table{Headers: []string{"Repository", "Default", "Matching branches"}, Rows: rows}
	rep := &common.Report{Title: "Azure Repos", Summary: summary, Sections: []common.Section{{Header: "Branches by repository", Table: table}}}
	return rep, nil
}

// Models and helpers

type repo struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	DefaultBranch string `json:"defaultBranch"`
	Project       struct {
		Name string `json:"name"`
	} `json:"project"`
}

type branch struct {
	Name string `json:"name"`
}

type listReposPayload struct {
	Value []repo `json:"value"`
}

type listBranchesPayload struct {
	Value []branch `json:"value"`
}

func listRepos(ctx context.Context, c *http.Client, org, pat string) ([]repo, error) {
	u := fmt.Sprintf("%s/_apis/git/repositories?api-version=7.0", org)
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
	var payload listReposPayload
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}
	return payload.Value, nil
}

func listBranches(ctx context.Context, c *http.Client, org, project, repoID, pat string) ([]branch, error) {
	u := fmt.Sprintf("%s/%s/_apis/git/repositories/%s/refs?filter=refs/heads/&api-version=7.0", org, url.PathEscape(project), url.PathEscape(repoID))
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
	var payload listBranchesPayload
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}
	// branches have Name like refs/heads/feature/1234-some-title -> return last part
	out := make([]branch, 0, len(payload.Value))
	for _, b := range payload.Value {
		name := b.Name
		if strings.HasPrefix(name, "refs/heads/") {
			name = strings.TrimPrefix(name, "refs/heads/")
		}
		out = append(out, branch{Name: name})
	}
	return out, nil
}

func addPAT(req *http.Request, pat string) {
	tok := base64.StdEncoding.EncodeToString([]byte(":" + pat))
	req.Header.Set("Authorization", "Basic "+tok)
	req.Header.Set("Accept", "application/json")
}
