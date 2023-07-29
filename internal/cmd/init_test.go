package cmd_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/trynoice/iris/internal/cmd"
	"github.com/trynoice/iris/internal/testutil"
)

func TestInitCommand(t *testing.T) {
	const cfgFileName = ".iris.yaml"
	expectedFiles := []string{
		"subject.txt",
		"body.txt",
		"body.html",
		"default.csv",
		"recipients.csv",
		cfgFileName,
	}

	t.Run("WithoutPositionalArg", func(t *testing.T) {
		tmpDir := t.TempDir()
		testutil.Chdir(t, tmpDir)

		c := cmd.InitCommand(viper.New(), cfgFileName)
		out := &bytes.Buffer{}
		c.SetOut(out)
		c.SetErr(out)
		err := c.Execute()
		assert.NoError(t, err)

		files, err := os.ReadDir(tmpDir)
		require.NoError(t, err)

		for _, file := range files {
			assert.Contains(t, expectedFiles, file.Name())
		}
	})

	t.Run("WithEmptyDirectory", func(t *testing.T) {
		tmpDir := t.TempDir()
		c := cmd.InitCommand(viper.New(), cfgFileName)
		out := &bytes.Buffer{}
		c.SetOut(out)
		c.SetErr(out)
		c.SetArgs([]string{tmpDir})
		err := c.Execute()
		assert.NoError(t, err)

		files, err := os.ReadDir(tmpDir)
		require.NoError(t, err)

		for _, file := range files {
			assert.Contains(t, expectedFiles, file.Name())
		}
	})

	t.Run("WithExistingConfig", func(t *testing.T) {
		cfgData := []byte(`
message:
    recipientDataCsvFile: data.csv
    recipientEmailColumnName: Recipient`)

		tmpDir := t.TempDir()
		testutil.CreateFile(t, tmpDir, cfgFileName, string(cfgData))

		v := viper.New()
		v.SetConfigName(".iris")
		v.SetConfigType("yaml")

		c := cmd.InitCommand(v, cfgFileName)
		out := &bytes.Buffer{}
		c.SetOut(out)
		c.SetErr(out)
		c.SetArgs([]string{tmpDir})
		err := c.Execute()
		assert.NoError(t, err)

		files, err := os.ReadDir(tmpDir)
		require.NoError(t, err)

		for _, f := range files {
			assert.Contains(t, []string{
				"subject.txt",
				"body.txt",
				"body.html",
				"data.csv",
				".iris.yaml",
			}, f.Name())
		}

		// must not change config
		gotCfgData, err := os.ReadFile(filepath.Join(tmpDir, cfgFileName))
		assert.NoError(t, err)
		assert.Equal(t, cfgData, gotCfgData)

		// must generate files according to the given config
		bodyTxtData, err := os.ReadFile(filepath.Join(tmpDir, "body.txt"))
		assert.NoError(t, err)
		assert.Contains(t, string(bodyTxtData), "{{.Recipient}}")

		bodyHtmlData, err := os.ReadFile(filepath.Join(tmpDir, "body.html"))
		assert.NoError(t, err)
		assert.Contains(t, string(bodyHtmlData), "{{.Recipient}}")

		dataCsvData, err := os.ReadFile(filepath.Join(tmpDir, "data.csv"))
		assert.NoError(t, err)
		assert.Contains(t, string(dataCsvData), "Recipient")
	})

	for file, content := range map[string]string{
		"subject.txt":    "test-subject",
		"body.txt":       "test-text-body",
		"body.html":      "test-html-body",
		"default.csv":    "test-default-csv",
		"recipients.csv": "test-recipient.csv",
	} {
		t.Run("WithExisting__"+file, func(t *testing.T) {
			tmpDir := t.TempDir()
			testutil.CreateFile(t, tmpDir, file, content)
			c := cmd.InitCommand(viper.New(), cfgFileName)
			out := &bytes.Buffer{}
			c.SetOut(out)
			c.SetErr(out)
			c.SetArgs([]string{tmpDir})
			err := c.Execute()
			assert.NoError(t, err)

			files, err := os.ReadDir(tmpDir)
			require.NoError(t, err)

			for _, f := range files {
				assert.Contains(t, expectedFiles, f.Name())
			}

			got, err := os.ReadFile(filepath.Join(tmpDir, file))
			require.NoError(t, err)
			assert.Equal(t, content, string(got))
		})
	}
}
