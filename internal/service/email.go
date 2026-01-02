package service

type EmailService interface {
	SendResetLink(email, token string) error
}

type consoleEmailService struct{}

func (s *consoleEmailService) SendResetLink(email, token string) error {
	println("Token: " + token)

	return nil
}

func NewEmailService() EmailService {
	return &consoleEmailService{}
}
