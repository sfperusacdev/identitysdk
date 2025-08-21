package setup

import (
	"bufio"
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/gosimple/slug"
	"github.com/pressly/goose/v3"
	"github.com/sfperusacdev/identitysdk"
	"github.com/sfperusacdev/identitysdk/configs"
	"github.com/sfperusacdev/identitysdk/helpers/facecropper"
	"github.com/sfperusacdev/identitysdk/helpers/properties"
	"github.com/sfperusacdev/identitysdk/helpers/properties/models"
	propertiesfx "github.com/sfperusacdev/identitysdk/helpers/properties/properties_fx"
	propsprovider "github.com/sfperusacdev/identitysdk/helpers/properties/props_provider"
	"github.com/sfperusacdev/identitysdk/helpers/signpdf"
	"github.com/sfperusacdev/identitysdk/helpers/sunat"
	"github.com/sfperusacdev/identitysdk/httpapi"
	connection "github.com/sfperusacdev/identitysdk/pg-connection"
	"github.com/sfperusacdev/identitysdk/services"
	"github.com/sfperusacdev/identitysdk/xreq"
	"github.com/spf13/cobra"
	"github.com/user0608/numeroaletras"
	"go.uber.org/fx"
	"gopkg.in/yaml.v2"
)

type ServiceDetails struct {
	Name        string
	Description string
}

type ServiceOptions struct {
	configProvider configs.ConfigsProviderFunc
	details        ServiceDetails
	migrationsDir  fs.FS
	propertiesDir  fs.FS
}

type ServiceOption func(*ServiceOptions)

func WithMigrationSource(sf fs.FS) ServiceOption {
	return func(o *ServiceOptions) {
		if sf == nil {
			slog.Warn("Migrations filesystem is nil, operation skipped")
			return
		}
		o.migrationsDir = sf
	}
}

func WithSystemPropertiesSource(sf fs.FS) ServiceOption {
	return func(o *ServiceOptions) {
		if sf == nil {
			slog.Warn("Properties filesystem is nil, operation skipped")
			return
		}
		o.propertiesDir = sf
	}
}

func WithConfigProvider(provider configs.ConfigsProviderFunc) ServiceOption {
	return func(o *ServiceOptions) {
		if provider == nil {
			slog.Warn("Service config provider is nil, operation skipped")
			return
		}
		o.configProvider = provider
	}
}

func WithDetails(serviceID, description string) ServiceOption {
	return func(o *ServiceOptions) {
		o.details = ServiceDetails{
			Name:        strings.TrimSpace(serviceID),
			Description: strings.TrimSpace(description),
		}
	}
}

type Service struct {
	configPath *configs.ConfigPath
	Command    *cobra.Command
	options    ServiceOptions
	version    string
}

func NewService(
	version string,
	opts ...ServiceOption,
) *Service {
	options := &ServiceOptions{
		configProvider: configs.DefaultConfigsProviderFunc,
	}
	for _, apply := range opts {
		apply(options)
	}

	service := &Service{
		version: version,
		options: *options,
	}

	service.Command = &cobra.Command{
		Use:  path.Base(os.Args[0]),
		Args: service.prepareConfigPath,
	}

	service.Command.PersistentFlags().StringP("config", "c", "", "location of the application's configuration file")
	service.Command.PersistentFlags().BoolP("auto", "a", false, "auto migrate database")

	service.Command.AddCommand(
		&cobra.Command{
			Use: "version", Short: "Print the version number",
			Run: func(cmd *cobra.Command, args []string) { fmt.Println(version) },
		},
		&cobra.Command{
			Use:   "config-example",
			Short: "Generates a base YAML configuration template for the application (config.yml)",
			Run: func(cmd *cobra.Command, args []string) {
				data, err := yaml.Marshal(&configs.GeneralServiceConfig{})
				if err != nil {
					slog.Error("Failed to marsahl yml", "error", err)
					os.Exit(1)
				}
				fmt.Println(string(data))
			},
		},
	)
	if options.migrationsDir != nil {
		service.Command.AddCommand(
			&cobra.Command{
				Use:   "database-script",
				Short: "Generates a complete SQL script to create the database schema",
				Run: func(cmd *cobra.Command, args []string) {
					const migrration_dir = "migrations"

					migrationsDir, err := fs.ReadDir(options.migrationsDir, "migrations")
					if err != nil {
						slog.Error("unable to read migrations directory", "error", err)
						os.Exit(1)
					}
					var rgx = regexp.MustCompile(`(?s)-- \+goose Up.*-- \+goose Down`)
					var output strings.Builder
					for _, entry := range migrationsDir {
						if path.Ext(entry.Name()) != ".sql" {
							continue
						}
						fileContent, err := fs.ReadFile(
							options.migrationsDir,
							path.Join(migrration_dir, entry.Name()),
						)
						if err != nil {
							slog.Error("unable to read migrations file", "file", entry.Name(), "error", err)
							os.Exit(1)
						}
						mathed := rgx.Find(fileContent)
						if mathed == nil {
							continue
						}
						output.WriteString(fmt.Sprintf("--%s\n", entry.Name()))
						skipBytes := []byte("--")
						var scanner = bufio.NewScanner(bytes.NewBuffer(mathed))
						for scanner.Scan() {
							line := scanner.Bytes()
							if bytes.HasPrefix(line, skipBytes) ||
								bytes.TrimSpace(line) == nil {
								continue
							}
							output.WriteString(string(line))
							output.WriteByte('\n')
						}
					}
					fmt.Println(output.String())
				},
			},
			service.migrationCommand("upgrade", "Upgrade the database schema to the latest version", "up"),
			service.migrationCommand("downgrade", "Downgrade the database schema to a previous version", "down"),
			service.migrationCommand("status", "Show database version status", "status"),
		)
	}
	if options.propertiesDir != nil {
		var packageName *string

		var command = &cobra.Command{
			Use: "system-props",
			Run: func(cmd *cobra.Command, args []string) {
				if packageName == nil {
					var val = "properties"
					packageName = &val
				}
				entries, err := service.readProperties(options.propertiesDir)
				if err != nil {
					slog.Error("reading properties", "error", err)
					os.Exit(1)
				}
				var str strings.Builder
				str.WriteString(fmt.Sprintf("package %s\n\n", *packageName))
				str.WriteString(`import "github.com/sfperusacdev/identitysdk/helpers/properties"`)
				str.WriteByte('\n')
				str.WriteByte('\n')
				var length = len(entries)
				for i, entry := range entries {
					var name = strings.ReplaceAll(slug.Make(entry.ID), "-", "_")
					name = strings.ToUpper(name)
					if entry.Description != "" {
						str.WriteString(fmt.Sprintf("// %s\n", entry.Description))
					}
					if entry.Type != "" {
						str.WriteString(fmt.Sprintf("// type: %s\n", entry.Type))
					}
					str.WriteString(
						fmt.Sprintf(`const %s properties.SystemProperty = "%s"`, name, entry.ID),
					)
					if i != length-1 {
						str.WriteByte('\n')
						str.WriteByte('\n')
					}
				}
				fmt.Println(str.String())
			},
		}
		packageName = command.Flags().StringP("package", "p", "properties", "Specifies the Go package name for the generated code")
		service.Command.AddCommand(command)
	}
	return service
}

func (s *Service) prepareConfigPath(cmd *cobra.Command, args []string) error {
	configPath := configs.ConfigPath(cmd.Flag("config").Value.String())
	if configPath == "" {
		return errors.New("the --config file path was not provided or is empty")
	}
	s.configPath = &configPath
	return nil
}

func (s *Service) configs() (configs.GeneralServiceConfigProvider, configs.DatabaseConfigProvider, error) {
	if s.configPath == nil {
		slog.Error("configPath is nil")
		return nil, nil, errors.New("config path is nil")
	}

	ceneralConfig, dbconfig, err := s.options.configProvider(*s.configPath)
	if err != nil {
		slog.Error("Error fetching database configuration", "error", err)
		return nil, nil, err
	}
	return ceneralConfig, dbconfig, nil
}

func (s *Service) getCconnectionManager() (connection.StorageManager, error) {
	_, c, err := s.configs()
	if err != nil {
		slog.Error("Error fetching database configs", "error", err)
		return nil, err
	}

	return connection.NewConnection(connection.DBConfigParams{
		DBHost:     c.GetHost(),
		DBPort:     fmt.Sprint(c.GetPort()),
		DBName:     c.GetDBName(),
		DBUsername: c.GetUsername(),
		DBPassword: c.GetPassword(),
		DBLogLevel: c.GetLogLevel(),
	})

}
func (s *Service) getDatabaseConnection() (*sql.DB, error) {
	connectionManager, err := s.getCconnectionManager()
	if err != nil {
		slog.Error("Error opening database connection", "error", err)
		return nil, err
	}

	conn := connectionManager.Conn(context.Background())
	db, err := conn.DB()
	if err != nil {
		slog.Error("Error recovering database connection", "error", err)
		return nil, err
	}

	return db, nil
}

func (s *Service) migrationCommand(use, shortDesc, migrationType string) *cobra.Command {
	return &cobra.Command{
		Use:   use,
		Short: shortDesc,
		Args:  s.prepareConfigPath,
		Run: func(cmd *cobra.Command, args []string) {
			if s.options.migrationsDir == nil {
				slog.Warn("Migrations directory not set, skipping migration initialization")
				return
			}
			db, err := s.getDatabaseConnection()
			if err != nil {
				slog.Error("Failed to establish database connection", "error", err)
				os.Exit(1)
			}
			goose.SetBaseFS(s.options.migrationsDir)
			err = goose.RunWithOptionsContext(context.Background(), migrationType, db, "migrations", []string{})
			if err != nil {
				slog.Error(fmt.Sprintf("Database %s failed", migrationType), "error", err)
				os.Exit(1)
			}
		},
	}
}

func (s *Service) setupIdentity(c configs.GeneralServiceConfigProvider) error {
	identitysdk.SetIdentityServer(c.Identity())
	identitysdk.SetAccessToken(c.IdentityAccessToken())

	if err := identitysdk.IdentityServerCheckHealth(); err != nil {
		slog.Error("indentity health check", "error", err)
		return err
	}
	slog.Info("Identity server OK!!")
	return nil
}

func (s *Service) publishServiceDetails(c configs.GeneralServiceConfigProvider) {
	var accessToken = c.IdentityAccessToken()
	if s.options.details.Name == "" {
		return
	}
	if accessToken == "" {
		slog.Warn("Failed to publish service details: missing access token")
		return
	}

	var buff bytes.Buffer
	if err := json.NewEncoder(&buff).Encode(map[string]string{
		"code":        s.options.details.Name,
		"description": s.options.details.Description,
	}); err != nil {
		slog.Warn("Failed to encode service details", "error", err)
		return
	}

	if err := xreq.MakeRequest(
		context.Background(),
		c.Identity(),
		"/api/v1/internal/system/resources",
		xreq.WithMethod(http.MethodPost),
		xreq.WithRequestBody(&buff),
		xreq.WithJsonContentType(),
		xreq.WithAccessToken(accessToken),
	); err != nil {
		slog.Warn("Failed to publish service details", "error", err)
	}
}

func (s *Service) Run(opts ...fx.Option) error {
	s.Command.Run = func(cmd *cobra.Command, args []string) {
		connectionManager, err := s.getCconnectionManager()
		if err != nil {
			slog.Error("Error opening database connection", "error", err)
			os.Exit(1)
		}
		gsc, _, err := s.configs()
		if err != nil {
			slog.Error("Error loading configs", "error", err)
			os.Exit(1)
		}

		automigration, err := cmd.Flags().GetBool("auto")
		if err != nil {
			slog.Error("Failed to read 'auto' flag", "error", err)
			os.Exit(1)
		}
		if automigration && s.options.migrationsDir != nil {
			ctx := context.Background()
			goose.SetBaseFS(s.options.migrationsDir)

			gormConn := connectionManager.Conn(ctx)
			conn, err := gormConn.DB()
			if err != nil {
				slog.Error("Failed to retrieve database connection", "error", err)
				os.Exit(1)
			}

			slog.Info("Running database migrations...")
			err = goose.RunWithOptionsContext(ctx, "up", conn, "migrations", []string{})
			if err != nil {
				slog.Error("Error running migrations", "error", err)
				os.Exit(1)
			}
			slog.Info("Migrations completed successfully")
		}

		var systemProperties = []models.DetailedSystemProperty{}
		if s.options.propertiesDir != nil {
			entries, err := s.readProperties(s.options.propertiesDir)
			if err != nil {
				slog.Error("Failed to retrieve system properties", "error", err)
				os.Exit(1)
			}
			systemProperties = entries
		}

		opts = append(
			opts,
			fx.Supply(systemProperties),
			fx.Provide(
				func() configs.ConfigPath {
					if s.configPath == nil {
						return ""
					}
					return *s.configPath
				},
				func() configs.GeneralServiceConfigProvider { return gsc },
				func() connection.StorageManager { return connectionManager },
				func(c configs.GeneralServiceConfigProvider) httpapi.ServeURLString {
					return httpapi.ServeURLString(c.ListenAddress())
				},
			),

			fx.Provide(services.NewExternalBridgeService),
			// tools
			fx.Provide(facecropper.NewFaceCropService),

			fx.Provide(
				fx.Annotate(
					propsprovider.NewSystemPropsPgProvider,
					fx.As(new(properties.SystemPropsProvider)),
				),
				fx.Annotate(
					numeroaletras.NewNumeroALetras,
					fx.As(new(sunat.NumeroALetras)),
				),
				fx.Annotate(
					signpdf.NewPopplerPDFInspector,
					fx.As(new(signpdf.PDFInspector)),
				),
				fx.Annotate(
					signpdf.NewPyhankoPDFSigner,
					fx.As(new(signpdf.PDFSigner)),
				),
			),
			propertiesfx.Module,
			httpapi.Module,
			fx.Invoke(s.setupIdentity, s.publishServiceDetails, httpapi.StartWebServer),
		)
		app := fx.New(opts...)
		app.Run()
		slog.Info("Application stopped")
	}
	if err := s.Command.Execute(); err != nil {
		slog.Error("command execution failed", "error", err)
		return err
	}
	return nil
}
