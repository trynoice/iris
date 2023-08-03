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
	Sender                   string   `yaml:"sender,omitempty"`
	ReplyToAddresses         []string `yaml:"replyToAddresses,omitempty"`
	DefaultDataCsvFile       string   `yaml:"defaultDataCsvFile,omitempty"`
	RecipientDataCsvFile     string   `yaml:"recipientDataCsvFile,omitempty"`
	RecipientEmailColumnName string   `yaml:"recipientEmailColumnName,omitempty"`
	MinifyHtml               bool     `yaml:"minifyHtml,omitempty"`
}

// Read attempts to read the config file in the current working directory. It
// falls back to sensible defaults if the entire config file or some config
// options are not provided.
func Read(v *viper.Viper) (*Config, error) {
	v.SetDefault("service.rateLimit", 10)
	v.SetDefault("service.retries", 3)
	v.SetDefault("message.minifyHtml", true)

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

// Write attempts to write the config options to the file at the give dest.
func Write(cfg *Config, dest string) error {
	if cfg == nil {
		return fmt.Errorf("passed nil config")
	}

	// using yaml encoder to write the configuration file in `camelCase`.
	const errFmt = "failed to write config: %w"
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf(errFmt, err)
	}

	err = os.WriteFile(dest, data, os.ModePerm)
	if err != nil {
		return fmt.Errorf(errFmt, err)
	}

	return nil
}
