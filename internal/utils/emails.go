package utils

import (
	"aicademy-backend/internal/domain/user"
	"bytes"
	"crypto/tls"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/gomail.v2"
)

type EmailData struct {
	URL       string
	FirstName string
	Subject   string
}

func ParseTemplateDir(dir string) (*template.Template, error) {
	var paths []string
	err := filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(strings.ToLower(path), ".html") {
			paths = append(paths, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if len(paths) == 0 {
		return nil, fmt.Errorf("no HTML templates found in %s", dir)
	}
	tmpl, err := template.ParseFiles(paths...)
	if err != nil {
		return nil, fmt.Errorf("failed to parse templates: %v", err)
	}
	return tmpl, nil
}

func SendEmail(user *user.User, data *EmailData, templateName string) error {
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPortStr := os.Getenv("SMTP_PORT")
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASS")
	emailFrom := os.Getenv("EMAIL_FROM")

	if smtpHost == "" || smtpPortStr == "" || smtpUser == "" || smtpPass == "" {
		return fmt.Errorf("missing SMTP configuration")
	}

	if emailFrom == "" {
		emailFrom = smtpUser
	}

	smtpPort, err := strconv.Atoi(smtpPortStr)
	if err != nil {
		return fmt.Errorf("invalid SMTP port: %v", err)
	}

	var body bytes.Buffer
	tmpl, err := ParseTemplateDir("templates")
	if err != nil {
		if os.Getenv("APP_ENV") != "production" {
			log.Printf("Could not parse template: %v", err)
		}
		return err
	}

	if os.Getenv("APP_ENV") != "production" {
		log.Printf("Available templates: %v", tmpl.DefinedTemplates())
		log.Printf("Executing template: %s", templateName)
		log.Printf("Template data: %+v", data)
	}

	err = tmpl.ExecuteTemplate(&body, templateName, data)
	if err != nil {
		if os.Getenv("APP_ENV") != "production" {
			log.Printf("Could not execute template %s: %v", templateName, err)
		}
		return fmt.Errorf("failed to execute template: %v", err)
	}

	htmlContent := body.String()

	if len(htmlContent) == 0 {
		return fmt.Errorf("template generated empty content")
	}

	if os.Getenv("APP_ENV") != "production" {
		if len(htmlContent) > 500 {
			log.Printf("Generated HTML preview: %s...", htmlContent[:500])
		} else {
			log.Printf("Generated HTML: %s", htmlContent)
		}
		log.Printf("Email data: FirstName=%s, Subject=%s, URL=%s", data.FirstName, data.Subject, data.URL)
	}

	m := gomail.NewMessage()
	m.SetHeader("From", emailFrom)
	m.SetHeader("To", user.Email)
	m.SetHeader("Subject", data.Subject)
	m.SetBody("text/html", htmlContent)

	d := gomail.NewDialer(smtpHost, smtpPort, smtpUser, smtpPass)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	if err := d.DialAndSend(m); err != nil {
		log.Printf("Failed to send email to %s: %v", user.Email, err)
		return err
	}

	if os.Getenv("APP_ENV") != "production" {
		log.Printf("Email sent successfully to %s", user.Email)
	} else {
		log.Printf("Password reset email sent: %s", user.Email)
	}

	return nil
}

func SendResetPasswordEmail(user *user.User, resetToken string) error {
	firstName := strings.Split(user.Email, "@")[0]

	if len(firstName) > 0 {
		firstName = strings.ToUpper(string(firstName[0])) + firstName[1:]
	}

	resetURL := fmt.Sprintf("%s/reset-password/%s", os.Getenv("CLIENT_ORIGIN"), resetToken)
	if os.Getenv("CLIENT_ORIGIN") == "" {
		resetURL = fmt.Sprintf("http://localhost:3000/reset-password/%s", resetToken)
	}

	emailData := &EmailData{
		URL:       resetURL,
		FirstName: firstName,
		Subject:   "Reset Your Password - AICademy",
	}

	return SendEmail(user, emailData, "resetPassword")
}
