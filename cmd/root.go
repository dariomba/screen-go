package cmd

import (
	"fmt"

	"github.com/dariomba/screen-go/internal/app"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	defaultConfigDirPath = "./configs"
	defaultConfigFile    = "config"
	envPrefix            = "SCREENGO_API"
)

func NewRootCmd(ctr *app.Container) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "screen-go",
		Short: "Memory-aware screenshot service with async job processing",
		Long: `Screen-Go is a self-hosted screenshot-as-a-service API.

It renders website screenshots via headless Chrome with intelligent memory management
that prevents OOM crashes. Jobs are queued when memory is scarce and processed as 
capacity becomes available, making it safe to deploy with predictable resource limits.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return initViper()
		},
	}

	// Add subcommands
	rootCmd.AddCommand(createServeCmd(ctr))

	return rootCmd
}

func initViper() error {
	// Search for config.yaml in the current directory
	viper.AddConfigPath(defaultConfigDirPath)
	viper.SetConfigName(defaultConfigFile)
	viper.SetConfigType("yaml")

	// Enable environment variable support
	viper.SetEnvPrefix(envPrefix)
	viper.AutomaticEnv()

	// Read config file if it exists (not required)
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}

	return nil
}
