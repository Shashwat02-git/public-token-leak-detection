package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

type SlackMessage struct {
	Blocks []Block `json:"blocks"`
}

type Block struct {
	Type string      `json:"type"`
	Text *TextObject `json:"text,omitempty"`
}

type TextObject struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func SendSlackNotification(leaks []Leak) error {
	err := godotenv.Load()
	if err != nil {
		return err
	}

	webhookURL := os.Getenv("SLACK_WEBHOOK_URL")

	var message SlackMessage
	headerText := ":warning: *Security Scan: Token Leak Report* :warning:"
	message.Blocks = append(message.Blocks, Block{
		Type: "header",
		Text: &TextObject{Type: "plain_text", Text: "Token Leak Report"},
	})
	message.Blocks = append(message.Blocks, Block{
		Type: "section",
		Text: &TextObject{Type: "mrkdwn", Text: headerText},
	})
	message.Blocks = append(message.Blocks, Block{Type: "divider"})

	if len(leaks) == 0 {
		message.Blocks = append(message.Blocks, Block{
			Type: "section",
			Text: &TextObject{Type: "mrkdwn", Text: "*No token leaks were found.*"},
		})
	} else {
		for _, leak := range leaks {
			token := leak.Token.TokenValue
			if len(token) > 8 {
				token = token[:4] + "..." + token[len(token)-4:]
			}

			leakText := fmt.Sprintf(
				"*Author:* %s\n*Domain:* %s\n*File:* `%s`\n*Token Owner:* %s\n*Exposed Token:* `%s`",
				leak.Metadata.Author,
				leak.Metadata.Domain,
				leak.Path,
				leak.Token.Owner,
				token,
			)

			message.Blocks = append(message.Blocks, Block{
				Type: "section",
				Text: &TextObject{Type: "mrkdwn", Text: leakText},
			})
			message.Blocks = append(message.Blocks, Block{Type: "divider"})
		}
	}

	var payloadBuffer bytes.Buffer
	encoder := json.NewEncoder(&payloadBuffer)
	if err := encoder.Encode(message); err != nil {
		return err
	}

	req, err := http.NewRequest("POST", webhookURL, &payloadBuffer)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err
	}

	log.Println("Slack notification sent successfully!")
	return nil
}
