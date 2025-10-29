package main

import (
	"bytes"
	"encoding/json"
	"fmt"
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

func main() {
	// Read the contents of the inventory file and decode into a slice of structs of Token type
	inventory, err := os.ReadFile("inventory.json")
	if err != nil {
		log.Fatalf("Failed to read inventory file: %v", err)
	}

	var tokens []Token

	decoder := json.NewDecoder(bytes.NewReader(inventory))
	if err := decoder.Decode(&tokens); err != nil {
		log.Fatalf("Failed to decode inventory file: %v", err)
	}

	// Read each file from the source_files directory
	dirPath := "./source_files"
	files, err := os.ReadDir(dirPath)
	if err != nil {
		log.Fatalf("Failed to get directory path: %v", err)
	}

	// Check for token leaks in each file
	for _, file := range files {
		fullPath := filepath.Join(dirPath, file.Name())
		content, err := os.ReadFile(fullPath)
		if err != nil {
			log.Fatalf("Failed to read file: %v", err)
		}
		for _, token := range tokens {
			if strings.Contains(string(content), token.TokenValue) {
				fmt.Printf("Token leak found in file: %s\n", fullPath)
			}
		}
	}
}
