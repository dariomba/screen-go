//go:generate go tool github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=configs/oapi-codegen-config.yaml api/openapi.yaml
//go:generate go tool github.com/sqlc-dev/sqlc/cmd/sqlc generate -f configs/sqlc.yaml
package main

import (
	"log"

	"github.com/dariomba/screen-go/cmd"
	"github.com/dariomba/screen-go/internal/app"
)

func main() {
	ctr := app.NewContainer()
	rootCmd := cmd.NewRootCmd(ctr)

	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("Error executing command: %v", err)
	}
}
