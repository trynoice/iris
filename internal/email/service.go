package email

import (
	"fmt"
	"io"

	"github.com/ashutoshgngwr/iris-cli/internal/config"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"go.uber.org/ratelimit"
	"golang.org/x/term"
)

type Service interface {
	Send(from string, to string, m *Message) error
}

type ServiceOption func(upstream Service) Service

func NewAwsSesService(cfg *config.AwsSesServiceConfig, opts ...ServiceOption) (Service, error) {
	sessionOpts := session.Options{}
	if cfg.UseSharedConfig {
		sessionOpts.SharedConfigState = session.SharedConfigEnable
	} else {
		sessionOpts.SharedConfigState = session.SharedConfigDisable
	}

	if cfg.Region != "" {
		sessionOpts.Config.Region = aws.String(cfg.Region)
	}

	if cfg.Profile != "" {
		sessionOpts.Config.Credentials = credentials.NewSharedCredentials("", cfg.Profile)
	}

	s, err := session.NewSessionWithOptions(sessionOpts)
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
	return applyOptions(&printService{w: w}, opts...)
}

type printService struct {
	w io.Writer
}

func (s *printService) Send(from string, to string, m *Message) error {
	w, _, err := term.GetSize(0)
	if err != nil {
		w = 60
	} else {
		w -= 20 // `| 0 | HTML Body |  |` = 20 chars
	}

	tw := table.NewWriter()
	for _, row := range []table.Row{
		{"Subject", text.WrapSoft(m.Subject, w)},
		{"Text Body", text.WrapSoft(m.TextBody, w)},
		{"HTML Body", text.WrapSoft(m.HtmlBody, w)},
	} {
		tw.AppendRow(row)
		tw.AppendSeparator()
	}

	_, err = fmt.Fprintln(s.w, tw.Render())
	return err
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
	for i := 0; i <= s.retryCount; i++ {
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
