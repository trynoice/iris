package cmd

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/trynoice/iris/internal/config"
	"github.com/trynoice/iris/internal/email"
)

func SendCommand(v *viper.Viper) *cobra.Command {
	isDryRun := false
	c := &cobra.Command{
		Use:   "send [dir]",
		Short: "Send emails using the working files in the current directory",
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

			t, err := email.NewTemplate(wd, cfg.Message.MinifyHtml)
			if err != nil {
				return err
			}

			r, err := email.NewDataReader(wd, cfg.Message.DefaultDataCsvFile, cfg.Message.RecipientDataCsvFile)
			if err != nil {
				return err
			}

			defer r.Close()
			opts := []email.ServiceOption{
				email.WithRateLimit(cfg.Service.RateLimit),
				email.WithRetries(cfg.Service.Retries),
			}

			var svc email.Service
			if isDryRun {
				svc = email.NewPrintService(cmd.OutOrStdout(), opts...)
			} else if cfg.Service.AwsSes != nil {
				if svc, err = email.NewAwsSesService(cfg.Service.AwsSes, opts...); err != nil {
					return fmt.Errorf("failed to initialise aws ses client: %w", err)
				}
			} else {
				return fmt.Errorf("cannot select a suitable emailing service based on the provided configuration")
			}

			if !isDryRun && !askSendEmailConfirmation(cmd.InOrStdin(), cmd.OutOrStdout()) {
				return nil
			}

			for {
				recipientData, err := r.Read()
				if err == io.EOF {
					break
				} else if err != nil {
					return err
				}

				e, err := t.Render(recipientData)
				if err != nil {
					return err
				}

				sender := cfg.Message.Sender
				recipient := recipientData[cfg.Message.RecipientEmailColumnName]
				cmd.Println("dispatching to", recipient)
				if err := svc.Send(sender, recipient, e); err != nil {
					return err
				}
			}
			return nil
		},
	}

	c.Flags().BoolVarP(&isDryRun, "dry-run", "d", isDryRun, "print rendered emails without sending them")
	return c
}

func askSendEmailConfirmation(in io.Reader, out io.Writer) bool {
	reader := bufio.NewReader(in)
	for {
		fmt.Fprint(out, "confirm sending emails? [y/n] ")
		answer, _ := reader.ReadString('\n')
		answer = strings.ToLower(strings.TrimSpace(answer))
		if answer == "y" || answer == "yes" {
			return true
		} else if answer == "n" || answer == "no" {
			return false
		} else {
			continue
		}
	}
}
