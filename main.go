package main

import (
	"fmt"
	"os"

	"github.com/ashutoshgngwr/iris-cli/internal/cmd"
	"github.com/ashutoshgngwr/iris-cli/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	configName = ".iris"
	configType = "yaml"
)

func main() {
	v := viper.New()
	v.SetConfigName(configName)
	v.SetConfigType(configType)
	v.AddConfigPath(".")
	cfg, err := config.Read(v)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	rootCmd := &cobra.Command{
		Use:          "iris",
		Short:        "A CLI tool for dispatching templated emails",
		SilenceUsage: true,
	}

	rootCmd.AddCommand(cmd.InitCommand(configName + "." + configType))
	rootCmd.AddCommand(cmd.SendCommand(cfg))

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
