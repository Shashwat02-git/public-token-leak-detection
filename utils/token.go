package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/joho/godotenv"
)

type Token struct {
	Provider    string `json:"provider"`
	TokenType   string `json:"token_type"`
	TokenValue  string `json:"token_value"`
	Owner       string `json:"owner"`
	Remediation string `json:"remediation"`
}

type Metadata struct {
	Author      string `json:"author"`
	Domain      string `json:"domain"`
	AuthorEmail string `json:"authorIP"`
}

type Leak struct {
	Path     string
	FilePath string
	Token    Token
	Metadata Metadata
	Location string
}

type FileJob struct {
	Path     string
	Content  []byte
	Metadata Metadata
}

type ProcessResult struct {
	Leaks []Leak
	Err   error
}

// Creates jobs from files
func CollectFileJobs(rootDir string) ([]FileJob, error) {
	var jobs []FileJob

	projectFolders, err := os.ReadDir(rootDir)
	if err != nil {
		return nil, err
	}

	for _, dir := range projectFolders {
		if !dir.IsDir() {
			continue
		}

		projectPath := filepath.Join(rootDir, dir.Name())
		metadataPath := filepath.Join(projectPath, "metadata.json")

		metaContent, err := os.ReadFile(metadataPath)
		if err != nil {
			return nil, err
		}

		var metadata Metadata
		decoder := json.NewDecoder(bytes.NewReader(metaContent))
		if err := decoder.Decode(&metadata); err != nil {
			return nil, err
		}

		walkErr := filepath.WalkDir(projectPath, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if d.IsDir() {
				return nil
			}

			if d.Name() == "metadata.json" {
				return nil
			}

			// Read the file content
			content, readErr := os.ReadFile(path)
			if readErr != nil {
				log.Printf("Warning: failed to read file %s. Error: %v", path, readErr)
				return nil
			}

			jobs = append(jobs, FileJob{
				Path:     path,
				Content:  content,
				Metadata: metadata,
			})

			return nil
		})

		if walkErr != nil {
			return nil, walkErr
		}

	}

	return jobs, nil
}

// If the numebr of tokens is less than number of workers, do parallel search for all tokens
func checkTokensSimple(content []byte, tokens []Token) []Token {
	var wg sync.WaitGroup
	matches := make(chan Token, len(tokens))

	for _, token := range tokens {
		wg.Add(1)
		go func(t Token) {
			defer wg.Done()
			if strings.Contains(string(content), t.TokenValue) {
				matches <- t
			}
		}(token)
	}

	go func() {
		wg.Wait()
		close(matches)
	}()

	var result []Token
	for token := range matches {
		result = append(result, token)
	}

	return result
}

// Check tokens with worker pool to prevent goroutine explosion
func checkTokens(content []byte, tokens []Token, numWorkers int) []Token {
	if len(tokens) == 0 {
		return nil
	}

	//Check all tokens in parallel if less than number of workers
	if len(tokens) <= numWorkers {
		return checkTokensSimple(content, tokens)
	}

	tokenChan := make(chan Token, len(tokens))
	resultChan := make(chan Token, len(tokens))

	for _, token := range tokens {
		tokenChan <- token
	}
	close(tokenChan)

	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for token := range tokenChan {
				if strings.Contains(string(content), token.TokenValue) {
					resultChan <- token
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	var matches []Token
	for token := range resultChan {
		matches = append(matches, token)
	}

	return matches
}

// Checks for tokens in the files from the jobs channel, uses a cache for location
func FileProcessingWorker(jobs <-chan FileJob, results chan<- ProcessResult, tokens []Token, wg *sync.WaitGroup, geoCache *sync.Map) {
	defer wg.Done()

	const numTokenWorkers = 20

	var leaks []Leak
	for job := range jobs {
		tokenMatches := checkTokens(job.Content, tokens, numTokenWorkers)

		//GeoLocation enrichment from cache
		for _, token := range tokenMatches {
			location := getCachedLocation(job.Metadata.AuthorEmail, geoCache)
			newLeak := Leak{
				Path:     job.Path,
				Token:    token,
				Metadata: job.Metadata,
				Location: location,
			}
			leaks = append(leaks, newLeak)
		}
	}

	results <- ProcessResult{Leaks: leaks}
}

// Returns location from cache, if not found, fetches it from ip-api.com
func getCachedLocation(ip string, cache *sync.Map) string {
	if cached, ok := cache.Load(ip); ok {
		return cached.(string)
	}

	location, err := GetGeoLocationFromIPs([]string{ip})
	if err != nil || location == "" {
		location = "Location not available"
	}

	cache.Store(ip, location)
	return location
}

func CheckForTokenLeaks(rootDir string, tokens []Token) ([]Leak, error) {
	jobs, err := CollectFileJobs(rootDir)
	if err != nil {
		return nil, err
	}

	if len(jobs) == 0 {
		return []Leak{}, nil
	}

	numWorkers := 10
	if len(jobs) < numWorkers {
		numWorkers = len(jobs)
	}

	jobsChan := make(chan FileJob, len(jobs))
	resultsChan := make(chan ProcessResult, numWorkers)

	for _, job := range jobs {
		jobsChan <- job
	}
	close(jobsChan)

	geoCache := sync.Map{}

	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go FileProcessingWorker(jobsChan, resultsChan, tokens, &wg, &geoCache)
	}

	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	var allLeaks []Leak
	for result := range resultsChan {
		if result.Err != nil {
			return nil, result.Err
		}
		allLeaks = append(allLeaks, result.Leaks...)
	}

	return allLeaks, nil
}

func SearchTokensOnGithub(tokens []Token) ([]Leak, error) {
	_ = godotenv.Load()
	githubToken := os.Getenv("GITHUB_TOKEN")
	if githubToken == "" {
		return nil, fmt.Errorf("GITHUB_TOKEN environment variable not set")
	}

	client := &http.Client{Timeout: 30 * time.Second}
	var leaks []Leak

	for _, token := range tokens {
		if token.TokenValue == "" {
			continue
		}

		// Build search URL
		reqURL := &url.URL{
			Scheme: "https",
			Host:   "api.github.com",
			Path:   "/search/code",
		}

		// Add query parameters
		q := reqURL.Query()
		q.Add("q", fmt.Sprintf("\"%s\"", token.TokenValue))
		reqURL.RawQuery = q.Encode()

		// Create request
		req, _ := http.NewRequest("GET", reqURL.String(), nil)
		req.Header.Set("Accept", "application/vnd.github+json")
		req.Header.Set("Authorization", "token "+githubToken)

		// Execute request
		resp, err := client.Do(req)
		if err != nil {
			continue
		}

		// Parse response
		var result struct {
			Items []struct {
				HTMLURL string `json:"html_url"`
				Path    string `json:"path"`
				Repo    struct {
					FullName string `json:"full_name"`
					HTMLURL  string `json:"html_url"`
				} `json:"repository"`
			} `json:"items"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			resp.Body.Close()
			continue
		}
		resp.Body.Close()

		// Process results
		for _, item := range result.Items {
			// Extract owner and repo from full name
			repoParts := strings.Split(item.Repo.FullName, "/")
			if len(repoParts) != 2 {
				continue
			}
			owner, repo := repoParts[0], repoParts[1]

			// Get commit details for author info
			var authorName, authorEmail, location string
			commit, err := getCommitDetails(client, owner, repo, item.Path, githubToken)
			if err == nil {
				authorName = commit.Commit.Author.Name
				authorEmail = commit.Commit.Author.Email

				// Get location from author's email domain if available
				if authorEmail != "" {
					emailParts := strings.Split(authorEmail, "@")
					if len(emailParts) == 2 {
						domain := emailParts[1]
						if ips, err := GetIPFromDomain(domain); err == nil {
							if loc, err := GetGeoLocationFromIPs(ips); err == nil {
								location = loc
							}
						}
					}
				}
			}

			// Create file URL by combining repo URL and file path
			fileURL := fmt.Sprintf("%s/blob/HEAD/%s", strings.TrimSuffix(item.Repo.HTMLURL, ".git"), item.Path)

			leaks = append(leaks, Leak{
				Path:     fileURL,   // Direct URL to the file
				FilePath: item.Path, // Just the file path
				Token:    token,
				Metadata: Metadata{
					Domain:      "github.com",
					Author:      authorName,
					AuthorEmail: authorEmail,
				},
				Location: location,
			})

			fmt.Printf("Found potential leak in %s by %s <%s> [%s]\n",
				fileURL, authorName, authorEmail, location)
		}

		time.Sleep(2 * time.Second) // Rate limiting
	}

	return leaks, nil
}
