package configs

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/spf13/viper"
)

type ConfigPath string

type GeneralServiceConfigProvider interface {
	ServiceID() string
	ListenAddress() string
	GRPCAddress() string
	RabbitMQURL() string
	Identity() string
	IdentityAccessToken() string
	CacheDir() string
	StagingDir() string // area temporal para almacenar los archivos subidos
}

type DatabaseConfigProvider interface {
	GetHost() string
	GetPort() int
	GetUsername() string
	GetPassword() string
	GetDBName() string
	GetLogLevel() string
}

type ConfigsProviderFunc func(configPath ConfigPath) (GeneralServiceConfigProvider, DatabaseConfigProvider, *viper.Viper, error)

type GeneralServiceConfig struct {
	ListenAddressValue       string         `mapstructure:"address" yaml:"address"`
	GRPCAddressValue         string         `mapstructure:"grpc_address" yaml:"grpc_address"`
	RabbitMQURLValue         string         `mapstructure:"rabbitmq_url" yaml:"rabbitmq_url"`
	IdentityValue            string         `mapstructure:"identity" yaml:"identity"`
	IdentityAccessTokenValue string         `mapstructure:"identity_access_token" yaml:"identity_access_token"`
	CacheDirVal              string         `mapstructure:"cache_dir" yaml:"cache_dir"`
	StagingDirVal            string         `mapstructure:"staging_dir" yaml:"staging_dir"`
	DatabaseEntity           DatabaseConfig `mapstructure:"database" yaml:"database"`
}

type DatabaseConfig struct {
	Host     string `mapstructure:"host" yaml:"host"`
	Port     int    `mapstructure:"port" yaml:"port"`
	DBName   string `mapstructure:"db_name" yaml:"db_name"`
	Username string `mapstructure:"username" yaml:"username"`
	Password string `mapstructure:"password" yaml:"password"`
	LogLevel string `mapstructure:"logLevel" yaml:"logLevel"`
}

var _ GeneralServiceConfigProvider = (*GeneralServiceConfig)(nil)
var _ DatabaseConfigProvider = (*GeneralServiceConfig)(nil)

var service_id string

func SetServiceId(values string) {
	service_id = values
}

// ListenAddress implements GeneralServiceConfigProvider.
func (c *GeneralServiceConfig) ServiceID() string {
	return service_id
}

// ListenAddress implements GeneralServiceConfigProvider.
func (c *GeneralServiceConfig) ListenAddress() string {
	return c.ListenAddressValue
}

// GRPCAddress implements GeneralServiceConfigProvider.
func (c *GeneralServiceConfig) GRPCAddress() string {
	return c.GRPCAddressValue
}

// GRPCAddress implements GeneralServiceConfigProvider.
func (c *GeneralServiceConfig) RabbitMQURL() string {
	return c.RabbitMQURLValue
}

// Identity implements GeneralServiceConfigProvider.
func (c *GeneralServiceConfig) Identity() string {
	if c.IdentityValue == "" {
		c.IdentityValue = "https:api.identity2.sfperusac.com"
	}
	return c.IdentityValue
}

// CacheDir implements GeneralServiceConfigProvider.
func (c *GeneralServiceConfig) CacheDir() string {
	if c.CacheDirVal == "" {
		return ".cache"
	}
	return strings.TrimSpace(c.CacheDirVal)
}

// CacheDir implements GeneralServiceConfigProvider.
func (c *GeneralServiceConfig) StagingDir() string {
	if c.StagingDirVal == "" {
		return ".staging"
	}
	return strings.TrimSpace(c.StagingDirVal)
}

// IdentityAccessToken implements GeneralServiceConfigProvider.
func (c *GeneralServiceConfig) IdentityAccessToken() string {
	return c.IdentityAccessTokenValue
}

// GetDBName implements DatabaseConfigProvider.
func (c *GeneralServiceConfig) GetDBName() string {
	return c.DatabaseEntity.DBName
}

// GetHost implements DatabaseConfigProvider.
func (c *GeneralServiceConfig) GetHost() string {
	return c.DatabaseEntity.Host
}

// GetLogLevel implements DatabaseConfigProvider.
func (c *GeneralServiceConfig) GetLogLevel() string {
	return c.DatabaseEntity.LogLevel
}

// GetPassword implements DatabaseConfigProvider.
func (c *GeneralServiceConfig) GetPassword() string {
	return c.DatabaseEntity.Password
}

// GetPort implements DatabaseConfigProvider.
func (c *GeneralServiceConfig) GetPort() int {
	return c.DatabaseEntity.Port
}

// GetUsername implements DatabaseConfigProvider.
func (c *GeneralServiceConfig) GetUsername() string {
	return c.DatabaseEntity.Username
}

func DefaultConfigsProviderFunc(configPath ConfigPath) (GeneralServiceConfigProvider, DatabaseConfigProvider, *viper.Viper, error) {
	var stringConfigPath = strings.TrimSpace(string(configPath))
	if configPath == "" {
		slog.Error("configPath is empty")
		return nil, nil, nil, errors.New("config path is empty")
	}

	v := viper.New()
	v.SetConfigFile(stringConfigPath)

	if err := v.ReadInConfig(); err != nil {
		return nil, nil, nil, fmt.Errorf("read config %q: %w", stringConfigPath, err)
	}

	var c GeneralServiceConfig
	if err := v.Unmarshal(&c); err != nil {
		slog.Error("unmarshal config", "error", err)
		return nil, nil, nil, err
	}

	return &c, &c, v, nil
}
