package utils

import (
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

type Leak struct {
	Path  string
	Token Token
}

func CheckForTokenLeaks(dirPath string, files []os.DirEntry, tokens []Token) []Leak {
	var leaks []Leak
	for _, file := range files {
		fullPath := filepath.Join(dirPath, file.Name())
		content, err := os.ReadFile(fullPath)
		if err != nil {
			log.Fatalf("Failed to read file: %v", err)
		}
		for _, token := range tokens {
			if strings.Contains(string(content), token.TokenValue) {
				leaks = append(leaks, Leak{fullPath, token})
			}
		}
	}
	return leaks
}
