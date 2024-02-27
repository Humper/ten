package psql

import (
	"context"
	"fmt"
	"time"

	"github.com/cenkalti/backoff/v4"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/humper/tor_exit_nodes/models"
	"github.com/humper/tor_exit_nodes/pkg/database"
	"github.com/humper/tor_exit_nodes/pkg/util"
)

const (
	// dbConnectTimeout is the maximum time allowed for a db connection.
	dbConnectTimeout = 1 * time.Minute
)

// Configuration holds all configuration. Some fields may contain sensitive data.

type Configuration struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
	SSLMode  string `yaml:"sslmode"`
}

// Load opens a crdb database from a yaml file. Convenience overload.
func Load(ctx context.Context, filename string) (*database.Database, error) {
	cfg, err := util.ReadYamlFile[Configuration](filename)
	if err != nil {
		return nil, err
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s", cfg.Host, cfg.User, cfg.Password, cfg.DBName, cfg.Port, cfg.SSLMode)

	b := util.NewBackoff(dbConnectTimeout)

	var gormDB *gorm.DB

	err = backoff.Retry(func() error {
		gormDB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			return err
		}
		return nil
	}, b)

	if err != nil {
		return nil, err
	}

	err = gormDB.AutoMigrate(&models.User{}, &models.TorExitNode{})
	if err != nil {
		return nil, err
	}

	// hack to bootstrap the database
	u := &users{db: gormDB}
	_, err = u.GetAll(ctx, &models.Pagination{Limit: 1})

	return &database.Database{
		Users:        u,
		TorExitNodes: &torExitNodes{db: gormDB},
	}, nil
}
