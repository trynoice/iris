package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/trynoice/iris/internal/config"
	"gopkg.in/yaml.v3"
)

func TestRead(t *testing.T) {
	t.Run("WithoutConfigFile", func(t *testing.T) {
		tmpDir := t.TempDir()
		v := viper.New()
		v.AddConfigPath(tmpDir)
		v.SetConfigName(".iris")
		v.SetConfigType("yaml")

		// must return default values without an error
		got, err := config.Read(v)
		assert.NoError(t, err)
		assert.NotEmpty(t, got.Service.RateLimit)
		assert.NotEmpty(t, got.Message.MinifyHtml)
	})

	t.Run("WithConfigFile", func(t *testing.T) {
		tmpDir := t.TempDir()
		want := &config.Config{
			Service: config.ServiceConfig{
				RateLimit: 1,
			},
			Message: config.MessageConfig{
				Sender: "test-sender",
			},
		}

		cfgData, err := yaml.Marshal(want)
		require.NoError(t, err)

		err = os.WriteFile(filepath.Join(tmpDir, ".iris.yaml"), cfgData, os.ModePerm)
		require.NoError(t, err)

		v := viper.New()
		v.AddConfigPath(tmpDir)
		v.SetConfigName(".iris")
		v.SetConfigType("yaml")

		got, err := config.Read(v)
		assert.NoError(t, err)

		// must return overridden values
		assert.Equal(t, want.Service.RateLimit, got.Service.RateLimit)
		assert.Equal(t, want.Message.Sender, got.Message.Sender)

		// must return default values
		assert.NotEmpty(t, got.Service.Retries)
		assert.NotEmpty(t, got.Message.MinifyHtml)
	})
}

func TestWrite(t *testing.T) {
	want := &config.Config{
		Service: config.ServiceConfig{
			AwsSes: &config.AwsSesServiceConfig{
				UseSharedConfig: true,
			},
			RateLimit: 1,
			Retries:   1,
		},
		Message: config.MessageConfig{
			Sender:                   "test-sender",
			DefaultDataCsvFile:       "default.csv",
			RecipientDataCsvFile:     "recipients.csv",
			RecipientEmailColumnName: "Email",
			MinifyHtml:               false,
		},
	}

	tmpDir := t.TempDir()
	cfgFile := filepath.Join(tmpDir, ".iris.yaml")
	err := config.Write(want, cfgFile)
	assert.NoError(t, err)

	gotData, err := os.ReadFile(cfgFile)
	assert.NoError(t, err)

	got := &config.Config{}
	err = yaml.Unmarshal(gotData, got)
	require.Nil(t, err)
	assert.Equal(t, want, got)
}
