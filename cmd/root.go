// Package cmd contains the command-line interface implementation for the markdown-to-pdf application.
package cmd

import (
	"github.com/alexhokl/helper/cli"
	"github.com/spf13/cobra"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:          "markdown-to-pdf",
	Short:        "A CLI application to convert markdown files to PDF",
	SilenceUsage: true,
}

func Execute() {
	_ = rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.markdown-to-pdf.yaml)")
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func initConfig() {
	cli.ConfigureViper(cfgFile, "markdown-to-pdf", false, "")
}
