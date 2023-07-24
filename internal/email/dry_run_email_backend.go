package email

import (
	"fmt"
	"io"

	"github.com/mitchellh/go-wordwrap"
	"gopkg.in/yaml.v3"
)

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
		Recipient: e.Recipient,
		Subject:   wordwrap.WrapString(e.Subject, 80),
		TextBody:  wordwrap.WrapString(e.TextBody, 80),
		HtmlBody:  wordwrap.WrapString(e.HtmlBody, 80),
	}); err != nil {
		return fmt.Errorf("failed to write email: %w", err)
	}

	return nil
}
