package smtp

import (
	"fmt"
	"net/smtp"
	"os"
)

type SmtpService struct {
	User     string
	Password string
	Host     string
	Port     string
}

func NewSmtpService() *SmtpService {
	return &SmtpService{
		User:     os.Getenv("SMTP_USER"),
		Password: os.Getenv("SMTP_PASSWORD"),
		Host:     "smtp.gmail.com", // Defaulting to gmail as per user example
		Port:     "587",
	}
}

func (s *SmtpService) SendEmail(to string, subject string, body string) error {
	if s.User == "" || s.Password == "" {
		return fmt.Errorf("SMTP credentials not found")
	}

	auth := smtp.PlainAuth("", s.User, s.Password, s.Host)
	addr := s.Host + ":" + s.Port

	msg := []byte(fmt.Sprintf("To: %s\r\n"+
		"Subject: %s\r\n"+
		"MIME-Version: 1.0\r\n"+
		"Content-Type: text/plain; charset=\"utf-8\"\r\n"+
		"\r\n"+
		"%s\r\n", to, subject, body))

	err := smtp.SendMail(addr, auth, s.User, []string{to}, msg)
	if err != nil {
		return err
	}
	return nil
}
