package email

import (
	"fmt"
	"net/smtp"
	"strings"
)

type SMTPSender struct {
	host     string
	port     string
	username string
	password string
	from     string
	baseURL  string
}

func NewSMTPSender(host, port, username, password, from, baseURL string) *SMTPSender {
	return &SMTPSender{
		host:     host,
		port:     port,
		username: username,
		password: password,
		from:     from,
		baseURL:  baseURL,
	}
}

// SendVerificationEmail отправляет письмо подтверждения согласно UC-1.1.1
func (s *SMTPSender) SendVerificationEmail(to, token string) error {
	verifyURL := fmt.Sprintf("%s/verify?token=%s", s.baseURL, token)

	subject := "Подтверждение аккаунта - Hubigr"
	body := fmt.Sprintf(`
Добро пожаловать в Hubigr!

Для завершения регистрации подтвердите ваш email, перейдя по ссылке:
%s

Ссылка действительна в течение 1 часа.

Если вы не регистрировались на Hubigr, проигнорируйте это письмо.

--
Команда Hubigr
`, verifyURL)

	return s.sendEmail(to, subject, body)
}

// SendPasswordResetEmail отправляет письмо сброса пароля согласно UC-1.1.3
func (s *SMTPSender) SendPasswordResetEmail(to, token string) error {
	resetURL := fmt.Sprintf("%s/reset-password?token=%s", s.baseURL, token)

	subject := "Сброс пароля - Hubigr"
	body := fmt.Sprintf(`
Запрос на сброс пароля

Для создания нового пароля перейдите по ссылке:
%s

Ссылка действительна в течение 1 часа.

Если вы не запрашивали сброс пароля, проигнорируйте это письмо.

--
Команда Hubigr
`, resetURL)

	return s.sendEmail(to, subject, body)
}

func (s *SMTPSender) sendEmail(to, subject, body string) error {
	// Формирование сообщения
	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s",
		s.from, to, subject, body)

	// Аутентификация
	auth := smtp.PlainAuth("", s.username, s.password, s.host)

	// Отправка
	addr := s.host + ":" + s.port
	return smtp.SendMail(addr, auth, s.from, []string{to}, []byte(msg))
}

// MockSender для тестирования
type MockSender struct {
	SentEmails []SentEmail
}

type SentEmail struct {
	To    string
	Token string
}

func NewMockSender() *MockSender {
	return &MockSender{SentEmails: make([]SentEmail, 0)}
}

func (m *MockSender) SendVerificationEmail(to, token string) error {
	m.SentEmails = append(m.SentEmails, SentEmail{To: to, Token: token})
	fmt.Printf("MOCK EMAIL: Verification sent to %s with token %s\n", maskEmailForMock(to), maskToken(token))
	return nil
}

func (m *MockSender) SendPasswordResetEmail(to, token string) error {
	m.SentEmails = append(m.SentEmails, SentEmail{To: to, Token: token})
	fmt.Printf("MOCK EMAIL: Password reset sent to %s with token %s\n", maskEmailForMock(to), maskToken(token))
	return nil
}

// maskEmailForMock маскирует email для mock sender
func maskEmailForMock(email string) string {
	if len(email) < 3 {
		return "***"
	}
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return "***@***"
	}
	local := parts[0]
	domain := parts[1]
	
	if len(local) <= 2 {
		return "***@" + domain
	}
	return local[:2] + "***@" + domain
}

// maskToken маскирует токен для безопасности
func maskToken(token string) string {
	if len(token) < 8 {
		return "***"
	}
	return token[:4] + "***" + token[len(token)-4:]
}

// EmailSender интерфейс для отправки email
type EmailSender interface {
	SendVerificationEmail(to, token string) error
	SendPasswordResetEmail(to, token string) error
}
