package email

import (
	"bytes"
	"fmt"
	"path/filepath"
	"sync"
	"text/template"

	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/html"
)

const (
	subjectFile  = "subject.txt"
	textBodyFile = "body.txt"
	htmlBodyFile = "body.html"
)

func NewTemplate(dir string, minifyHtml bool) (*Template, error) {
	template, err := template.ParseFiles(
		filepath.Join(dir, subjectFile),
		filepath.Join(dir, textBodyFile),
		filepath.Join(dir, htmlBodyFile),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to parse email templates: %w", err)
	}

	var m *minify.M
	if minifyHtml {
		m = minify.New()
		m.Add("text/html", &html.Minifier{
			KeepDocumentTags: true,
			KeepEndTags:      true,
		})
	}

	return &Template{
		mutex:        sync.Mutex{},
		template:     template,
		buffer:       &bytes.Buffer{},
		htmlMinifier: m,
	}, nil
}

type Template struct {
	mutex        sync.Mutex
	template     *template.Template
	buffer       *bytes.Buffer
	htmlMinifier *minify.M
}

func (t *Template) Render(data any) (*Message, error) {
	// needs mutex because a shared buffer is used for rendering templates.
	// although never invoked in parallel, it is still a good practice.
	t.mutex.Lock()
	defer t.mutex.Unlock()

	m := &Message{}
	t.buffer.Reset()
	if err := t.template.ExecuteTemplate(t.buffer, subjectFile, data); err != nil {
		return nil, fmt.Errorf("failed to render subject template: %w", err)
	}

	m.Subject = t.buffer.String()
	t.buffer.Reset()
	if err := t.template.ExecuteTemplate(t.buffer, textBodyFile, data); err != nil {
		return nil, fmt.Errorf("failed to render text body template: %w", err)
	}

	m.TextBody = t.buffer.String()
	t.buffer.Reset()
	if err := t.template.ExecuteTemplate(t.buffer, htmlBodyFile, data); err != nil {
		return nil, fmt.Errorf("failed to render html body template: %w", err)
	}

	m.HtmlBody = t.buffer.String()
	if t.htmlMinifier != nil {
		var err error
		m.HtmlBody, err = t.htmlMinifier.String("text/html", m.HtmlBody)
		if err != nil {
			return nil, fmt.Errorf("failed to minify html body: %w", err)
		}
	}
	return m, nil
}
