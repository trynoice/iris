package email

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/trynoice/iris/internal/testutil"
)

func TestTemplate(t *testing.T) {
	const subject = "test-subject {{ .data }}"
	const textBody = "test-text-body {{ .data }}"
	const htmlBody = `
<!doctype html>
<html>
	<body>test-html-body {{ .data }}</body>
</html>`

	t.Run("WithoutSubject", func(t *testing.T) {
		tmpDir := t.TempDir()
		testutil.CreateFile(t, tmpDir, textBodyFile, textBody)
		testutil.CreateFile(t, tmpDir, htmlBodyFile, htmlBody)

		template, err := NewTemplate(tmpDir, false)
		assert.Error(t, err)
		assert.Nil(t, template)
	})

	t.Run("WithoutTextBody", func(t *testing.T) {
		tmpDir := t.TempDir()
		testutil.CreateFile(t, tmpDir, subjectFile, subject)
		testutil.CreateFile(t, tmpDir, htmlBodyFile, htmlBody)

		template, err := NewTemplate(tmpDir, false)
		assert.Error(t, err)
		assert.Nil(t, template)
	})

	t.Run("WithoutHtmlBody", func(t *testing.T) {
		tmpDir := t.TempDir()
		testutil.CreateFile(t, tmpDir, subjectFile, subjectFile)
		testutil.CreateFile(t, tmpDir, textBodyFile, textBody)

		template, err := NewTemplate(tmpDir, false)
		assert.Error(t, err)
		assert.Nil(t, template)
	})

	t.Run("RenderWithHtmlMinification", func(t *testing.T) {
		tmpDir := t.TempDir()
		testutil.CreateFile(t, tmpDir, subjectFile, subject)
		testutil.CreateFile(t, tmpDir, textBodyFile, textBody)
		testutil.CreateFile(t, tmpDir, htmlBodyFile, htmlBody)

		template, err := NewTemplate(tmpDir, true)
		assert.NoError(t, err)
		assert.NotNil(t, template)

		m, err := template.Render(map[string]string{"data": tmpDir})
		assert.NoError(t, err)
		assert.NotNil(t, m)

		assert.Equal(t, "test-subject "+tmpDir, m.Subject)
		assert.Equal(t, "test-text-body "+tmpDir, m.TextBody)
		assert.Contains(t, m.HtmlBody, tmpDir)
		assert.NotRegexp(t, `\s\s+`, m.HtmlBody)
	})

	t.Run("RenderWithoutHtmlMinification", func(t *testing.T) {
		tmpDir := t.TempDir()
		testutil.CreateFile(t, tmpDir, subjectFile, subject)
		testutil.CreateFile(t, tmpDir, textBodyFile, textBody)
		testutil.CreateFile(t, tmpDir, htmlBodyFile, htmlBody)

		template, err := NewTemplate(tmpDir, false)
		assert.NoError(t, err)
		assert.NotNil(t, template)

		m, err := template.Render(map[string]string{"data": tmpDir})
		assert.NoError(t, err)
		assert.NotNil(t, m)

		assert.Equal(t, "test-subject "+tmpDir, m.Subject)
		assert.Equal(t, "test-text-body "+tmpDir, m.TextBody)
		assert.Contains(t, m.HtmlBody, tmpDir)
		assert.Regexp(t, `\s\s+`, m.HtmlBody)
	})
}
