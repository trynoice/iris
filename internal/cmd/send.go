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

			if !isDryRun && !askConfirmation("confirm sending emails?", cmd.InOrStdin(), cmd.OutOrStdout()) {
				return nil
			}

			var svc email.Service
			if isDryRun {
				svc = email.NewPrintService(cmd.OutOrStdout(), opts...)
			} else if cfg.Service.AwsSes != nil {
				if svc, err = email.NewAwsSesService(cfg.Service.AwsSes, opts...); err != nil {
					return fmt.Errorf("failed to initialise aws ses service: %w", err)
				}
			} else if cfg.Service.Smtp != nil {
				if svc, err = email.NewSmtpService(cfg.Service.Smtp, opts...); err != nil {
					return fmt.Errorf("failed to initialise smtp service: %w", err)
				}
			} else {
				return fmt.Errorf("cannot select a suitable emailing service based on the provided configuration")
			}

			defer svc.Close()

			for {
				recipientData, err := r.Read()
				if err == io.EOF {
					break
				} else if err != nil {
					return err
				}

				msg, err := t.Render(recipientData)
				if err != nil {
					return err
				}

				if err := svc.Send(&email.SendOptions{
					From:    cfg.Message.Sender,
					To:      recipientData[cfg.Message.RecipientEmailColumnName],
					ReplyTo: cfg.Message.ReplyToAddresses,
					Message: msg,
				}); err != nil {
					return err
				}
			}
			return nil
		},
	}

	c.Flags().BoolVarP(&isDryRun, "dry-run", "d", isDryRun, "print rendered emails without sending them")
	return c
}

func askConfirmation(msg string, in io.Reader, out io.Writer) bool {
	reader := bufio.NewReader(in)
	for {
		fmt.Fprintf(out, "%s [y/n] ", msg)
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
