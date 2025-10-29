package main

import (
	"bytes"
	"encoding/json"
	"log"
	"os"
	"public-leak-detection/utils"
)

func main() {
	// Read the contents of the inventory file and decode into a slice of structs of Token type
	inventory, err := os.ReadFile("inventory.json")
	if err != nil {
		log.Fatalf("Failed to read inventory file: %v", err)
	}

	var tokens []utils.Token

	decoder := json.NewDecoder(bytes.NewReader(inventory))
	if err := decoder.Decode(&tokens); err != nil {
		log.Fatalf("Failed to decode inventory file: %v", err)
	}

	// Read each file from the source_files directory
	dirPath := "./source_files"
	files := utils.GetFiles(dirPath)

	// Check for token leaks in each file
	leaks := utils.CheckForTokenLeaks(dirPath, files, tokens)

	emailList := make(map[string][]utils.Leak)

	for _, leak := range leaks {
		ownerEmail := emailList[leak.Token.Owner]
		ownerEmail = append(ownerEmail, leak)
		emailList[leak.Token.Owner] = ownerEmail
	}

	for owner, details := range emailList {
		utils.SendEmail(owner, "WARNING: You have token leaks", details)
	}
}
