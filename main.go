package main

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/trynoice/iris/internal/cmd"
)

const (
	configName = ".iris"
	configType = "yaml"
)

func main() {
	v := viper.New()
	v.SetConfigName(configName)
	v.SetConfigType(configType)

	rootCmd := &cobra.Command{
		Use:          "iris",
		Short:        "A CLI tool for dispatching templated emails",
		SilenceUsage: true,
	}

	rootCmd.AddCommand(cmd.InitCommand(v, configName+"."+configType))
	rootCmd.AddCommand(cmd.SendCommand(v))

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
