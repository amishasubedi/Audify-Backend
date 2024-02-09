package utils

import (
	"fmt"
	"os"

	"gopkg.in/gomail.v2"
)

type Profile struct {
	Name   string
	Email  string
	UserID string
}

func generateMailDialer() *gomail.Dialer {
	return gomail.NewDialer(
		"sandbox.smtp.mailtrap.io",
		2525,
		os.Getenv("MAILTRAP_USER"),
		os.Getenv("MAILTRAP_PASS"),
	)
}

func SendVerificationMail(token string, profile Profile) {
	d := generateMailDialer()
	m := gomail.NewMessage()

	m.SetHeader("From", "auth@myapp.com")
	m.SetHeader("To", profile.Email)
	m.SetHeader("Subject", "Welcome to MyApp")

	htmlContent := fmt.Sprintf(`<p>Hi %s, welcome to MyApp! Use the given OTP to verify your email: %s</p>`, profile.Name, token)
	m.SetBody("text/html", htmlContent)

	if err := d.DialAndSend(m); err != nil {
		fmt.Printf("Failed to send email: %v\n", err)
	} else {
		fmt.Println("Email sent to:", profile.Email)
	}
}
