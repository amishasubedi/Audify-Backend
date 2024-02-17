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

/*
* This method sends veirifcation token to user's email to verify their account
 */
func SendVerificationMail(token string, profile Profile) error {
	d := generateMailDialer()
	m := gomail.NewMessage()

	m.SetHeader("From", "auth@myapp.com")
	m.SetHeader("To", profile.Email)
	m.SetHeader("Subject", "Welcome to MyApp")

	htmlContent := fmt.Sprintf(`<p>Hi %s, welcome to MyApp! Use the given OTP to verify your email: %s</p>`, profile.Name, token)
	m.SetBody("text/html", htmlContent)

	if err := d.DialAndSend(m); err != nil {
		fmt.Printf("Failed to send email: %v\n", err)
		return err
	}

	fmt.Println("Email sent to:", profile.Email)
	return nil
}

type Option struct {
	Email string
	Link  string
}

/*
* This method sends password reset link to user's email
 */
func SendForgetPasswordLink(option Option) {
	d := generateMailDialer()
	m := gomail.NewMessage()

	m.SetHeader("From", "auth@myapp.com")
	m.SetHeader("To", option.Email)
	m.SetHeader("Subject", "Reset Password Link")

	htmlContent := fmt.Sprintf(`<p>We just received a request that you forgot your password, use the link below to create a new password: %s </p>`, option.Link)
	m.SetBody("text/html", htmlContent)

	if err := d.DialAndSend(m); err != nil {
		fmt.Printf("Failed to send email: %v\n", err)
	} else {
		fmt.Println("Email sent to:", option.Email)
	}

}
