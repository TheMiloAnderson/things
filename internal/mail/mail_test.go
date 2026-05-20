package mail

import (
	"strings"
	"testing"
)

func TestRenderVerification(t *testing.T) {
	html, text, err := RenderVerification("https://example.com/verify?t=abc", "Milo")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(html, "https://example.com/verify?t=abc") {
		t.Fatalf("html missing link: %s", html)
	}
	if !strings.Contains(text, "Milo") {
		t.Fatalf("text missing name: %s", text)
	}
}

func TestNoopSender(t *testing.T) {
	var n NoopSender
	if err := n.Send("a@b.com", "subj", "<p>x</p>", "x"); err != nil {
		t.Fatal(err)
	}
	if n.Last.To != "a@b.com" {
		t.Fatalf("got %+v", n.Last)
	}
}
