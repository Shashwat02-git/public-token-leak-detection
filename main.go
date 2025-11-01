package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"public-leak-detection/utils"
	"sync"
	"time"

	"github.com/joho/godotenv"
)

type EmailJob struct {
	Owner string
	Leak  []utils.Leak
}

func worker(wg *sync.WaitGroup, jobs <-chan EmailJob) {
	defer wg.Done()
	for job := range jobs {
		utils.SendEmail(job.Owner, "Token Leak Report", job.Leak)
	}
}

func runFullScan(w http.ResponseWriter) {
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

	// dirPath := "./source_files"

	// Check for token leaks in each file
	leaks, err := utils.SearchTokensOnGithub(tokens)
	if err != nil {
		log.Fatal(err)
	}

	emailList := make(map[string][]utils.Leak)
	for _, leak := range leaks {
		ownerEmail := emailList[leak.Token.Owner]
		ownerEmail = append(ownerEmail, leak)
		emailList[leak.Token.Owner] = ownerEmail
	}
	var wg sync.WaitGroup

	// Create worker pool to send emails
	const numWorkers = 5
	jobsChan := make(chan EmailJob, len(emailList))
	for owner, job := range emailList {
		jobsChan <- EmailJob{Owner: owner, Leak: job}
	}
	close(jobsChan)

	wg.Add(numWorkers)
	for w := 1; w <= numWorkers; w++ {
		go worker(&wg, jobsChan)
	}

	// Send notification to slack
	wg.Add(1)
	go func(l []utils.Leak) {
		utils.SendSlackNotification(l)
		defer wg.Done()
	}(leaks)

	wg.Wait()

	duration := time.Since(start)
	fmt.Printf("Time Taken: %s\n", duration)
	fmt.Fprintf(w, "Time Taken: %s\n", duration)
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Token Leak Detection Service Is Running!\n")
	})

	http.HandleFunc("/api/check", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Scan started\n")
		runFullScan(w)
	})

	_ = godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Detection system server starting on port %s", port)
	err := http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
	if err != nil {
		log.Fatal(err)
	}
}
