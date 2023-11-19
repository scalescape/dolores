package config

import (
	"errors"
	"fmt"
	"os"
)

var (
	ErrInvalidMonartConfig = errors.New("invalid monart configuration")
	ErrInvalidAPIToken     = errors.New("invalid api token")
	ErrInvalidUserID       = errors.New("invalid user id")
)

type Monart struct {
	APIToken string
	ID       string
}

func (m *Monart) Valid() error {
	if m.APIToken == "" {
		return ErrInvalidAPIToken
	}
	if m.ID == "" {
		return ErrInvalidUserID
	}
	return nil
}

func LoadMonartClient() (*Monart, error) {
	mon := &Monart{
		APIToken: os.Getenv("MONART_API_TOKEN"),
		ID:       os.Getenv("MONART_ID"),
	}
	if err := mon.Valid(); err != nil {
		return nil, fmt.Errorf("failed to validate config: %w %w", ErrInvalidMonartConfig, err)
	}
	return mon, nil
}
