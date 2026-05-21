package main

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"things/internal/mail"
	"things/internal/models"
)

const (
	verifyTokenTTL = 24 * time.Hour
	resetTokenTTL  = time.Hour
)

func (a *App) authPage(w http.ResponseWriter, r *http.Request, tmpl string, vm AuthViewModel) {
	a.render(w, r, tmpl, PageData{
		IsAuthenticated: false,
		Data:            vm,
	})
}

func (a *App) signupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		a.authPage(w, r, "signup.html", AuthViewModel{})
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if a.AuthLimiter != nil && a.AuthLimiter.blockedSignup(r) {
		http.Error(w, "Too many signup attempts. Try again in a few minutes.", http.StatusTooManyRequests)
		return
	}

	name := strings.TrimSpace(r.FormValue("username"))
	email := normalizeEmail(r.FormValue("email"))
	pass := r.FormValue("password")
	confirm := r.FormValue("password_confirm")

	vm := AuthViewModel{Username: name, Email: email}

	if !validUsername(name) {
		if a.AuthLimiter != nil {
			a.AuthLimiter.recordSignupFailure(r)
		}
		vm.Error = "Please enter a valid username."
		a.authPage(w, r, "signup.html", vm)
		return
	}
	if !validEmail(email) {
		if a.AuthLimiter != nil {
			a.AuthLimiter.recordSignupFailure(r)
		}
		vm.Error = "Please enter a valid email address."
		a.authPage(w, r, "signup.html", vm)
		return
	}
	if !validPassword(pass) {
		if a.AuthLimiter != nil {
			a.AuthLimiter.recordSignupFailure(r)
		}
		vm.Error = "Password must be at least 12 characters."
		a.authPage(w, r, "signup.html", vm)
		return
	}
	if pass != confirm {
		if a.AuthLimiter != nil {
			a.AuthLimiter.recordSignupFailure(r)
		}
		vm.Error = "Passwords do not match."
		a.authPage(w, r, "signup.html", vm)
		return
	}

	if _, err := a.Data.GetUserByEmail(email); err == nil {
		if a.AuthLimiter != nil {
			a.AuthLimiter.recordSignupFailure(r)
		}
		http.Error(w, "An account with that email already exists.", http.StatusConflict)
		return
	} else if !errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	if _, err := a.Data.GetUserByName(name); err == nil {
		if a.AuthLimiter != nil {
			a.AuthLimiter.recordSignupFailure(r)
		}
		http.Error(w, "That username is already taken.", http.StatusConflict)
		return
	} else if !errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	hash, err := hashPassword(pass)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	newUser := models.User{
		Name:         name,
		Email:        email,
		PasswordHash: hash,
	}
	id, err := a.Data.CreateUser(newUser)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	token, err := a.Data.IssueAuthToken(int(id), models.PurposeVerify, verifyTokenTTL)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	link := a.BaseURL + "/verify-email?token=" + url.QueryEscape(token)
	if err := mail.SendVerification(a.Mailer, email, name, link); err != nil {
		http.Error(w, "Could not send verification email. Please try again later.", http.StatusInternalServerError)
		return
	}

	a.notifyNewAccount(int(id), name, email)

	a.authPage(w, r, "signup_sent.html", AuthViewModel{Email: email})
}

func (a *App) notifyNewAccount(userID int, username, email string) {
	if a.NewAccountNotificationEmail == "" {
		return
	}
	if err := mail.SendNewAccountNotification(a.Mailer, a.NewAccountNotificationEmail, username, email, userID); err != nil {
		log.Printf("new account notification to %s: %v", a.NewAccountNotificationEmail, err)
	}
}

func (a *App) verifyEmailHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	token := strings.TrimSpace(r.URL.Query().Get("token"))
	if token == "" {
		a.authPage(w, r, "verify_email_result.html", AuthViewModel{Error: "invalid"})
		return
	}

	userID, err := a.Data.ConsumeAuthToken(token, models.PurposeVerify)
	if err != nil {
		status := "invalid"
		switch {
		case errors.Is(err, models.ErrTokenExpired):
			status = "expired"
		case errors.Is(err, models.ErrTokenUsed):
			status = "used"
		}
		a.authPage(w, r, "verify_email_result.html", AuthViewModel{Error: status})
		return
	}

	if err := a.Data.MarkEmailVerified(userID); err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/login?verified=1", http.StatusSeeOther)
}

func (a *App) forgotPasswordHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		a.authPage(w, r, "forgot_password.html", AuthViewModel{})
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if a.AuthLimiter != nil && a.AuthLimiter.blockedForgot(r) {
		http.Error(w, "Too many requests. Try again in a few minutes.", http.StatusTooManyRequests)
		return
	}

	email := normalizeEmail(r.FormValue("email"))
	if !validEmail(email) {
		if a.AuthLimiter != nil {
			a.AuthLimiter.recordForgotFailure(r)
		}
		a.authPage(w, r, "forgot_password.html", AuthViewModel{Email: email, Error: "Please enter a valid email address."})
		return
	}

	a.sendPasswordResetIfEligible(email)

	a.authPage(w, r, "forgot_password_sent.html", AuthViewModel{})
}

func (a *App) sendPasswordResetIfEligible(email string) {
	if a.AuthLimiter != nil && a.AuthLimiter.forgotEmailLimited(email) {
		return
	}

	u, err := a.Data.GetUserByEmail(email)
	if err != nil {
		return
	}
	if !u.IsEmailVerified() {
		return
	}

	token, err := a.Data.IssueAuthToken(u.ID, models.PurposeReset, resetTokenTTL)
	if err != nil {
		return
	}

	link := a.BaseURL + "/reset-password?token=" + url.QueryEscape(token)
	if err := mail.SendReset(a.Mailer, email, u.Name, link); err != nil {
		return
	}

	if a.AuthLimiter != nil {
		a.AuthLimiter.markForgotEmail(email)
	}
}

func (a *App) resetPasswordHandler(w http.ResponseWriter, r *http.Request) {
	token := strings.TrimSpace(r.URL.Query().Get("token"))
	if r.Method == http.MethodGet {
		if token == "" {
			a.authPage(w, r, "verify_email_result.html", AuthViewModel{Error: "invalid"})
			return
		}
		a.authPage(w, r, "reset_password.html", AuthViewModel{Token: token})
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	token = strings.TrimSpace(r.FormValue("token"))
	pass := r.FormValue("password")
	confirm := r.FormValue("password_confirm")

	if token == "" {
		a.authPage(w, r, "verify_email_result.html", AuthViewModel{Error: "invalid"})
		return
	}
	if !validPassword(pass) {
		a.authPage(w, r, "reset_password.html", AuthViewModel{Token: token, Error: "Password must be at least 12 characters."})
		return
	}
	if pass != confirm {
		a.authPage(w, r, "reset_password.html", AuthViewModel{Token: token, Error: "Passwords do not match."})
		return
	}

	userID, err := a.Data.ConsumeAuthToken(token, models.PurposeReset)
	if err != nil {
		status := "invalid"
		switch {
		case errors.Is(err, models.ErrTokenExpired):
			status = "expired"
		case errors.Is(err, models.ErrTokenUsed):
			status = "used"
		}
		a.authPage(w, r, "verify_email_result.html", AuthViewModel{Error: status})
		return
	}

	hash, err := hashPassword(pass)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
	if err := a.Data.UpdatePassword(userID, hash); err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/login?reset=1", http.StatusSeeOther)
}

func (a *App) resendVerificationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if a.AuthLimiter != nil && a.AuthLimiter.blockedResend(r) {
		http.Error(w, "Too many requests. Try again in a few minutes.", http.StatusTooManyRequests)
		return
	}

	email := normalizeEmail(r.FormValue("email"))
	if !validEmail(email) {
		if a.AuthLimiter != nil {
			a.AuthLimiter.recordResendFailure(r)
		}
		a.authPage(w, r, "signup_sent.html", AuthViewModel{Message: "If an unverified account exists for that email, we sent a new link."})
		return
	}

	if a.AuthLimiter != nil && a.AuthLimiter.resendEmailCooldown(email) {
		a.authPage(w, r, "signup_sent.html", AuthViewModel{Email: email, Message: "If an unverified account exists for that email, we sent a new link."})
		return
	}

	u, err := a.Data.GetUserByEmail(email)
	if err == nil && !u.IsEmailVerified() {
		token, err := a.Data.IssueAuthToken(u.ID, models.PurposeVerify, verifyTokenTTL)
		if err == nil {
			link := a.BaseURL + "/verify-email?token=" + url.QueryEscape(token)
			_ = mail.SendVerification(a.Mailer, email, u.Name, link)
			if a.AuthLimiter != nil {
				a.AuthLimiter.markResendEmail(email)
			}
		}
	} else if err != nil && !errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	a.authPage(w, r, "signup_sent.html", AuthViewModel{Email: email, Message: "If an unverified account exists for that email, we sent a new link."})
}
