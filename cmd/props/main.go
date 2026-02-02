package main

import (
	"log/slog"
	"os"

	"github.com/sfperusacdev/identitysdk/setup"
	"github.com/spf13/cobra"
)

func main() {
	var propertiesPath string
	var packageName string

	root := &cobra.Command{
		Use: "system-props",
		Run: func(cmd *cobra.Command, args []string) {
			fsys := os.DirFS(propertiesPath)

			service := &setup.Service{}

			err := service.GenerateSystemProps(fsys, packageName, os.Stdout)
			if err != nil {
				slog.Error("generating system props", "error", err)
				os.Exit(1)
			}
		},
	}

	root.Flags().StringVarP(&propertiesPath, "properties", "r", "_properties", "Properties directory")
	root.Flags().StringVarP(&packageName, "package", "p", "systemprops", "Package name")

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
