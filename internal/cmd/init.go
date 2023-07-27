package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ashutoshgngwr/iris-cli/internal/config"
	"github.com/spf13/cobra"
)

var defaultEmailFiles = map[string]string{
	"subject.txt": "Hello {{ .Name }}",
	"body.txt": `Iris is a CLI tool for sending templated bulk emails.

You can inject data into templates, e.g. a date - {{ .Date }} or your email - {{ .Recipient }}.`,

	"body.html": `<!DOCTYPE html>
<html>
  <head>
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <meta http-equiv="Content-Type" content="text/html; charset=UTF-8" />
    <title>Hello {{ .Name }}</title>
  </head>
  <body>
    <p>Iris is a CLI tool for sending templated bulk emails.</p>
    <p>
      You can inject data into templates, e.g. a date - {{ .Date }} or your
      email - {{ .Recipient }}.
    </p>
  </body>
</html>`,

	"default.csv":   "Date\nJanuary 2006",
	"recipient.csv": "Name,Recipient\nJohn,hello@example.test",
}

func InitCommand(configFileName string) *cobra.Command {
	c := &cobra.Command{
		Use:   "init <dir>",
		Short: "Create working files in the given directory",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if notExists(args[0]) {
				cmd.Println("creating directory", args[0])
				if err := os.MkdirAll(args[0], os.ModeDir|os.ModePerm); err != nil {
					return fmt.Errorf("failed to create directory: %w", err)
				}
			}

			cfgFile := filepath.Join(args[0], configFileName)
			if notExists(cfgFile) {
				cmd.Println("creating file", cfgFile)
				if err := config.WriteDefault(cfgFile); err != nil {
					return fmt.Errorf("failed to write default config: %w", err)
				}
			}

			for name, content := range defaultEmailFiles {
				file := filepath.Join(args[0], name)
				if notExists(file) {
					cmd.Println("creating file", file)
					if err := os.WriteFile(file, []byte(content), os.ModePerm); err != nil {
						return fmt.Errorf("failed to create %s: %w", file, err)
					}
				}

			}

			return nil
		},
	}

	return c
}

func notExists(file string) bool {
	_, err := os.Stat(file)
	return err != nil
}
