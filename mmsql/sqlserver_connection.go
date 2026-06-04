package mmsql

import (
	"context"
	"fmt"
	"log"
	"strings"

	connection "github.com/sfperusacdev/identitysdk/pg-connection"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

type SQLServerConfig struct {
	DBHost     string
	DBPort     string
	DBInstance string
	DBName     string
	DBUsername string
	DBPassword string
	DBLogLevel string
}

type SQLServerConnection struct {
	db *gorm.DB
}

var _ connection.StorageManager = (*SQLServerConnection)(nil)

func NewSQLServerConnection(config SQLServerConfig) (*SQLServerConnection, error) {
	conn := SQLServerConnection{}

	dsn, err := buildDSN(config)
	if err != nil {
		return nil, err
	}

	if err := conn.openConnection(dsn, config.DBLogLevel); err != nil {
		return nil, err
	}

	return &conn, nil
}

func buildDSN(config SQLServerConfig) (string, error) {
	host := strings.TrimSpace(config.DBHost)
	port := strings.TrimSpace(config.DBPort)
	instance := strings.TrimSpace(config.DBInstance)
	dbName := strings.TrimSpace(config.DBName)
	username := strings.TrimSpace(config.DBUsername)
	password := strings.TrimSpace(config.DBPassword)

	if host == "" {
		return "", fmt.Errorf("DBHost is required")
	}
	if dbName == "" {
		return "", fmt.Errorf("DBName is required")
	}
	if username == "" {
		return "", fmt.Errorf("DBUsername is required")
	}

	server := host
	if instance != "" {
		server = fmt.Sprintf("%s/%s", host, instance)
	} else if port != "" {
		server = fmt.Sprintf("%s:%s", host, port)
	}

	return fmt.Sprintf(
		"sqlserver://%s:%s@%s?database=%s",
		username,
		password,
		server,
		dbName,
	), nil
}

func (*SQLServerConnection) level(s string) logger.LogLevel {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "info":
		return logger.Info
	case "error":
		return logger.Error
	case "warn":
		return logger.Warn
	}
	return logger.Silent
}

func (c *SQLServerConnection) openConnection(dsn string, loglevel string) error {
	level := c.level(loglevel)

	db, err := gorm.Open(sqlserver.Open(dsn), &gorm.Config{
		SkipDefaultTransaction: true,
		Logger:                 logger.Default.LogMode(level),
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})
	if err != nil {
		return err
	}

	c.db = db
	log.Println("Database SQL Server connection established successfully.")

	return nil
}

type txCtxKey struct{}

func (c *SQLServerConnection) Conn(ctx context.Context) *gorm.DB {
	if tx, ok := ctx.Value(txCtxKey{}).(*gorm.DB); ok && tx != nil {
		return tx.WithContext(ctx)
	}

	return c.db.WithContext(ctx)
}

func (c *SQLServerConnection) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	if fn == nil {
		return nil
	}

	if _, ok := ctx.Value(txCtxKey{}).(*gorm.DB); ok {
		return fn(ctx)
	}

	return c.Conn(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(context.WithValue(ctx, txCtxKey{}, tx))
	})
}
