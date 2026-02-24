package cmd

import (
	"fmt"
	"strings"

	"github.com/dariomba/screen-go/internal/app"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	defaultConfigDirPath       = "./configs"
	defaultConfigFile          = "config"
	envPrefix                  = "SCREENGO_API"
	replaceHyphenWithCamelCase = false
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
			return initViper(cmd)
		},
	}

	// Add subcommands
	rootCmd.AddCommand(createServeCmd(ctr))

	return rootCmd
}

func initViper(cmd *cobra.Command) error {
	v := viper.New()

	// Set the base name of the config file, without the file extension.
	v.SetConfigName(defaultConfigFile)

	// Set as many paths as you like where viper should look for the
	// config file. We are only looking in the current working directory.
	v.AddConfigPath(defaultConfigDirPath)

	// Attempt to read the config file, gracefully ignoring errors
	// caused by a config file not being found. Return an error
	// if we cannot parse the config file.
	if err := v.ReadInConfig(); err != nil {
		// It's okay if there isn't a config file
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return err
		}
	}

	// When we bind flags to environment variables expect that the
	// environment variables are prefixed, e.g. a flag like --number
	// binds to an environment variable STING_NUMBER. This helps
	// avoid conflicts.
	v.SetEnvPrefix(envPrefix)

	// Environment variables can't have dashes in them, so bind them to their equivalent
	// keys with underscores, e.g. --favorite-color to STING_FAVORITE_COLOR
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	// Bind to environment variables
	// Works great for simple config names, but needs help for names
	// like --favorite-color which we fix in the bindFlags function
	v.AutomaticEnv()

	// Bind the current command's flags to viper
	bindFlags(cmd, v)

	return nil
}

// Bind each cobra flag to its associated viper configuration (config file and environment variable)
func bindFlags(cmd *cobra.Command, v *viper.Viper) {
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		// Determine the naming convention of the flags when represented in the config file
		configName := f.Name

		// If using camelCase in the config file, replace hyphens with a camelCased string.
		// Since viper does case-insensitive comparisons, we don't need to bother fixing the case, and only need to remove the hyphens.
		if replaceHyphenWithCamelCase {
			configName = strings.ReplaceAll(f.Name, "-", "")
		}

		// Apply the viper config value to the flag when the flag is not set and viper has a value
		if !f.Changed && v.IsSet(configName) {
			val := v.Get(configName)
			cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val))
		}
	})
}
