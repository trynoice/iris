package cmd

import (
	"io"

	"github.com/ashutoshgngwr/iris-cli/internal/config"
	"github.com/ashutoshgngwr/iris-cli/internal/email"
	"github.com/spf13/cobra"
)

func SendCommand(cfg *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "send",
		Short: "Send emails using the working files in the current directory",
		RunE: func(cmd *cobra.Command, args []string) error {
			t, err := email.NewTemplate(".")
			if err != nil {
				return err
			}

			r, err := email.NewDataReader(cfg.Message.DefaultDataCsvFile, cfg.Message.RecipientDataCsvFile)
			if err != nil {
				return err
			}

			defer r.Close()
			s := email.NewPrintService(cmd.OutOrStdout(), email.WithRateLimit(1))
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
				if err := s.Send(sender, recipient, e); err != nil {
					return err
				}
			}
			return nil
		},
	}
}
