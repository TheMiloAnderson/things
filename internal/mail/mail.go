package mail

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"text/template"
)

//go:embed templates/*
var templateFS embed.FS

type Sender interface {
	Send(to, subject, htmlBody, textBody string) error
}

type ResendSender struct {
	APIKey string
	From   string
	HTTP   *http.Client
}

func (r *ResendSender) Send(to, subject, htmlBody, textBody string) error {
	if r.APIKey == "" {
		return fmt.Errorf("mail: RESEND_API_KEY is not configured")
	}
	client := r.HTTP
	if client == nil {
		client = http.DefaultClient
	}
	body := map[string]any{
		"from":    r.From,
		"to":      []string{to},
		"subject": subject,
		"html":    htmlBody,
		"text":    textBody,
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, "https://api.resend.com/emails", bytes.NewReader(raw))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+r.APIKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("mail: resend API %s: %s", resp.Status, strings.TrimSpace(string(b)))
	}
	return nil
}

type SentMessage struct {
	To       string
	Subject  string
	HTMLBody string
	TextBody string
}

type NoopSender struct {
	Last SentMessage
}

func (n *NoopSender) Send(to, subject, htmlBody, textBody string) error {
	n.Last = SentMessage{To: to, Subject: subject, HTMLBody: htmlBody, TextBody: textBody}
	return nil
}

type emailTemplateData struct {
	Name string
	Link string
}

func renderTemplate(name string, data emailTemplateData) (string, error) {
	tmpl, err := template.ParseFS(templateFS, "templates/"+name)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func RenderVerification(link, name string) (html, text string, err error) {
	data := emailTemplateData{Name: name, Link: link}
	html, err = renderTemplate("verify.html", data)
	if err != nil {
		return "", "", err
	}
	text, err = renderTemplate("verify.txt", data)
	return html, text, err
}

func RenderReset(link, name string) (html, text string, err error) {
	data := emailTemplateData{Name: name, Link: link}
	html, err = renderTemplate("reset.html", data)
	if err != nil {
		return "", "", err
	}
	text, err = renderTemplate("reset.txt", data)
	return html, text, err
}

func SendVerification(sender Sender, to, name, link string) error {
	html, text, err := RenderVerification(link, name)
	if err != nil {
		return err
	}
	return sender.Send(to, "Verify your Things account", html, text)
}

func SendReset(sender Sender, to, name, link string) error {
	html, text, err := RenderReset(link, name)
	if err != nil {
		return err
	}
	return sender.Send(to, "Reset your Things password", html, text)
}
