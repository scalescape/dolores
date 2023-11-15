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
	ErrInvalidGoogleCreds    = errors.New("invalid google application credentials")
	ErrInvalidStorageBucket  = errors.New("invalid storage bucket")
	ErrInvalidKeyFile        = errors.New("invalid key file")
	ErrCloudProviderNotFound = errors.New("cloud provider not found")
)

var (
	AWS = "AWS"
	GCS = "GCS"
)

type CtxKey string

var EnvKey CtxKey = "ctx_environment"

var (
	HomePath = os.Getenv("HOME")
	Dir      = filepath.Join(HomePath, ".config", "dolores")
	File     = filepath.Join(Dir, "dolores.json")
)

type Cloud struct {
	ApplicationCredentials string `split_words:"true"`
	StorageBucket          string `split_words:"true"`
	StoragePrefix          string
}

type Metadata struct {
	CloudProvider          string    `json:"cloud_provider"`
	Bucket                 string    `json:"bucket"`
	Location               string    `json:"location"`
	Environment            string    `json:"environment"`
	CreatedAt              time.Time `json:"created_at"`
	ApplicationCredentials string    `json:"application_credentials"`
}

type Client struct {
	Cloud
	Provider string
}

func (c Client) BucketName() string {
	return c.Cloud.StorageBucket
}

func (c Client) Valid() error {
	if c.Provider == "" {
		return ErrCloudProviderNotFound
	}
	if c.Cloud.ApplicationCredentials == "" {
		return ErrInvalidGoogleCreds
	}
	if c.Cloud.StorageBucket == "" {
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
	if err := envconfig.Process("GOOGLE", &cfg.Cloud); err != nil {
		return Client{}, fmt.Errorf("processing config: %w", err)
	}

	md := d.Environments[env].Metadata
	if cloudProvider := md.CloudProvider; cloudProvider != "" {
		cfg.Provider = cloudProvider
	}

	if cfg.Cloud.ApplicationCredentials == "" {
		if creds := md.ApplicationCredentials; creds != "" {
			cfg.Cloud.ApplicationCredentials = creds
		}
	}

	if bucket := md.Bucket; bucket != "" {
		cfg.Cloud.StorageBucket = bucket
	}

	if location := md.Location; location != "" {
		cfg.Cloud.StoragePrefix = location
	}

	if err := cfg.Valid(); err != nil {
		return Client{}, err
	}
	return cfg, nil
}

func MetadataFileName() string {
	return "dolores.json"
}
