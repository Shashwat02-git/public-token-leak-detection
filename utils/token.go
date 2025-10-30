package utils

import (
	"bytes"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type Token struct {
	Provider   string `json:"provider"`
	TokenType  string `json:"token_type"`
	TokenValue string `json:"token_value"`
	Owner      string `json:"owner"`
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

func CheckForTokenLeaks(rootDir string, tokens []Token) ([]Leak, error) {
	var leaks []Leak

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
		err = decoder.Decode(&metadata)

		if err != nil {
			return nil, err
		}

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

			for _, token := range tokens {
				if strings.Contains(string(content), token.TokenValue) {
					location, err := GetGeoLocationFromIP(metadata.AuthorIP)
					if err != nil && location == "" {
						location = "Location not available"
					}

					newLeak := Leak{
						Path:     fullPath,
						Token:    token,
						Metadata: metadata,
						Location: location,
					}
					leaks = append(leaks, newLeak)
				}
			}
		}
	}

	return leaks, nil
}
