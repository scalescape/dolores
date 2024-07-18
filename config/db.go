package config

import (
	"fmt"
	"time"

	// postgres driver library.
	_ "github.com/lib/pq"
)

// database configuration for server.
type Database struct {
	Driver            string `default:"postgres"`
	Host              string `required:"true"`
	User              string `required:"true"`
	Password          string `required:"true"`
	Port              int    `default:"5432"`
	MaxIdleConns      int    `default:"20"       split_words:"true"`
	MaxOpenConns      int    `default:"30"       split_words:"true"`
	MaxConnLifetimeMs int    `default:"1000"     split_words:"true"`
	Name              string `required:"true"    split_words:"true"`
	SslMode           string `default:"disable"  split_words:"true"`
	AesKey            string `required:"true"    split_words:"true"`
}

func (db Database) MaxConnLifetime() time.Duration {
	return time.Millisecond * time.Duration(db.MaxConnLifetimeMs)
}

func (db Database) URL() string {
	return fmt.Sprintf("user=%s password=%s host=%s port=%d dbname=%s sslmode=%s", db.User, db.Password, db.Host, db.Port, db.Name, db.SslMode)
}
