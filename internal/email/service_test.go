package email_test

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/trynoice/iris/internal/email"
)

func TestAwsSesService(t *testing.T) {
	t.Run("WithNilMessage", func(t *testing.T) {
		c := &FakeAwsSesClient{RespondWithOutput: &ses.SendEmailOutput{}}
		s := email.NewAwsSesServiceWithClient(c)
		err := s.Send(&email.SendOptions{
			From: "test-from",
			To:   "test-to",
		})
		assert.Error(t, err)
	})

	t.Run("WithUpstreamError", func(t *testing.T) {
		c := &FakeAwsSesClient{RespondWithError: fmt.Errorf("test-error")}
		s := email.NewAwsSesServiceWithClient(c)
		err := s.Send(&email.SendOptions{
			From:    "test-from",
			To:      "test-to",
			Message: &email.Message{},
		})
		assert.Error(t, err)
	})

	t.Run("WithNoError", func(t *testing.T) {
		sendOpts := &email.SendOptions{
			From:    "test-from",
			To:      "test-to",
			ReplyTo: []string{"test-reply-to"},
			Message: &email.Message{
				Subject:  "test-subject",
				TextBody: "test-text-body",
				HtmlBody: "test-html-body",
			},
		}

		c := &FakeAwsSesClient{RespondWithOutput: &ses.SendEmailOutput{}}
		s := email.NewAwsSesServiceWithClient(c)
		err := s.Send(sendOpts)
		assert.NoError(t, err)

		i := c.LastSendEmailInput
		assert.Equal(t, sendOpts.From, *i.Source)
		assert.Equal(t, sendOpts.To, *i.Destination.ToAddresses[0])
		assert.Equal(t, sendOpts.ReplyTo, aws.StringValueSlice(i.ReplyToAddresses))
		assert.Equal(t, sendOpts.Message.Subject, *i.Message.Subject.Data)
		assert.Equal(t, sendOpts.Message.TextBody, *i.Message.Body.Text.Data)
		assert.Equal(t, sendOpts.Message.HtmlBody, *i.Message.Body.Html.Data)
	})
}

type FakeAwsSesClient struct {
	RespondWithOutput  *ses.SendEmailOutput
	RespondWithError   error
	LastSendEmailInput *ses.SendEmailInput
}

// SendEmail API operation for Amazon Simple Email Service.
func (c *FakeAwsSesClient) SendEmail(input *ses.SendEmailInput) (*ses.SendEmailOutput, error) {
	c.LastSendEmailInput = input
	return c.RespondWithOutput, c.RespondWithError
}

func TestPrintService(t *testing.T) {
	sendOpts := &email.SendOptions{
		From: "test-from",
		To:   "test-to",
		Message: &email.Message{
			Subject:  "test-subject",
			TextBody: "test-text-body",
			HtmlBody: "test-html-body",
		},
	}

	b := &bytes.Buffer{}
	s := email.NewPrintService(b)
	err := s.Send(sendOpts)
	assert.NoError(t, err)

	out := b.String()
	assert.Contains(t, out, sendOpts.Message.Subject)
	assert.Contains(t, out, sendOpts.Message.TextBody)
	assert.Contains(t, out, sendOpts.Message.HtmlBody)
}

func TestRateLimitedService(t *testing.T) {
	var s email.Service = &unreliableService{errorsBeforeSucceeding: 0}
	s = email.ApplyOptions(s, email.WithRateLimit(1))
	then := time.Now()
	for i := 0; i < 5; i++ {
		err := s.Send(&email.SendOptions{
			From:    "test-from",
			To:      "test-to",
			Message: &email.Message{},
		})
		require.NoError(t, err)
	}

	diff := time.Since(then)
	assert.GreaterOrEqual(t, diff, 4*time.Second)
}

func TestRetryService(t *testing.T) {
	tt := []struct {
		name       string
		retryCount int
		errorCount int
		wantErr    bool
	}{
		{
			name:       "WithMoreErrorsThanRetries",
			retryCount: 4,
			errorCount: 5,
			wantErr:    true,
		},
		{
			name:       "WithLessErrorsThanRetries",
			retryCount: 4,
			errorCount: 3,
			wantErr:    false,
		},
		{
			name:       "WithEqualErrorsToRetries",
			retryCount: 4,
			errorCount: 4,
			wantErr:    false,
		},
		{
			name:       "WithNoErrors",
			retryCount: 3,
			errorCount: 0,
			wantErr:    false,
		},
	}

	for _, test := range tt {
		t.Run(test.name, func(t *testing.T) {
			var s email.Service = &unreliableService{errorsBeforeSucceeding: test.errorCount}
			s = email.ApplyOptions(s, email.WithRetries(test.retryCount))
			err := s.Send(&email.SendOptions{
				From:    "test-from",
				To:      "test-to",
				Message: &email.Message{},
			})
			if test.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

type unreliableService struct {
	errorsBeforeSucceeding int
}

func (s *unreliableService) Send(opts *email.SendOptions) error {
	s.errorsBeforeSucceeding--
	if s.errorsBeforeSucceeding > -1 {
		return fmt.Errorf("test-error")
	}
	return nil
}
