//go:generate go tool github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=configs/oapi-codegen-config.yaml api/openapi.yaml
//go:generate go tool github.com/sqlc-dev/sqlc/cmd/sqlc generate -f configs/sqlc.yaml
package main

import (
	"github.com/dariomba/screen-go/cmd"
	"github.com/dariomba/screen-go/internal/app"
	"github.com/dariomba/screen-go/internal/logger"
)

func main() {
	ctr := app.NewContainer()
	rootCmd := cmd.NewRootCmd(ctr)

	if err := rootCmd.Execute(); err != nil {
		logger.Fatal().Err(err).Msg("Application failed to start")
	}
}
