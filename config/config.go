package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/kelseyhightower/envconfig"
)

var ErrInvalidGoogleCreds = errors.New("invalid google application credentials")
var (
	HomePath = os.Getenv("HOME")
	Dir      = filepath.Join(HomePath, ".config", "dolores")
	File     = filepath.Join(Dir, "dolores.json")
)

type Google struct {
	ApplicationCredentials string `split_words:"true"`
	StorageBucket          string `split_words:"true"`
}

type Metadata struct {
	Bucket      string    `json:"bucket"`
	Location    string    `json:"location"`
	Environment string    `json:"environment"`
	CreatedAt   time.Time `json:"created_at"`
}

type Client struct {
	Google
}

func (c Client) BucketName() string {
	return c.Google.StorageBucket
}

func (c Client) Valid() error {
	if c.Google.ApplicationCredentials == "" {
		return ErrInvalidGoogleCreds
	}
	return nil
}

func LoadClient(env string) (Client, error) {
	var cfg Client
	d, err := LoadFromDisk()
	if err != nil {
		return Client{}, fmt.Errorf("dolores not initialized yet: %w", err)
	}
	if err := envconfig.Process("GOOGLE", &cfg.Google); err != nil {
		return Client{}, fmt.Errorf("processing config: %w", err)
	}
	bucket := d.Environments[env].Bucket
	if bucket != "" {
		cfg.Google.StorageBucket = bucket
	}
	if err := cfg.Valid(); err != nil {
		return Client{}, err
	}
	return cfg, nil
}

func MetadataFileName() string {
	return "dolores.json"
}
