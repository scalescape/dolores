package cloud

import (
	"context"
	"fmt"

	"github.com/scalescape/dolores/server/cloud/aws"
	"github.com/scalescape/dolores/server/cloud/cld"
)

type StorageClient interface {
	WriteToObject(ctx context.Context, bucket string, file string, data []byte) error
	ReadObject(ctx context.Context, bucketName, fileName string) ([]byte, error)
	ListObject(ctx context.Context, bucketName, fileName string) ([]cld.Object, error)
}

type Option func(*Config)

func NewStorageClient(ctx context.Context, cfg *Config, opts ...Option) (StorageClient, error) {
	for _, opt := range opts {
		opt(cfg)
	}
	acfg, err := cfg.AWSConfig()
	if err != nil {
		return nil, fmt.Errorf("error build aws config: %w", err)
	}
	return aws.NewStorageClient(ctx, acfg)
}
