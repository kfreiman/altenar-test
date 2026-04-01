/*
Copyright © 2026 Casino
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "casino",
	Short: "Transaction processing service",
	Long: `A service for processing casino transactions via HTTP API or Kafka consumer.

Available commands:
  api       Start the HTTP API server
  consumer  Start the Kafka consumer`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Bind environment variables with prefix CASINO_
	viper.SetEnvPrefix("casino")
	viper.AutomaticEnv()

	// Global flags
	rootCmd.PersistentFlags().StringP("config", "c", "", "config file path (default: $HOME/.casino.yaml)")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "enable verbose output")
}
