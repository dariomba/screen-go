package cmd

import (
	"github.com/dariomba/screen-go/internal/app"
	"github.com/spf13/cobra"
)

func createServeCmd(ctr *app.Container) *cobra.Command {
	var serveCmd = &cobra.Command{
		Use:   "serve",
		Short: "Start the screenshot API server",
		Long:  `Starts the Screen-Go HTTP server that listens for screenshot requests, manages job processing, and serves results.`,
		Run: func(cmd *cobra.Command, args []string) {

		},
	}

	// Add persistent flags for configuration

	return serveCmd
}
