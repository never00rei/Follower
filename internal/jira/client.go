package jira

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/never00rei/Follower/internal/config"
)

const apiBaseURL = "https://api.atlassian.com"

type Client struct {
	httpClient *http.Client
	email      string
	apiKey     string
	cloudID    string
}

func NewClient(conf config.AtlassianConfig) (*Client, error) {
	if strings.TrimSpace(conf.AtlassianCloudID) == "" {
		return nil, fmt.Errorf("missing Atlassian cloud ID")
	}

	if strings.TrimSpace(conf.UserEmail) == "" {
		return nil, fmt.Errorf("missing Atlassian user email")
	}

	if strings.TrimSpace(conf.AtlassianAPIKey) == "" {
		return nil, fmt.Errorf("missing Atlassian API key")
	}

	return &Client{
		httpClient: &http.Client{Timeout: 15 * time.Second},
		email:      conf.UserEmail,
		apiKey:     conf.AtlassianAPIKey,
		cloudID:    conf.AtlassianCloudID,
	}, nil
}

func (c *Client) GetIssue(ctx context.Context, issueID string) (*Issue, error) {
	var issue Issue
	err := c.doJSON(ctx, http.MethodGet, "/rest/api/3/issue/"+issueID+"?fields=summary,description,status,issuetype,project,assignee,reporter,priority,labels,created,updated", nil, &issue)
	if err != nil {
		return nil, err
	}

	return &issue, nil
}

func (c *Client) doJSON(ctx context.Context, method, apiPath string, body io.Reader, out any) error {
	req, err := http.NewRequestWithContext(ctx, method, apiBaseURL+"/ex/jira/"+c.cloudID+apiPath, body)
	if err != nil {
		return err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(c.email, c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("jira request failed with status %s: %s", resp.Status, strings.TrimSpace(string(respBody)))
	}

	if out == nil {
		return nil
	}

	return json.NewDecoder(resp.Body).Decode(out)
}
