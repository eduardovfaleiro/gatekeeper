package service

import (
	"fmt"
	"net/smtp"
)

type EmailService interface {
	SendResetLink(email, token string) error
}

type stmpEmailService struct {
	host string
	port string
	user string
	pass string
	from string
}

func NewEmailService(host, port, user, pass, from string) EmailService {
	return &stmpEmailService{
		host: host,
		port: port,
		user: user,
		pass: pass,
		from: from,
	}
}

func (s *stmpEmailService) SendResetLink(email, token string) error {
	auth := smtp.PlainAuth("", s.user, s.pass, s.host)

	msg := []byte(fmt.Sprintf(
		"To: %s\r\n"+
			"Subject: Gatekeeper - Recuperação de Senha\r\n"+
			"\r\n"+
			"Link para resetar sua senha: http://localhost:3000/reset?token=%s\r\n",
		email, token))

	addr := fmt.Sprintf("%s:%s", s.host, s.port)
	err := smtp.SendMail(addr, auth, s.from, []string{email}, msg)
	if err != nil {
		return fmt.Errorf("falha ao enviar e-mail via SMTP: %w", err)
	}
	return nil
}
