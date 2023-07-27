package cmd

import (
	"fmt"
	"io"

	"github.com/ashutoshgngwr/iris-cli/internal/config"
	"github.com/ashutoshgngwr/iris-cli/internal/email"
	"github.com/spf13/cobra"
)

func SendCommand(cfg *config.Config) *cobra.Command {
	isDryRun := false
	c := &cobra.Command{
		Use:   "send",
		Short: "Send emails using the working files in the current directory",
		RunE: func(cmd *cobra.Command, args []string) error {
			t, err := email.NewTemplate(".", cfg.Message.MinifyHtml)
			if err != nil {
				return err
			}

			r, err := email.NewDataReader(cfg.Message.DefaultDataCsvFile, cfg.Message.RecipientDataCsvFile)
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
				cmd.Println("sending to", recipient)
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
