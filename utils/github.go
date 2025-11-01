package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type CommitAuthor struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type CommitDetails struct {
	Commit struct {
		Author CommitAuthor `json:"author"`
	} `json:"commit"`
	Author *struct {
		Login string `json:"login"`
	} `json:"author"`
}

func getCommitDetails(client *http.Client, owner, repo, path, token string) (*CommitDetails, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/commits?path=%s&per_page=1",
		url.PathEscape(owner), url.PathEscape(repo), url.PathEscape(path))

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", "token "+token)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get commit details: %s", resp.Status)
	}

	var commits []CommitDetails
	if err := json.NewDecoder(resp.Body).Decode(&commits); err != nil || len(commits) == 0 {
		return nil, fmt.Errorf("no commit details found")
	}

	return &commits[0], nil
}
