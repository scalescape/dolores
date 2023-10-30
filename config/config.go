package config

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/kelseyhightower/envconfig"
)

var (
	ErrInvalidGoogleCreds   = errors.New("invalid google application credentials")
	ErrInvalidStorageBucket = errors.New("invalid storage bucket")
)

type CtxKey string

var (
	EnvKey CtxKey = "ctx_environment"
)

var (
	HomePath = os.Getenv("HOME")
	Dir      = filepath.Join(HomePath, ".config", "dolores")
	File     = filepath.Join(Dir, "dolores.json")
)

type Google struct {
	ApplicationCredentials string `split_words:"true"`
	StorageBucket          string `split_words:"true"`
	StoragePrefix          string
}

type Metadata struct {
	Bucket                 string    `json:"bucket"`
	Location               string    `json:"location"`
	Environment            string    `json:"environment"`
	CreatedAt              time.Time `json:"created_at"`
	ApplicationCredentials string    `json:"application_credentials"`
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
	if c.Google.StorageBucket == "" {
		return ErrInvalidStorageBucket
	}
	return nil
}

func LoadClient(ctx context.Context, env string) (Client, error) {
	var cfg Client
	d, err := LoadFromDisk()
	if err != nil {
		return Client{}, fmt.Errorf("dolores not initialized yet: %w", err)
	}
	if err := envconfig.Process("GOOGLE", &cfg.Google); err != nil {
		return Client{}, fmt.Errorf("processing config: %w", err)
	}

	md := d.Environments[env].Metadata
	if cfg.Google.ApplicationCredentials == "" {
		if creds := md.ApplicationCredentials; creds != "" {
			cfg.Google.ApplicationCredentials = creds
		}
	}

	if bucket := md.Bucket; bucket != "" {
		cfg.Google.StorageBucket = bucket
		cfg.Google.StoragePrefix = md.Location
	}
	if err := cfg.Valid(); err != nil {
		return Client{}, err
	}
	return cfg, nil
}

func MetadataFileName() string {
	return "dolores.json"
}
