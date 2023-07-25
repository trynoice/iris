package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	Backend BackendConfig
	Email   EmailConfig
}

type BackendConfig struct {
	Provider  BackendProvider
	RateLimit uint
}

type BackendProvider string

const (
	AwsSes BackendProvider = "aws-ses"
)

type EmailConfig struct {
	Sender                   string
	SubjectFile              string
	TextBodyFile             string
	HtmlBodyFile             string
	DefaultDataCsvFile       string
	RecipientDataCsvFile     string
	RecipientEmailColumnName string
}

// Read attempts to read the config file in the current working directory. It
// falls back to sensible defaults if the entire config file or some config
// options are not provided.
func Read(v Viper) (*Config, error) {
	setDefaultConfig(v)
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return cfg, nil
}

// WriteDefault attempts to write a config file with default options in the
// current working directory. It doesn't overwrite an existing config file.
func WriteDefault(v Viper) error {
	setDefaultConfig(v)
	if err := v.SafeWriteConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileAlreadyExistsError); !ok {
			return fmt.Errorf("failed to write config: %w", err)
		}
	}

	return nil
}

func setDefaultConfig(v Viper) {
	v.SetDefault("backend", &BackendConfig{
		Provider:  AwsSes,
		RateLimit: 14,
	})

	v.SetDefault("email", &EmailConfig{
		Sender:                   "Iris CLI <iris@example.test>",
		SubjectFile:              "subject.txt",
		TextBodyFile:             "body.txt",
		HtmlBodyFile:             "body.html",
		DefaultDataCsvFile:       "default.csv",
		RecipientDataCsvFile:     "recipient.csv",
		RecipientEmailColumnName: "Recipient",
	})
}
