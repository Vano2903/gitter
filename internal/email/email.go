package email

import (
	"regexp"

	gomail "gopkg.in/mail.v2"
)

func SendEmail(fromEmail, fromPassword, toEmail, subject, body string) error {
	message := gomail.NewMessage()

	message.SetHeader("From", fromEmail)
	message.SetHeader("To", toEmail)
	message.SetHeader("Subject", subject)
	message.SetBody("text/html", body)

	dailer := gomail.NewDialer("smtp.gmail.com", 587, fromEmail, fromPassword)

	return dailer.DialAndSend(message)
}

//function that use regex to validate email
func IsValid(email string) bool {
	re := regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
	return re.MatchString(email)
}
