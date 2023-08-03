package email

import (
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/mitchellh/go-wordwrap"
	"github.com/olekukonko/tablewriter"
	"github.com/trynoice/iris/internal/config"
	"go.uber.org/ratelimit"
	"golang.org/x/term"
)

type Service interface {
	Send(opts *SendOptions) error
}

type SendOptions struct {
	From    string
	To      string
	ReplyTo []string
	Message *Message
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

	return NewAwsSesServiceWithClient(ses.New(s), opts...), nil
}

func NewAwsSesServiceWithClient(client AwsSesClient, opts ...ServiceOption) Service {
	return ApplyOptions(&awsSesService{client: client}, opts...)
}

type AwsSesClient interface {
	// SendEmail API operation for Amazon Simple Email Service.
	SendEmail(input *ses.SendEmailInput) (*ses.SendEmailOutput, error)
}

type awsSesService struct {
	client AwsSesClient
}

func (s *awsSesService) Send(opts *SendOptions) error {
	if opts == nil {
		return fmt.Errorf("send options must not be nil")
	}

	if opts.Message == nil {
		return fmt.Errorf("message must not be nil")
	}

	if _, err := s.client.SendEmail(&ses.SendEmailInput{
		Source: aws.String(opts.From),
		Destination: &ses.Destination{
			ToAddresses: []*string{
				aws.String(opts.To),
			},
		},
		ReplyToAddresses: aws.StringSlice(opts.ReplyTo),
		Message: &ses.Message{
			Subject: &ses.Content{
				Charset: aws.String("utf-8"),
				Data:    aws.String(opts.Message.Subject),
			},
			Body: &ses.Body{
				Text: &ses.Content{
					Charset: aws.String("utf-8"),
					Data:    aws.String(opts.Message.TextBody),
				},
				Html: &ses.Content{
					Charset: aws.String("utf-8"),
					Data:    aws.String(opts.Message.HtmlBody),
				},
			},
		},
	}); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

func NewPrintService(w io.Writer, opts ...ServiceOption) Service {
	return ApplyOptions(&printService{w: w}, opts...)
}

type printService struct {
	w io.Writer
}

func (s *printService) Send(opts *SendOptions) error {
	if opts == nil {
		return fmt.Errorf("send options must not be nil")
	}

	if opts.Message == nil {
		return fmt.Errorf("message must not be nil")
	}

	pw := getTerminalWidth(100)
	// `| HTML Body |  |` = 16 chars is longest for static data in a row
	if pw > 16 {
		pw -= 16
	}
	if pw > 100 {
		pw = 100
	}

	tw := tablewriter.NewWriter(s.w)
	tw.SetColWidth(pw)
	tw.SetAutoWrapText(false)
	tw.SetRowLine(true)
	tw.AppendBulk([][]string{
		{"Subject", wordwrap.WrapString(opts.Message.Subject, uint(pw))},
		{"Text Body", wordwrap.WrapString(opts.Message.TextBody, uint(pw))},
		{"HTML Body", wordwrap.WrapString(opts.Message.HtmlBody, uint(pw))},
	})
	tw.Render()
	return nil
}

func getTerminalWidth(defaultW int) int {
	w, _, err := term.GetSize(0)
	if err != nil {
		return defaultW
	}
	return w
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

func (s *rateLimitedService) Send(opts *SendOptions) error {
	s.limiter.Take()
	return s.upstream.Send(opts)
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

func (s *retryService) Send(opts *SendOptions) error {
	var err error
	for i := 0; i <= s.retryCount; i++ {
		err = s.upstream.Send(opts)
		if err == nil {
			break
		}
	}

	return err
}

// ApplyOptions wraps the given `upstream` service in the given service options
// (decorators).
func ApplyOptions(upstream Service, opts ...ServiceOption) Service {
	s := upstream
	for _, option := range opts {
		s = option(s)
	}
	return s
}
