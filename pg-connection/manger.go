package connection

import (
	"context"
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

type DBConfigParams struct {
	DBHost     string
	DBPort     string
	DBName     string
	DBUsername string
	DBPassword string
	DBLogLevel string
}

type StorageManager interface {
	Conn(ctx context.Context) *gorm.DB
	WithTx(ctx context.Context, fc func(ctx context.Context) error) error
}

type connection struct{ conn *gorm.DB }

var _ StorageManager = (*connection)(nil)

func NewConnection(config DBConfigParams) (StorageManager, error) {
	var conn = connection{}
	const layer = "host=%s user=%s password=%s dbname=%s port=%s sslmode=disable"
	var dsn = fmt.Sprintf(layer, config.DBHost, config.DBUsername, config.DBPassword, config.DBName, config.DBPort)
	if err := conn.openConnection(dsn, config.DBLogLevel); err != nil {
		return nil, err
	}
	return &conn, nil
}

func (*connection) level(s string) logger.LogLevel {
	switch s {
	case "info":
		return logger.Info
	case "error":
		return logger.Error
	case "warn":
		return logger.Warn
	}
	return logger.Silent
}

func (c *connection) openConnection(dsn string, loglevel string) error {
	var level = c.level(loglevel)
	dialector := postgres.Open(dsn)
	var err error
	c.conn, err = gorm.Open(dialector, &gorm.Config{
		SkipDefaultTransaction: true,
		Logger:                 logger.Default.LogMode(level),
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})
	if err == nil && c.conn != nil {
		log.Println("Database connection established successfully.")
	}
	return err
}

// context
type key int

var contextConnectionKey key

func (c *connection) Conn(ctx context.Context) *gorm.DB {
	value := ctx.Value(contextConnectionKey)
	if db, ok := value.(*gorm.DB); ok {
		return db.WithContext(ctx)
	}
	return c.conn.WithContext(ctx)
}

func (c *connection) WithTx(ctx context.Context, txFunc func(ctx context.Context) error) error {
	if _, ok := ctx.Value(contextConnectionKey).(*gorm.DB); ok {
		return txFunc(ctx)
	}

	// Start a new transaction
	return c.Conn(ctx).Transaction(func(tx *gorm.DB) error {
		txCtx := context.WithValue(ctx, contextConnectionKey, tx)
		return txFunc(txCtx)
	})
}
