package aws

import (
	"errors"
	"fmt"
)

var ErrInvalidAWSCredentials = errors.New("invalid AWS credentials")

type Config struct {
	Credentials
	Token  string
	Region string
}

type Credentials struct {
	AccessKeyID     string `json:"access_key_id"`
	SecretAccessKey string `json:"secret_access_key"`
}

func (c Credentials) Valid() error {
	if c.AccessKeyID == "" {
		return fmt.Errorf("invalid access key id: %w", ErrInvalidAWSCredentials)
	}
	if c.SecretAccessKey == "" {
		return fmt.Errorf("invalid access secret key: %w", ErrInvalidAWSCredentials)
	}
	return nil
}
