package config

import (
	"errors"
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

var ErrInvalidGoogleCreds = errors.New("invalid google application credentials")

type Google struct {
	ApplicationCredentials string `split_words:"true"`
	StorageBucket          string `split_words:"true" required:"true"`
}

type Metadata struct {
	Bucket    string            `json:"bucket"`
	Locations map[string]string `json:"locations"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
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

func LoadClient() (Client, error) {
	var cfg Client
	if err := envconfig.Process("GOOGLE", &cfg.Google); err != nil {
		return Client{}, fmt.Errorf("processing config: %w", err)
	}
	if err := cfg.Valid(); err != nil {
		return Client{}, err
	}
	return cfg, nil
}
