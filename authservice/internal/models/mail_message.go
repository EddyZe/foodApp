package models

type MailMessage struct {
	from    string
	to      []string
	body    string
	subject string
}

func NewMailMessage(from, subject, body string, to ...string) *MailMessage {
	return &MailMessage{
		from:    from,
		to:      to,
		body:    body,
		subject: subject,
	}
}

func (m *MailMessage) From() string {
	return m.from
}

func (m *MailMessage) To() []string {
	return m.to
}

func (m *MailMessage) Body() string {
	return m.body
}

func (m *MailMessage) Subject() string {
	return m.subject
}
