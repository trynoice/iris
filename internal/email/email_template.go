package email

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"sync"
	"text/template"

	"github.com/ashutoshgngwr/iris-cli/internal/config"
)

type Email struct {
	Sender    string
	Recipient string
	Subject   string
	TextBody  string
	HtmlBody  string
}

func NewEmailTemplate(cfg *config.EmailConfig) (*EmailTemplate, error) {
	subjectTemplate, err := template.ParseFiles(cfg.SubjectFile)
	if err != nil {
		return nil, fmt.Errorf("failed to parse subject template: %w", err)
	}

	textBodyTemplate, err := template.ParseFiles(cfg.TextBodyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to parse text body template: %w", err)
	}

	htmlBodyTemplate, err := template.ParseFiles(cfg.HtmlBodyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to parse html body template: %w", err)
	}

	defaultDataFile, err := os.Open(cfg.DefaultDataCsvFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open default csv: %w", err)
	}

	defer defaultDataFile.Close()
	defaultDataRows, err := csv.NewReader(defaultDataFile).ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read data from default csv: %w", err)
	}

	if len(defaultDataRows) != 2 {
		return nil, fmt.Errorf("default data csv must have exactly two rows")
	}

	defaultData := buildMap(defaultDataRows[0], defaultDataRows[1])
	recipientDataFile, err := os.Open(cfg.RecipientDataCsvFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open recipient csv: %w", err)
	}

	recipientDataReader := csv.NewReader(recipientDataFile)
	recipientDataReader.ReuseRecord = true
	recipientDataRecord, err := recipientDataReader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read headers from recipient csv: %w", err)
	}

	recipientDataHeaders := make([]string, len(recipientDataRecord))
	copy(recipientDataHeaders, recipientDataRecord)
	return &EmailTemplate{
		mutex:                sync.Mutex{},
		sender:               cfg.Sender,
		subjectTemplate:      subjectTemplate,
		textBodyTemplate:     textBodyTemplate,
		htmlBodyTemplate:     htmlBodyTemplate,
		defaultData:          defaultData,
		recipientDataReader:  recipientDataReader,
		recipientDataHeaders: recipientDataHeaders,
		recipientColumnName:  cfg.RecipientEmailColumnName,
		renderBuffer:         &bytes.Buffer{},
	}, nil
}

type EmailTemplate struct {
	mutex                sync.Mutex
	sender               string
	subjectTemplate      *template.Template
	textBodyTemplate     *template.Template
	htmlBodyTemplate     *template.Template
	defaultData          map[string]string
	recipientDataReader  *csv.Reader
	recipientDataHeaders []string
	recipientColumnName  string
	renderBuffer         *bytes.Buffer
}

func (t *EmailTemplate) RenderNext() (*Email, error) {
	// needs mutex because csv reader is configured to reuse record slice and a
	// shared buffer is used for rendering templates. although never invoked in
	// parallel, it is still a good practice.
	t.mutex.Lock()
	defer t.mutex.Unlock()

	csvRecord, err := t.recipientDataReader.Read()
	if err == io.EOF {
		return nil, err
	} else if err != nil {
		return nil, fmt.Errorf("failed to read data from recipient csv: %w", err)
	}

	recipientData := buildMap(t.recipientDataHeaders, csvRecord)
	for key, defaultValue := range t.defaultData {
		if _, ok := recipientData[key]; !ok {
			recipientData[key] = defaultValue
		}
	}

	t.renderBuffer.Reset()
	if err := t.subjectTemplate.Execute(t.renderBuffer, recipientData); err != nil {
		return nil, fmt.Errorf("failed to render subject template: %w", err)
	}

	subject := t.renderBuffer.String()
	t.renderBuffer.Reset()
	if err := t.textBodyTemplate.Execute(t.renderBuffer, recipientData); err != nil {
		return nil, fmt.Errorf("failed to render text body template: %w", err)
	}

	textBody := t.renderBuffer.String()
	t.renderBuffer.Reset()
	if err := t.htmlBodyTemplate.Execute(t.renderBuffer, recipientData); err != nil {
		return nil, fmt.Errorf("failed to render html body template: %w", err)
	}

	htmlBody := t.renderBuffer.String()
	return &Email{
		Sender:    t.sender,
		Recipient: recipientData[t.recipientColumnName],
		Subject:   subject,
		TextBody:  textBody,
		HtmlBody:  htmlBody,
	}, nil
}

func buildMap(keys []string, values []string) map[string]string {
	data := map[string]string{}
	for i, key := range keys {
		data[key] = values[i]
	}
	return data
}
