package cmd

import (
	"context"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/dariomba/screen-go/internal/app"
	"github.com/dariomba/screen-go/internal/logger"
	"github.com/spf13/cobra"
)

type serveCmdFlags struct {
	HttpHost string
	HttpPort string

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string

	ShutdownTimeout time.Duration
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
				logger.Info().
					Str("host", ctr.HttpHost).
					Str("port", ctr.HttpPort).
					Msg("Server is starting")
				if err := ctr.HTTPServer().ListenAndServe(); err != nil && err != http.ErrServerClosed {
					logger.Fatal().Err(err).Msg("Failed to start server")
				}
			}()

			<-ctx.Done()
			ctr.Shutdown()
		},
	}

	// Add flags for configuration
	addServeCmdFlags(serveCmd, &serveCmdFlags)

	return serveCmd
}

func addServeCmdFlags(cmd *cobra.Command, flags *serveCmdFlags) {
	cmd.Flags().StringVar(&flags.HttpHost, "http-host", "", "HTTP host to listen on")
	cmd.Flags().StringVar(&flags.HttpPort, "http-port", "8080", "HTTP port to listen on")
	cmd.Flags().StringVar(&flags.DBHost, "db-host", "localhost", "Database host")
	cmd.Flags().StringVar(&flags.DBPort, "db-port", "5432", "Database port")
	cmd.Flags().StringVar(&flags.DBUser, "db-user", "postgres", "Database user")
	cmd.Flags().StringVar(&flags.DBPassword, "db-password", "postgres", "Database password")
	cmd.Flags().StringVar(&flags.DBName, "db-name", "screengodb", "Database name")
	cmd.Flags().DurationVar(&flags.ShutdownTimeout, "shutdown-timeout", 30*time.Second, "Timeout for graceful shutdown")
}

func addContainerParams(ctr *app.Container, flags *serveCmdFlags) {
	ctr.HttpHost = flags.HttpHost
	ctr.HttpPort = flags.HttpPort

	ctr.DBHost = flags.DBHost
	ctr.DBPort = flags.DBPort
	ctr.DBUser = flags.DBUser
	ctr.DBPassword = flags.DBPassword
	ctr.DBName = flags.DBName

	ctr.ShutdownTimeout = flags.ShutdownTimeout
}
