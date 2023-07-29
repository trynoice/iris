package cmd_test

import (
	"bytes"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/trynoice/iris/internal/cmd"
	"github.com/trynoice/iris/internal/testutil"
)

func TestSendCommand(t *testing.T) {
	const subject = "test-subject-{{ .name }}"
	const textBody = "test-text-body-{{ .name }}"
	const htmlBody = "test-html-body-{{ .name }}"
	const dataCsv = "data,name\nabc,abc@iris.test\ndef,def@iris.test"
	const cfgFile = ".iris.yaml"
	const cfgFileContent = `
service:
    rateLimit: 1
    retries: 1
message:
    sender: cli@iris.test
    recipientDataCsvFile: data.csv
    recipientEmailColumnName: email`

	newViper := func() *viper.Viper {
		v := viper.New()
		v.SetConfigName(".iris")
		v.SetConfigType("yaml")
		return v
	}

	t.Run("WithoutEmailTemplate", func(t *testing.T) {
		tmpDir := t.TempDir()
		testutil.CreateFile(t, tmpDir, cfgFile, cfgFileContent)
		testutil.CreateFile(t, tmpDir, "data.csv", dataCsv)
		c := cmd.SendCommand(newViper())
		c.SetOut(&bytes.Buffer{})
		c.SetErr(&bytes.Buffer{})
		c.SetArgs([]string{tmpDir})
		require.NoError(t, c.Flags().Set("dry-run", "true"))
		err := c.Execute()
		assert.Error(t, err)
	})

	t.Run("WithoutDataCsv", func(t *testing.T) {
		tmpDir := t.TempDir()
		testutil.CreateFile(t, tmpDir, cfgFile, cfgFileContent)
		testutil.CreateFile(t, tmpDir, "subject.txt", subject)
		testutil.CreateFile(t, tmpDir, "body.txt", textBody)
		testutil.CreateFile(t, tmpDir, "body.html", htmlBody)

		c := cmd.SendCommand(newViper())
		c.SetOut(&bytes.Buffer{})
		c.SetErr(&bytes.Buffer{})
		c.SetArgs([]string{tmpDir})
		require.NoError(t, c.Flags().Set("dry-run", "true"))
		err := c.Execute()
		assert.Error(t, err)
	})

	t.Run("WithEmailTemplateAndDataCsv", func(t *testing.T) {
		tmpDir := t.TempDir()
		testutil.CreateFile(t, tmpDir, cfgFile, cfgFileContent)
		testutil.CreateFile(t, tmpDir, "subject.txt", subject)
		testutil.CreateFile(t, tmpDir, "body.txt", textBody)
		testutil.CreateFile(t, tmpDir, "body.html", htmlBody)
		testutil.CreateFile(t, tmpDir, "data.csv", dataCsv)

		t.Run("WithPositionalArg", func(t *testing.T) {
			testutil.Chdir(t, tmpDir)
			c := cmd.SendCommand(newViper())
			c.SetOut(&bytes.Buffer{})
			c.SetErr(&bytes.Buffer{})
			require.NoError(t, c.Flags().Set("dry-run", "true"))
			err := c.Execute()
			assert.NoError(t, err)
		})

		t.Run("WithoutPositionalArg", func(t *testing.T) {
			c := cmd.SendCommand(newViper())
			c.SetOut(&bytes.Buffer{})
			c.SetErr(&bytes.Buffer{})
			c.SetArgs([]string{tmpDir})
			require.NoError(t, c.Flags().Set("dry-run", "true"))
			err := c.Execute()
			assert.NoError(t, err)
		})
	})
}
