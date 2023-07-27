package email

import (
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/mitchellh/go-wordwrap"
	"gopkg.in/yaml.v3"
)

type Service interface {
	Send(from string, to string, m *Message) error
}

func NewAwsSesService() (Service, error) {
	s, err := session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to load aws config and credentials: %w", err)
	}

	return &awsSesService{ses: ses.New(s)}, nil
}

type awsSesService struct {
	ses *ses.SES
}

func (s *awsSesService) Send(from string, to string, m *Message) error {
	if _, err := s.ses.SendEmail(&ses.SendEmailInput{
		Source: aws.String(from),
		Destination: &ses.Destination{
			CcAddresses: []*string{},
			ToAddresses: []*string{
				aws.String(to),
			},
		},
		Message: &ses.Message{
			Subject: &ses.Content{
				Charset: aws.String("utf-8"),
				Data:    aws.String(m.Subject),
			},
			Body: &ses.Body{
				Text: &ses.Content{
					Charset: aws.String("utf-8"),
					Data:    aws.String(m.TextBody),
				},
				Html: &ses.Content{
					Charset: aws.String("utf-8"),
					Data:    aws.String(m.HtmlBody),
				},
			},
		},
	}); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

func NewPrintService(w io.Writer) Service {
	encoder := yaml.NewEncoder(w)
	encoder.SetIndent(4)
	return &printService{encoder: encoder}
}

type printService struct {
	encoder *yaml.Encoder
}

func (b *printService) Send(from string, to string, m *Message) error {
	if err := b.encoder.Encode(map[string]string{
		"to":       to,
		"subject":  wordwrap.WrapString(m.Subject, 80),
		"textBody": wordwrap.WrapString(m.TextBody, 80),
		"htmlBody": wordwrap.WrapString(m.HtmlBody, 80),
	}); err != nil {
		return fmt.Errorf("failed to write email: %w", err)
	}

	return nil
}
