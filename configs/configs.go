package configs

import (
	"errors"
	"log/slog"
	"path"
	"strings"

	"github.com/spf13/viper"
)

type ConfigPath string

type GeneralServiceConfigProvider interface {
	ListenAddress() string
	Identity() string
	IdentityAccessToken() string
	CacheDir() string
}

type DatabaseConfigProvider interface {
	GetHost() string
	GetPort() int
	GetUsername() string
	GetPassword() string
	GetDBName() string
	GetLogLevel() string
}

type ConfigsProviderFunc func(configPath ConfigPath) (GeneralServiceConfigProvider, DatabaseConfigProvider, error)

type GeneralServiceConfig struct {
	ListenAddressValue       string         `mapstructure:"address" yaml:"address"`
	IdentityValue            string         `mapstructure:"identity" yaml:"identity"`
	IdentityAccessTokenValue string         `mapstructure:"identity_access_token" yaml:"identity_access_token"`
	CacheDirVal              string         `mapstructure:"cache_dir" yaml:"cache_dir"`
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

// ListenAddress implements GeneralServiceConfigProvider.
func (c *GeneralServiceConfig) ListenAddress() string {
	return c.ListenAddressValue
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

func (c *GeneralServiceConfig) validate() error {
	// TODO agregar validaciones específicas si es necesario
	return nil
}

func DefaultConfigsProviderFunc(configPath ConfigPath) (GeneralServiceConfigProvider, DatabaseConfigProvider, error) {
	var stringConfigPath = strings.TrimSpace(string(configPath))
	var c GeneralServiceConfig
	if configPath == "" {
		slog.Error("configPath is empty")
		return nil, nil, errors.New("config path is empty")
	}
	v := viper.New()
	v.SetConfigFile(stringConfigPath)
	v.SetConfigType("yaml")
	if err := v.ReadInConfig(); err != nil {
		slog.Error("reading config", "file", path.Base(stringConfigPath), "error", err)
		return nil, nil, err
	}
	if err := v.Unmarshal(&c); err != nil {
		slog.Error("unmarshal config", "file", path.Base(stringConfigPath), "error", err)
		return nil, nil, err
	}

	if err := c.validate(); err != nil {
		return nil, nil, err
	}
	return &c, &c, nil
}
