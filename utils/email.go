package utils

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/smtp"
	"os"

	"github.com/joho/godotenv"
)

// Generate email body from HTML template
func GenerateHTMLReport(leaks []Leak) (string, error) {
	// Redact sensitive data for the report
	renderData := make([]Leak, len(leaks))
	copy(renderData, leaks)
	for i := range renderData {
		token := renderData[i].Token.TokenValue
		if len(token) > 8 {
			renderData[i].Token.TokenValue = token[:4] + "..." + token[len(token)-4:]
		}
	}

	// Parse the HTML template file
	tmpl, err := template.ParseFiles("./templates/report.html")
	if err != nil {
		return "", fmt.Errorf("error parsing template: %w", err)
	}

	// Create a buffer to hold the generated HTML
	var body bytes.Buffer

	// Execute the template and write the output to the buffer
	err = tmpl.Execute(&body, renderData)
	if err != nil {
		return "", fmt.Errorf("error executing template: %w", err)
	}

	// Return the buffer's content as a string
	return body.String(), nil
}

func SendEmail(to string, subject string, leaks []Leak) error {
	// Try to load .env but don't fail if it's not there (for production)
	_ = godotenv.Load()

	// Sending email through gmail smtp
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"
	sender := os.Getenv("SENDER")
	password := os.Getenv("PASSWORD")

	recipient := to
	auth := smtp.PlainAuth("", sender, password, smtpHost)

	fromHeader := fmt.Sprintf("From: %s\n", sender)
	toHeader := fmt.Sprintf("To: %s\n", recipient)
	subjectHeader := fmt.Sprintf("Subject: %s\n", subject)
	mime := "MIME-version: 1.0;\n"
	contentType := "Content-Type: text/html; charset=\"UTF-8\";\n"

	body, err := GenerateHTMLReport(leaks)
	if err != nil {
		return err
	}

	var message []byte
	message = fmt.Appendf(message, "%s%s%s%s%s\n%s", fromHeader, toHeader, subjectHeader, mime, contentType, body)

	err = smtp.SendMail(
		fmt.Sprintf("%s:%s", smtpHost, smtpPort),
		auth,
		sender,
		[]string{recipient},
		message,
	)

	if err != nil {

		return err
	}

	log.Printf("Email sent successfully to %s!", to)
	return nil
}
