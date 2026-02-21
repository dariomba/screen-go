package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "screen-go",
	Short: "Memory-aware screenshot service with async job processing",
	Long: `Screen-Go is a self-hosted screenshot-as-a-service API.

It renders website screenshots via headless Chrome with intelligent memory management
that prevents OOM crashes. Jobs are queued when memory is scarce and processed as 
capacity becomes available, making it safe to deploy with predictable resource limits.

Features:
  • Async job processing with status polling
  • PNG and PDF output formats  
  • Full-page capture support
  • Memory-aware dispatcher using weighted semaphore
  • Postgres-backed job persistence
  • Configurable memory budgets and timeouts`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./config/config.yaml)")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Search config in config/config.yaml
		viper.AddConfigPath("./config")
		viper.SetConfigType("yaml")
		viper.SetConfigName("config")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
