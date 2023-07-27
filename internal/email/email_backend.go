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

type EmailBackend interface {
	Send(e *Email) error
}

type Email struct {
	Sender    string
	Recipient string
	Subject   string
	TextBody  string
	HtmlBody  string
}

func NewAwsSesEmailBackend() (EmailBackend, error) {
	s, err := session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to load aws config and credentials: %w", err)
	}

	return &awsSesEmailBackend{
		ses: ses.New(s),
	}, nil
}

type awsSesEmailBackend struct {
	ses *ses.SES
}

func (b *awsSesEmailBackend) Send(e *Email) error {
	if _, err := b.ses.SendEmail(&ses.SendEmailInput{
		Source: aws.String(e.Sender),
		Destination: &ses.Destination{
			CcAddresses: []*string{},
			ToAddresses: []*string{
				aws.String(e.Recipient),
			},
		},
		Message: &ses.Message{
			Subject: &ses.Content{
				Charset: aws.String("utf-8"),
				Data:    aws.String(e.Subject),
			},
			Body: &ses.Body{
				Text: &ses.Content{
					Charset: aws.String("utf-8"),
					Data:    aws.String(e.TextBody),
				},
				Html: &ses.Content{
					Charset: aws.String("utf-8"),
					Data:    aws.String(e.HtmlBody),
				},
			},
		},
	}); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

func NewDryRunEmailBackend(w io.Writer) EmailBackend {
	encoder := yaml.NewEncoder(w)
	encoder.SetIndent(4)
	return &dryRunEmailBackend{encoder: encoder}
}

type dryRunEmailBackend struct {
	encoder *yaml.Encoder
}

func (b *dryRunEmailBackend) Send(e *Email) error {
	if err := b.encoder.Encode(&Email{
		Sender:    e.Sender,
		Recipient: e.Recipient,
		Subject:   wordwrap.WrapString(e.Subject, 80),
		TextBody:  wordwrap.WrapString(e.TextBody, 80),
		HtmlBody:  wordwrap.WrapString(e.HtmlBody, 80),
	}); err != nil {
		return fmt.Errorf("failed to write email: %w", err)
	}

	return nil
}
