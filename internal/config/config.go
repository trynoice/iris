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
	AwsSes    *AwsSesServiceConfig `yaml:"awsSes,omitempty"`
	RateLimit int                  `yaml:"rateLimit,omitempty"`
	Retries   int                  `yaml:"retries,omitempty"`
}

type AwsSesServiceConfig struct {
	UseSharedConfig bool   `yaml:"useSharedConfig,omitempty"`
	Region          string `yaml:"region,omitempty"`
	Profile         string `yaml:"profile,omitempty"`
}

type MessageConfig struct {
	Sender                   string `yaml:"sender,omitempty"`
	DefaultDataCsvFile       string `yaml:"defaultDataCsvFile,omitempty"`
	RecipientDataCsvFile     string `yaml:"recipientDataCsvFile,omitempty"`
	RecipientEmailColumnName string `yaml:"recipientEmailColumnName,omitempty"`
	MinifyHtml               bool   `yaml:"minifyHtml,omitempty"`
}

// Read attempts to read the config file in the current working directory. It
// falls back to sensible defaults if the entire config file or some config
// options are not provided.
func Read(v *viper.Viper) (*Config, error) {
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

// WriteDefault attempts to write the default config options to a file at the
// give path.
func WriteDefault(v *viper.Viper, file string) error {
	setDefaultConfig(v)
	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		return err
	}

	// using yaml encoder to write the configuration file in `camelCase`.
	const errFmt = "failed to write config: %w"
	f, err := os.Create(file)
	if err != nil {
		return fmt.Errorf(errFmt, err)
	}

	defer f.Close()
	e := yaml.NewEncoder(f)
	if err := e.Encode(cfg); err != nil {
		return fmt.Errorf(errFmt, err)
	}

	if err := e.Close(); err != nil {
		return fmt.Errorf(errFmt, err)
	}

	return nil
}

func setDefaultConfig(v *viper.Viper) {
	v.SetDefault("service.awsSes.useSharedConfig", true)
	v.SetDefault("service.rateLimit", 14)
	v.SetDefault("service.retries", 3)
	v.SetDefault("message.sender", "Iris CLI <iris@example.test>")
	v.SetDefault("message.defaultDataCsvFile", "default.csv")
	v.SetDefault("message.recipientDataCsvFile", "recipient.csv")
	v.SetDefault("message.recipientEmailColumnName", "Recipient")
	v.SetDefault("message.minifyHtml", true)
}
