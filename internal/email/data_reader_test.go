package email_test

import (
	"io"
	"path/filepath"
	"testing"

	"github.com/ashutoshgngwr/iris-cli/internal/email"
	"github.com/ashutoshgngwr/iris-cli/internal/testutil"
	"github.com/stretchr/testify/assert"
)

func TestDataReader(t *testing.T) {
	const defaultCsv = "col1,col2\nabc,def"
	const dataCsv = "col2,col3\nghi,jkl\nmno,pqr"

	t.Run("WithoutDataFile", func(t *testing.T) {
		tmpDir := t.TempDir()
		testutil.CreateFile(t, tmpDir, "default.csv", defaultCsv)

		defaultFile := filepath.Join(tmpDir, "default.csv")
		dataFile := filepath.Join(tmpDir, "data.csv")
		r, err := email.NewDataReader(defaultFile, dataFile)
		assert.Error(t, err)
		assert.Nil(t, r)
	})

	t.Run("WithoutDefaultFile", func(t *testing.T) {
		tmpDir := t.TempDir()
		testutil.CreateFile(t, tmpDir, "data.csv", dataCsv)

		defaultFile := filepath.Join(tmpDir, "default.csv")
		dataFile := filepath.Join(tmpDir, "data.csv")
		r, err := email.NewDataReader(defaultFile, dataFile)
		assert.NoError(t, err)
		assert.NotNil(t, r)

		record, err := r.Read()
		assert.NoError(t, err)
		assert.Empty(t, record["col1"])
		assert.Equal(t, "ghi", record["col2"])
		assert.Equal(t, "jkl", record["col3"])

		record, err = r.Read()
		assert.NoError(t, err)
		assert.Empty(t, record["col1"])
		assert.Equal(t, "mno", record["col2"])
		assert.Equal(t, "pqr", record["col3"])

		_, err = r.Read()
		assert.ErrorIs(t, err, io.EOF)
	})

	t.Run("WithDefaultFile", func(t *testing.T) {
		tmpDir := t.TempDir()
		testutil.CreateFile(t, tmpDir, "default.csv", defaultCsv)
		testutil.CreateFile(t, tmpDir, "data.csv", dataCsv)

		defaultFile := filepath.Join(tmpDir, "default.csv")
		dataFile := filepath.Join(tmpDir, "data.csv")
		r, err := email.NewDataReader(defaultFile, dataFile)
		assert.NoError(t, err)
		assert.NotNil(t, r)

		record, err := r.Read()
		assert.NoError(t, err)
		assert.Equal(t, "abc", record["col1"])
		assert.Equal(t, "ghi", record["col2"])
		assert.Equal(t, "jkl", record["col3"])

		record, err = r.Read()
		assert.NoError(t, err)
		assert.Equal(t, "abc", record["col1"])
		assert.Equal(t, "mno", record["col2"])
		assert.Equal(t, "pqr", record["col3"])

		_, err = r.Read()
		assert.ErrorIs(t, err, io.EOF)
	})
}
