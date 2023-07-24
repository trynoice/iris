package cmd

import (
	"fmt"

	"github.com/ashutoshgngwr/iris-cli/internal/config"
	"github.com/spf13/cobra"
)

func InitCommand(v config.Viper) *cobra.Command {
	c := &cobra.Command{
		Use:   "init",
		Short: "create config file in the current working directory",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := config.WriteDefault(v); err != nil {
				return fmt.Errorf("failed to write config: %w", err)
			}

			return nil
		},
	}

	return c
}
