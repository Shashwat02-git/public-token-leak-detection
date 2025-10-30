package utils

import (
	"bytes"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type Token struct {
	Provider    string `json:"provider"`
	TokenType   string `json:"token_type"`
	TokenValue  string `json:"token_value"`
	Owner       string `json:"owner"`
	Remediation string `json:"remediation"`
}

type Metadata struct {
	Author   string `json:"author"`
	Domain   string `json:"domain"`
	AuthorIP string `json:"authorIP"`
}

type Leak struct {
	Path     string
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
	metadataCache := make(map[string]Metadata)

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

		metadataCache[projectPath] = metadata

		files, err := os.ReadDir(projectPath)
		if err != nil {
			return nil, err
		}

		for _, file := range files {
			if file.IsDir() || file.Name() == "metadata.json" {
				continue
			}

			fullPath := filepath.Join(projectPath, file.Name())
			content, err := os.ReadFile(fullPath)
			if err != nil {
				log.Printf("Warning: failed to read file %s. Error: %v", fullPath, err)
				continue
			}

			jobs = append(jobs, FileJob{
				Path:     fullPath,
				Content:  content,
				Metadata: metadata,
			})
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
			location := getCachedLocation(job.Metadata.AuthorIP, geoCache)
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

	location, err := GetGeoLocationFromIP(ip)
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
