package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"public-leak-detection/utils"
	"time"
)

func main() {
	start := time.Now()
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

	dirPath := "./source_files"

	// Check for token leaks in each file
	leaks, err := utils.CheckForTokenLeaks(dirPath, tokens)
	if err != nil {
		log.Fatal(err)
	}

	emailList := make(map[string][]utils.Leak)

	for _, leak := range leaks {
		ownerEmail := emailList[leak.Token.Owner]
		ownerEmail = append(ownerEmail, leak)
		emailList[leak.Token.Owner] = ownerEmail
	}

	for owner, details := range emailList {
		utils.SendEmail(owner, "WARNING: You have token leaks", details)
	}
	utils.SendSlackNotification(leaks)
	duration := time.Since(start)
	fmt.Printf("%s\n", duration)
}
