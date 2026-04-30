```go
package main

import (
	"cursito/internal"
	"embed"
	"log/slog"
	"os"

	"github.com/sfperusacdev/identitysdk/setup"
)

//go:embed version
var version string

//go:embed all:migrations
var migrationsDir embed.FS

func main() {
	service := setup.NewService(version,
		setup.WithMigrationSource(migrationsDir),
	)
	err := service.Run(internal.Module)
	if err != nil {
		slog.Error("Failed to run the service", "error", err)
		os.Exit(1)
	}
}
```