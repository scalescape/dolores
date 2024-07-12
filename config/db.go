package config

import (
	"fmt"
	"time"

	// postgres driver library
	_ "github.com/lib/pq"
)

// database configuration for server
type Database struct {
	Driver            string `default:"postgres"`
	Host              string `required:"true"`
	User              string `required:"true"`
	Password          string `required:"true"`
	Port              int    `default:"5432"`
	MaxIdleConns      int    `split_words:"true" default:"20"`
	MaxOpenConns      int    `split_words:"true" default:"30"`
	MaxConnLifetimeMs int    `split_words:"true" default:"1000"`
	Name              string `split_words:"true" required:"true"`
	SslMode           string `split_words:"true" default:"disable"`
	AesKey            string `split_words:"true" required:"true"`
}

func (db Database) MaxConnLifetime() time.Duration {
	return time.Millisecond * time.Duration(db.MaxConnLifetimeMs)
}

func (db Database) URL() string {
	return fmt.Sprintf("user=%s password=%s host=%s port=%d dbname=%s sslmode=%s", db.User, db.Password, db.Host, db.Port, db.Name, db.SslMode)
}
