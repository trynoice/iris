package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/trynoice/iris/internal/config"
)

var defaultConfig = &config.Config{
	Service: config.ServiceConfig{
		AwsSes: &config.AwsSesServiceConfig{
			UseSharedConfig: true,
		},
		RateLimit: 10,
		Retries:   3,
	},
	Message: config.MessageConfig{
		Sender:                   "Iris CLI <iris@example.test>",
		ReplyToAddresses:         []string{"inbox@example.test", "another@example.test"},
		DefaultDataCsvFile:       "default.csv",
		RecipientDataCsvFile:     "recipients.csv",
		RecipientEmailColumnName: "Email",
		MinifyHtml:               true,
	},
}

var defaultEmailFiles = map[string]string{
	"subject.txt": "Hello {{.Name}}",
	"body.txt": `Iris is a CLI tool for sending templated bulk emails.

You can inject data into templates, e.g. a date - {{.Date}} or your email - {{.Email}}.`,

	"body.html": `<!DOCTYPE html>
<html>
  <head>
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <meta http-equiv="Content-Type" content="text/html; charset=UTF-8" />
    <title>Hello {{.Name}}</title>
  </head>
  <body>
    <p>Iris is a CLI tool for sending templated bulk emails.</p>
    <p>
      You can inject data into templates, e.g. a date - {{.Date}} or your
      email - {{.Email}}.
    </p>
  </body>
</html>`,

	"default.csv":    "Date\nJanuary 2006",
	"recipients.csv": "Name,Email\nJack,jack@example.test\nJill,jill@example.test",
}

func InitCommand(v *viper.Viper, configFileName string) *cobra.Command {
	c := &cobra.Command{
		Use:   "init [dir]",
		Short: "Create working files in the given directory",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			wd := "."
			if len(args) > 0 {
				wd = args[0]
			}

			v.AddConfigPath(wd)
			cfg, err := config.Read(v)
			if err != nil {
				return err
			}

			if notExists(wd) {
				cmd.Println("creating directory", wd)
				if err := os.MkdirAll(wd, os.ModeDir|os.ModePerm); err != nil {
					return fmt.Errorf("failed to create directory: %w", err)
				}
			}

			cfgFile := filepath.Join(wd, configFileName)
			if notExists(cfgFile) {
				// consider this default configuration for generating all files
				// if user didn't supply a config.
				cfg = defaultConfig
				cmd.Println("creating file", cfgFile)
				if err := config.Write(defaultConfig, cfgFile); err != nil {
					return fmt.Errorf("failed to write default config: %w", err)
				}
			}

			for name, content := range defaultEmailFiles {
				switch name {
				case "default.csv":
					name = cfg.Message.DefaultDataCsvFile
					content = strings.ReplaceAll(content, "Email", cfg.Message.RecipientEmailColumnName)
				case "recipients.csv":
					name = cfg.Message.RecipientDataCsvFile
					content = strings.ReplaceAll(content, "Email", cfg.Message.RecipientEmailColumnName)
				default:
					replacement := fmt.Sprintf("{{.%s}}", cfg.Message.RecipientEmailColumnName)
					content = strings.ReplaceAll(content, "{{.Email}}", replacement)
				}

				if name == "" { // defaults csv is optional
					continue
				}

				file := filepath.Join(wd, name)
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
