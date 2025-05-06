package setup

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path"

	"github.com/pressly/goose/v3"
	"github.com/sfperusacdev/identitysdk"
	"github.com/sfperusacdev/identitysdk/httpapi"
	"github.com/spf13/cobra"
	connection "github.com/user0608/pg-connection"
	"go.uber.org/fx"
	"gopkg.in/yaml.v2"
)

type Service struct {
	configPath     *ConfigPath
	migrationsDir  embed.FS
	version        string
	Command        *cobra.Command
	configProvider ConfigsProviderFunc
}

func NewService(
	version string,
	migrationsDir embed.FS,
) *Service {
	service := &Service{
		version:        version,
		migrationsDir:  migrationsDir,
		configProvider: DefaultConfigsProviderFunc,
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
				data, err := yaml.Marshal(&GeneralServiceConfig{})
				if err != nil {
					slog.Error("Failed to marsahl yml", "error", err)
					os.Exit(1)
				}
				fmt.Println(string(data))
			},
		},
		service.migrationCommand("upgrade", "Upgrade the database schema to the latest version", "up"),
		service.migrationCommand("downgrade", "Downgrade the database schema to a previous version", "down"),
		service.migrationCommand("status", "Show database version status", "status"),
	)
	return service
}

func (s *Service) SetConfigsProviderFunc(provider ConfigsProviderFunc) { s.configProvider = provider }

func (s *Service) prepareConfigPath(cmd *cobra.Command, args []string) error {
	configPath := ConfigPath(cmd.Flag("config").Value.String())
	if configPath == "" {
		return errors.New("the --config file path was not provided or is empty")
	}
	s.configPath = &configPath
	return nil
}

func (s *Service) configs() (GeneralServiceConfigProvider, DatabaseConfigProvider, error) {
	if s.configPath == nil {
		slog.Error("configPath is nil")
		return nil, nil, errors.New("config path is nil")
	}

	ceneralConfig, dbconfig, err := s.configProvider(*s.configPath)
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

	return connection.OpenWithConfigs(connection.DBConfigParams{
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

			db, err := s.getDatabaseConnection()
			if err != nil {
				slog.Error("Failed to establish database connection", "error", err)
				os.Exit(1)
			}
			goose.SetBaseFS(s.migrationsDir)
			err = goose.RunWithOptionsContext(context.Background(), migrationType, db, "migrations", []string{})
			if err != nil {
				slog.Error(fmt.Sprintf("Database %s failed", migrationType), "error", err)
				os.Exit(1)
			}
		},
	}
}
func (s *Service) setupIdentity(c GeneralServiceConfigProvider) error {
	identitysdk.SetIdentityServer(c.Identity())
	identitysdk.SetAccessToken(c.IdentityAccessToken())

	if err := identitysdk.IdentityServerCheckHealth(); err != nil {
		slog.Error("indentity health check", "error", err)
		return err
	}
	slog.Info("Identity server OK!!")
	return nil
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

		opts = append(
			opts,
			fx.Provide(
				func() GeneralServiceConfigProvider { return gsc },
				func() connection.StorageManager { return connectionManager },
				func(c GeneralServiceConfigProvider) httpapi.ServeURLString {
					return httpapi.ServeURLString(c.ListenAddress())
				},
			),
			httpapi.Module,
			fx.Invoke(s.setupIdentity, httpapi.StartWebServer),
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
