package azureboards

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"scrum-eye/internal/config"
	"strings"
	"time"
)

type Client struct {
	organization string
	project      string
	token        string
	team         string
	area         string

	baseRestUrl  string
	baseOdataUrl string
	httpClient   *http.Client
}

func NewClient(azureCfg config.AzureDevOpsTeam) *Client {
	baseRestUrl := "https://dev.azure.com/" + azureCfg.Organisation
	baseOdataUrl := "https://analytics.dev.azure.com/" + azureCfg.Organisation

	return &Client{
		organization: azureCfg.Organisation,
		token:        azureCfg.Token,
		team:         azureCfg.TeamId,
		area:         azureCfg.AreaPath,
		project:      azureCfg.ProjectId,
		baseRestUrl:  baseRestUrl,
		baseOdataUrl: baseOdataUrl,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

func (c *Client) GetCurrentIterationPath(ctx context.Context) (string, error) {
	iter, err := c.getCurrentIteration(ctx)
	if err != nil {
		return "", err
	}
	return iter.Path, nil
}

func (c *Client) getCurrentIteration(ctx context.Context) (*Iteration, error) {
	path := fmt.Sprintf("/%s/%s/_apis/work/teamsettings/iterations", c.project, c.team)

	query := url.Values{}
	query.Set("api-version", "7.1")
	query.Set("$timeframe", "current")

	var resp iterationsListResponse
	if err := c.doRestRequest(ctx, http.MethodGet, path, query, &resp); err != nil {
		return nil, fmt.Errorf("getCurrentIteration: %w", err)
	}

	if resp.Count == 0 || len(resp.Value) == 0 {
		return nil, fmt.Errorf("no current iteration found for team %q", c.team)
	}

	// Azure DevOps обычно возвращает один current-iteration
	return &resp.Value[0], nil
}

func (c *Client) doRestRequest(ctx context.Context, method, path string, query url.Values, out any) error {
	return c.doRequest(ctx, method, c.baseRestUrl, path, query, out)
}

func (c *Client) doRequest(ctx context.Context, method, baseUrl, path string, query url.Values, out any) error {
	u, err := url.Parse(baseUrl)
	if err != nil {
		return err
	}
	u.Path = strings.TrimRight(u.Path, "/") + path
	if len(query) > 0 {
		u.RawQuery = query.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), nil)
	if err != nil {
		return err
	}

	// PAT авторизация: Basic base64(":"+PAT)
	token := ":" + c.token
	encoded := base64.StdEncoding.EncodeToString([]byte(token))
	req.Header.Set("Authorization", "Basic "+encoded)
	req.Header.Set("Accept", "application/json")
	fmt.Printf("req: %+v", u.String())
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("azure devops api returned %s for %s", resp.Status, u.String())
	}

	if out == nil {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(out)
}
