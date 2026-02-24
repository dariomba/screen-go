package cmd

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"

	"github.com/dariomba/screen-go/internal/app"
	"github.com/spf13/cobra"
)

type serveCmdFlags struct {
	HttpHost string
	HttpPort string
}

func createServeCmd(ctr *app.Container) *cobra.Command {
	var serveCmdFlags serveCmdFlags

	var serveCmd = &cobra.Command{
		Use:   "serve",
		Short: "Start the screenshot API server",
		Long:  `Starts the Screen-Go HTTP server that listens for screenshot requests, manages job processing, and serves results.`,
		Run: func(cmd *cobra.Command, args []string) {
			addContainerParams(ctr, &serveCmdFlags)

			ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
			defer stop()

			go func() {
				log.Printf("Starting server on %s:%s...\n", ctr.HttpHost, ctr.HttpPort)
				if err := ctr.HTTPServer().ListenAndServe(); err != nil && err != http.ErrServerClosed {
					log.Fatalf("Server failed: %v", err)
				}
			}()

			<-ctx.Done()
			log.Println("Shutting down server...")
			if err := ctr.HTTPServer().Shutdown(context.Background()); err != nil {
				log.Fatalf("Server shutdown failed: %v", err)
			}
			log.Println("Server gracefully stopped")
		},
	}

	// Add flags for configuration
	addServeCmdFlags(serveCmd, &serveCmdFlags)

	return serveCmd
}

func addServeCmdFlags(cmd *cobra.Command, flags *serveCmdFlags) {
	cmd.Flags().StringVar(&flags.HttpHost, "http-host", "", "HTTP host to listen on")
	cmd.Flags().StringVar(&flags.HttpPort, "http-port", "8080", "HTTP port to listen on")
}

func addContainerParams(ctr *app.Container, flags *serveCmdFlags) {
	ctr.HttpHost = flags.HttpHost
	ctr.HttpPort = flags.HttpPort
}
