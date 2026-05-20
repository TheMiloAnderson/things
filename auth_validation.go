package main

import (
	"regexp"
	"strings"
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"
)

const (
	minPasswordLen = 12
	bcryptCost     = 12
)

var emailRegex = regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}$`)

func normalizeEmail(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

func validEmail(email string) bool {
	return emailRegex.MatchString(email)
}

func validUsername(name string) bool {
	if name == "" || utf8.RuneCountInString(name) > 255 {
		return false
	}
	return !strings.ContainsAny(name, "\x00\r\n")
}

func validPassword(pass string) bool {
	return utf8.RuneCountInString(pass) >= minPasswordLen
}

func hashPassword(pass string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(pass), bcryptCost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

type AuthViewModel struct {
	Flash      string
	Error      string
	Message    string
	Email      string
	Username   string
	Token      string
	ShowResend bool
}
