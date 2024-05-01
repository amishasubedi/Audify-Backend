package utils

import (
	"backend/internal/utils/templates"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/gomail.v2"
)

type Profile struct {
	Name   string
	Email  string
	UserID string
}

func generateMailDialer() *gomail.Dialer {
	return gomail.NewDialer(
		"smtp.sendgrid.net",
		587,
		"apikey",
		os.Getenv("SENDGRID_API_KEY"),
	)
}

/*
* This method sends veirifcation token to user's email to verify their account
 */
func SendVerificationMail(token string, profile Profile) error {
	d := generateMailDialer()
	m := gomail.NewMessage()

	m.SetHeader("From", "email@audify.life")
	m.SetHeader("To", profile.Email)
	m.SetHeader("Subject", "Welcome to Audify")

	options := templates.Options{
		Title:     "Welcome to Audify",
		Message:   fmt.Sprintf("Hi %s, Welcome to Audify! Use the given OTP to verify your email.", profile.Name),
		LogoCID:   "logo",
		BannerCID: "welcome",
		Link:      "#",
		BtnTitle:  token,
	}

	htmlContent := templates.GenerateTemplate(options)
	m.SetBody("text/html", htmlContent)

	basePath := "../internal/utils/templates"

	m.Attach(filepath.Join(basePath, "logo.png"), gomail.SetHeader(map[string][]string{"Content-ID": {"<logo>"}}))
	m.Attach(filepath.Join(basePath, "welcome.png"), gomail.SetHeader(map[string][]string{"Content-ID": {"<welcome>"}}))

	if err := d.DialAndSend(m); err != nil {
		fmt.Printf("Failed to send email: %v\n", err)
		return err
	}

	fmt.Println("Email sent to:", profile.Email)
	return nil
}
