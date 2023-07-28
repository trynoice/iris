package cmd_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/ashutoshgngwr/iris-cli/internal/cmd"
	"github.com/ashutoshgngwr/iris-cli/internal/testutil"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitCommand(t *testing.T) {
	const cfgFileName = ".iris.yaml"
	expectedFiles := []string{
		"subject.txt",
		"body.txt",
		"body.html",
		"default.csv",
		"recipient.csv",
		cfgFileName,
	}

	t.Run("WithoutPositionalArg", func(t *testing.T) {
		c := cmd.InitCommand(viper.New(), cfgFileName)
		out := &bytes.Buffer{}
		c.SetOut(out)
		c.SetErr(out)
		err := c.Execute()
		assert.Error(t, err)
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

	for file, content := range map[string]string{
		"subject.txt":   "test-subject",
		"body.txt":      "test-text-body",
		"body.html":     "test-html-body",
		"default.csv":   "test-default-csv",
		"recipient.csv": "test-recipient.csv",
		cfgFileName:     "test-config",
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
