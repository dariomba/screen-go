//go:generate go tool github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=configs/oapi-codegen-config.yaml api/openapi.yaml
package main

import "fmt"

func main() {
	fmt.Println("Hello World!")
}
