package email_test

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/trynoice/iris/internal/email"
	"github.com/trynoice/iris/internal/testutil"
)

func TestDataReader(t *testing.T) {
	const defaultFile = "default.csv"
	const defaultFileContent = "col1,col2\nabc,def"
	const dataFile = "data.csv"
	const dataFileContent = "col2,col3\nghi,jkl\nmno,pqr"

	t.Run("WithNonExistingDataFile", func(t *testing.T) {
		tmpDir := t.TempDir()
		testutil.CreateFile(t, tmpDir, defaultFile, defaultFileContent)

		r, err := email.NewDataReader(tmpDir, defaultFile, dataFile)
		assert.Error(t, err)
		assert.Nil(t, r)
	})

	t.Run("WithNonExistingDefaultFile", func(t *testing.T) {
		tmpDir := t.TempDir()
		testutil.CreateFile(t, tmpDir, dataFile, dataFileContent)

		r, err := email.NewDataReader(tmpDir, defaultFile, dataFile)
		assert.Error(t, err)
		assert.Nil(t, r)
	})

	t.Run("WithoutDefaultFile", func(t *testing.T) {
		tmpDir := t.TempDir()
		testutil.CreateFile(t, tmpDir, defaultFile, defaultFileContent)
		testutil.CreateFile(t, tmpDir, dataFile, dataFileContent)

		r, err := email.NewDataReader(tmpDir, "", dataFile)
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
		testutil.CreateFile(t, tmpDir, defaultFile, defaultFileContent)
		testutil.CreateFile(t, tmpDir, dataFile, dataFileContent)

		r, err := email.NewDataReader(tmpDir, defaultFile, dataFile)
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
