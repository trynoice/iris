package email

import (
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/mitchellh/go-wordwrap"
	"go.uber.org/ratelimit"
	"gopkg.in/yaml.v3"
)

type Service interface {
	Send(from string, to string, m *Message) error
}

type ServiceOption func(upstream Service) Service

func NewAwsSesService(opts ...ServiceOption) (Service, error) {
	s, err := session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to load aws config and credentials: %w", err)
	}

	return applyOptions(&awsSesService{ses: ses.New(s)}, opts...), nil
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

func NewPrintService(w io.Writer, opts ...ServiceOption) Service {
	encoder := yaml.NewEncoder(w)
	encoder.SetIndent(4)
	return applyOptions(&printService{encoder: encoder}, opts...)
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

func WithRateLimit(frequency int) ServiceOption {
	return func(upstream Service) Service {
		return &rateLimitedService{
			upstream: upstream,
			limiter:  ratelimit.New(frequency),
		}
	}
}

type rateLimitedService struct {
	upstream Service
	limiter  ratelimit.Limiter
}

func (s *rateLimitedService) Send(from string, to string, m *Message) error {
	s.limiter.Take()
	return s.upstream.Send(from, to, m)
}

func WithRetries(retryCount int) ServiceOption {
	return func(upstream Service) Service {
		return &retryService{
			upstream:   upstream,
			retryCount: retryCount,
		}
	}
}

type retryService struct {
	upstream   Service
	retryCount int
}

func (s *retryService) Send(from string, to string, m *Message) error {
	var err error
	for i := 0; i < s.retryCount; i++ {
		err = s.upstream.Send(from, to, m)
		if err == nil {
			break
		}
	}

	return err
}

func applyOptions(upstream Service, opts ...ServiceOption) Service {
	s := upstream
	for _, option := range opts {
		s = option(s)
	}
	return s
}
