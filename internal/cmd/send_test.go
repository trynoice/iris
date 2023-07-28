package cmd_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/trynoice/iris/internal/cmd"
	"github.com/trynoice/iris/internal/config"
	"github.com/trynoice/iris/internal/testutil"
)

func TestSendCommand(t *testing.T) {
	const subject = "test-subject-{{ .name }}"
	const textBody = "test-text-body-{{ .name }}"
	const htmlBody = "test-html-body-{{ .name }}"
	const dataCsv = "data,name\nabc,abc@iris.test\ndef,def@iris.test"
	cfg := &config.Config{
		Service: config.ServiceConfig{
			RateLimit: 1,
			Retries:   1,
		},
		Message: config.MessageConfig{
			RecipientDataCsvFile:     "data.csv",
			RecipientEmailColumnName: "email",
			Sender:                   "cli@iris.test",
		},
	}

	t.Run("WithoutEmailTemplate", func(t *testing.T) {
		tmpDir := t.TempDir()
		require.NoError(t, os.Chdir(tmpDir))
		testutil.CreateFile(t, tmpDir, "data.csv", dataCsv)

		c := cmd.SendCommand(cfg)
		c.SetOut(&bytes.Buffer{})
		c.SetErr(&bytes.Buffer{})
		require.NoError(t, c.Flags().Set("dry-run", "true"))
		err := c.Execute()
		assert.Error(t, err)
	})

	t.Run("WithoutDataCsv", func(t *testing.T) {
		tmpDir := t.TempDir()
		require.NoError(t, os.Chdir(tmpDir))
		testutil.CreateFile(t, tmpDir, "subject.txt", subject)
		testutil.CreateFile(t, tmpDir, "body.txt", textBody)
		testutil.CreateFile(t, tmpDir, "body.html", htmlBody)

		c := cmd.SendCommand(cfg)
		c.SetOut(&bytes.Buffer{})
		c.SetErr(&bytes.Buffer{})
		require.NoError(t, c.Flags().Set("dry-run", "true"))
		err := c.Execute()
		assert.Error(t, err)
	})

	t.Run("WithEmailTemplateAndDataCsv", func(t *testing.T) {
		tmpDir := t.TempDir()
		require.NoError(t, os.Chdir(tmpDir))
		testutil.CreateFile(t, tmpDir, "subject.txt", subject)
		testutil.CreateFile(t, tmpDir, "body.txt", textBody)
		testutil.CreateFile(t, tmpDir, "body.html", htmlBody)
		testutil.CreateFile(t, tmpDir, "data.csv", dataCsv)

		c := cmd.SendCommand(cfg)
		c.SetOut(&bytes.Buffer{})
		c.SetErr(&bytes.Buffer{})
		require.NoError(t, c.Flags().Set("dry-run", "true"))
		err := c.Execute()
		assert.NoError(t, err)
	})
}
