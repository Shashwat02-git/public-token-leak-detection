package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type SlackMessage struct {
	Blocks []Block `json:"blocks"`
}

type Block struct {
	Type     string        `json:"type"`
	Text     *TextObject   `json:"text,omitempty"`
	Elements []*TextObject `json:"elements,omitempty"`
}

type TextObject struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// Add this helper function at the top of the file
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Then update the SendSlackNotification function
func SendSlackNotification(leaks []Leak) error {
	webhookURL := os.Getenv("SLACK_WEBHOOK_URL")
	if webhookURL == "" {
		return fmt.Errorf("SLACK_WEBHOOK_URL environment variable not set")
	}

	const maxBlocks = 40 // Leave room for headers and footers
	var allErrors []string

	for i := 0; i < len(leaks); i += maxBlocks {
		end := min(i+maxBlocks, len(leaks))
		batch := leaks[i:end]

		var message SlackMessage
		headerText := fmt.Sprintf("*Security Scan: Token Leak Report (Batch %d/%d)*",
			(i/maxBlocks)+1,
			(len(leaks)+maxBlocks-1)/maxBlocks)

		message.Blocks = append(message.Blocks,
			Block{
				Type: "header",
				Text: &TextObject{Type: "plain_text", Text: "Token Leak Report"},
			},
			Block{
				Type: "section",
				Text: &TextObject{Type: "mrkdwn", Text: headerText},
			},
			Block{Type: "divider"},
		)

		for _, leak := range batch {
			token := leak.Token.TokenValue
			if len(token) > 8 {
				token = token[:4] + "..." + token[len(token)-4:]
			}

			leakText := fmt.Sprintf(
				"*Author:* %s\n*Domain:* %s\n*File:* `%s`\n*Location:* %s\n*Token Owner:* %s\n*Exposed Token:* `%s`\n*Remediation:* %s",
				leak.Metadata.Author,
				leak.Metadata.Domain,
				leak.Path,
				leak.Location,
				leak.Token.Owner,
				token,
				leak.Token.Remediation,
			)

			message.Blocks = append(message.Blocks,
				Block{
					Type: "section",
					Text: &TextObject{Type: "mrkdwn", Text: leakText},
				},
				Block{Type: "divider"},
			)
		}

		// Add batch completion message
		message.Blocks = append(message.Blocks, Block{
			Type: "context",
			Elements: []*TextObject{
				{Type: "mrkdwn", Text: fmt.Sprintf("Batch %d of %d completed", (i/maxBlocks)+1, (len(leaks)+maxBlocks-1)/maxBlocks)},
			},
		})

		payload, err := json.Marshal(message)
		if err != nil {
			log.Printf("Error marshaling message: %v", err)
			allErrors = append(allErrors, fmt.Sprintf("Batch %d: %v", (i/maxBlocks)+1, err))
			continue
		}

		req, err := http.NewRequest("POST", webhookURL, bytes.NewBuffer(payload))
		if err != nil {
			log.Printf("Error creating request for batch %d: %v", (i/maxBlocks)+1, err)
			allErrors = append(allErrors, fmt.Sprintf("Batch %d: %v", (i/maxBlocks)+1, err))
			continue
		}
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("Error sending batch %d: %v", (i/maxBlocks)+1, err)
			allErrors = append(allErrors, fmt.Sprintf("Batch %d: %v", (i/maxBlocks)+1, err))
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			log.Printf("Error response from Slack for batch %d: %s - %s",
				(i/maxBlocks)+1, resp.Status, string(body))
			allErrors = append(allErrors,
				fmt.Sprintf("Batch %d: %s - %s", (i/maxBlocks)+1, resp.Status, string(body)))
		} else {
			log.Printf("Successfully sent batch %d/%d", (i/maxBlocks)+1, (len(leaks)+maxBlocks-1)/maxBlocks)
		}

		// Respect Slack's rate limits
		time.Sleep(1 * time.Second)
	}

	if len(allErrors) > 0 {
		return fmt.Errorf("encountered %d errors while sending notifications: %v",
			len(allErrors), strings.Join(allErrors, "; "))
	}

	return nil
}
