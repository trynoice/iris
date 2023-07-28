package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ashutoshgngwr/iris-cli/internal/config"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestRead(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "iris-config-test-*")
	require.Nil(t, err)
	defer os.RemoveAll(tmpDir)

	t.Run("WithoutConfigFile", func(t *testing.T) {
		v := viper.New()
		v.AddConfigPath(tmpDir)
		v.SetConfigName(".iris")
		v.SetConfigType("yaml")

		// must return default values without an error
		got, err := config.Read(v)
		assert.Nil(t, err)
		assert.NotEmpty(t, got.Service.RateLimit)
		assert.NotEmpty(t, got.Message.Sender)
	})

	t.Run("WithConfigFile", func(t *testing.T) {
		want := &config.Config{
			Service: config.ServiceConfig{
				RateLimit: 1,
			},
			Message: config.MessageConfig{
				Sender: "test-sender",
			},
		}

		cfgData, err := yaml.Marshal(want)
		require.Nil(t, err)

		err = os.WriteFile(filepath.Join(tmpDir, ".iris.yaml"), cfgData, os.ModePerm)
		require.Nil(t, err)

		v := viper.New()
		v.AddConfigPath(tmpDir)
		v.SetConfigName(".iris")
		v.SetConfigType("yaml")

		got, err := config.Read(v)
		assert.Nil(t, err)

		// must return overridden values
		assert.Equal(t, want.Service.RateLimit, got.Service.RateLimit)
		assert.Equal(t, want.Message.Sender, got.Message.Sender)

		// must return default values
		assert.NotEmpty(t, got.Service.Retries)
		assert.NotEmpty(t, got.Message.RecipientEmailColumnName)
	})
}

func TestWriteDefault(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "iris-config-test-*")
	require.Nil(t, err)
	defer os.RemoveAll(tmpDir)

	v := viper.New()
	v.AddConfigPath(tmpDir)
	v.SetConfigName(".iris")
	v.SetConfigType("yaml")

	cfgFile := filepath.Join(tmpDir, ".iris.yaml")
	err = config.WriteDefault(v, cfgFile)
	assert.Nil(t, err)

	gotData, err := os.ReadFile(cfgFile)
	assert.Nil(t, err)

	got := &config.Config{}
	err = yaml.Unmarshal(gotData, got)
	require.Nil(t, err)
	assert.NotEmpty(t, got.Service.Retries)
	assert.NotEmpty(t, got.Message.RecipientEmailColumnName)
}
