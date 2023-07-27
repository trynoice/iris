package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Service ServiceConfig `yaml:"service,omitempty"`
	Message MessageConfig `yaml:"message,omitempty"`
}

type ServiceConfig struct {
	AwsSes    *AwsSesBackendConfig `yaml:"awsSes,omitempty"`
	RateLimit uint                 `yaml:"rateLimit,omitempty"`
}

type AwsSesBackendConfig struct{}

type MessageConfig struct {
	Sender                   string `yaml:"sender,omitempty"`
	DefaultDataCsvFile       string `yaml:"defaultDataCsvFile,omitempty"`
	RecipientDataCsvFile     string `yaml:"recipientDataCsvFile,omitempty"`
	RecipientEmailColumnName string `yaml:"recipientEmailColumnName,omitempty"`
}

var defaultCfg = &Config{
	Service: ServiceConfig{
		AwsSes:    nil,
		RateLimit: 14,
	},
	Message: MessageConfig{
		Sender:                   "Iris CLI <iris@example.test>",
		DefaultDataCsvFile:       "default.csv",
		RecipientDataCsvFile:     "recipient.csv",
		RecipientEmailColumnName: "Recipient",
	},
}

// Read attempts to read the config file in the current working directory. It
// falls back to sensible defaults if the entire config file or some config
// options are not provided.
func Read(v Viper) (*Config, error) {
	v.SetDefault("service", &defaultCfg.Service)
	v.SetDefault("message", &defaultCfg.Message)
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

// WriteDefault attempts to write the default config options to a file at the
// give path. If the file already exists, the function returns without writing
// to it or raising an error.
func WriteDefault(file string) error {
	const errFmt = "failed to write config: %w"
	f, err := os.Create(file)
	if err != nil {
		return fmt.Errorf(errFmt, err)
	}

	defer f.Close()
	e := yaml.NewEncoder(f)
	if err := e.Encode(defaultCfg); err != nil {
		return fmt.Errorf(errFmt, err)
	}

	if err := e.Close(); err != nil {
		return fmt.Errorf(errFmt, err)
	}

	return nil
}
