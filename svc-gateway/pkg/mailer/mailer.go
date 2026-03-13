package mailer

import (
	"crypto/tls"
	"log/slog"

	"gopkg.in/gomail.v2"
)

type MailRequest struct {
	To      []string
	Subject string
	Body    string
}

type Mailer struct {
	dialer *gomail.Dialer
	ch     chan *MailRequest
}

func NewMailer(host string, port int, username, password string) *Mailer {
	d := gomail.NewDialer(host, port, username, password)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	slog.Info("mailer initialized", "host", host, "port", port, "username", username)

	return &Mailer{
		dialer: d,
		ch:     make(chan *MailRequest, 100),
	}
}

func (m *Mailer) Start() {
	slog.Info("mailer worker started")
	for req := range m.ch {
		msg := gomail.NewMessage()
		msg.SetHeader("From", m.dialer.Username)
		msg.SetHeader("To", req.To...)
		msg.SetHeader("Subject", req.Subject)
		msg.SetBody("text/html", req.Body)

		if err := m.dialer.DialAndSend(msg); err != nil {
			// Log the error but continue processing other emails
			slog.Error("failed to send email", "error", err, "to", req.To)
			continue
		}
	}
}

func (m *Mailer) SendMail(req *MailRequest) {
	m.ch <- req
}

func (m *Mailer) Close() {
	close(m.ch)
}
