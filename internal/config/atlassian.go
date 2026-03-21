package config

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"
)

type tenantInfo struct {
	CloudID string `json:"cloudId"`
}

func ResolveCloudID(siteURL string) (string, error) {
	siteURL = strings.TrimSpace(siteURL)
	if siteURL == "" {
		return "", fmt.Errorf("atlassian site url is required")
	}

	parsedURL, err := url.Parse(siteURL)
	if err != nil {
		return "", err
	}

	if parsedURL.Scheme == "" {
		parsedURL.Scheme = "https"
	}

	if parsedURL.Host == "" {
		parsedURL.Host = parsedURL.Path
		parsedURL.Path = ""
	}

	parsedURL.Path = path.Join(parsedURL.Path, "_edge", "tenant_info")
	parsedURL.RawQuery = ""
	parsedURL.Fragment = ""

	resp, err := http.Get(parsedURL.String())
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("tenant info lookup failed with status %s", resp.Status)
	}

	var info tenantInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return "", err
	}

	if info.CloudID == "" {
		return "", fmt.Errorf("tenant info response did not include cloudId")
	}

	return info.CloudID, nil
}
