package testutil

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// CreateFile creates or updates the given `file` in the given `dir` with the
// given `content` and registers a cleanup function on the provided `t` to
// remove the file once the test completes.
func CreateFile(t *testing.T, dir string, file string, content string) {
	p := filepath.Join(dir, file)
	err := os.WriteFile(p, []byte(content), os.ModePerm)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, os.Remove(p))
	})
}

// Chdir changes the current working directory to the named directory and
// registers a cleanup function on the provided `t` to restore working directory
// to its initial value once the test completes.
func Chdir(t *testing.T, dir string) {
	original, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() {
		require.NoError(t, os.Chdir(original))
	})
}
