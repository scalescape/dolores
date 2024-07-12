package cloud

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/scalescape/dolores/server/cloud/aws"
)

type Platform string

var (
	GCP Platform = "GCP"
	AWS Platform = "AWS"
)

var ErrInvalidAWSCredentials = errors.New("invalid AWS credentials")

type Config struct {
	OrgID             string
	CreateManagedZone bool
	ProjectID         string
	Credentials       []byte
	Platform
	Region string
}

func (c *Config) AWSConfig() (aws.Config, error) {
	creds, err := c.AwsCredentials()
	if err != nil {
		return aws.Config{}, fmt.Errorf("error parsing aws credentials: %w", err)
	}
	return aws.Config{Credentials: *creds, Region: c.Region}, nil
}

func (c *Config) AwsCredentials() (*aws.Credentials, error) {
	creds := new(aws.Credentials)
	if err := json.Unmarshal(c.Credentials, creds); err != nil {
		return nil, err
	}
	return creds, creds.Valid()
}
