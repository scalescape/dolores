package client

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/rs/zerolog/log"
	"github.com/scalescape/dolores/config"
	"github.com/scalescape/dolores/store/google"
)

var ErrInvalidPublicKeys = errors.New("invalid public keys")

const metadataFile = "dolores.md"

type Service struct {
	store google.StorageClient
}

func (s Service) Upload(ctx context.Context, req EncryptedConfig, bucket string) error {
	prefix, err := s.getObjectPrefix(ctx, req.Environment, bucket)
	if err != nil {
		return err
	}
	fileName := req.Name
	if prefix != "" {
		fileName = fmt.Sprintf("%s/%s", prefix, fileName)
	}
	data, err := base64.StdEncoding.DecodeString(req.Data)
	if err != nil {
		return err
	}
	return s.store.WriteToObject(ctx, bucket, fileName, data)
}

func (s Service) GetOrgPublicKeys(ctx context.Context, env, bucketName string) ([]string, error) {
	pubKey := os.Getenv("DOLORES_PUBLIC_KEY")
	if pubKey != "" {
		return []string{pubKey}, nil
	}
	return nil, ErrInvalidPublicKeys
}

func (s Service) getObjectPrefix(ctx context.Context, env, bucket string) (string, error) {
	md, err := s.readMetadata(ctx, bucket, metadataFile)
	if err != nil {
		return "", fmt.Errorf("failed to read metadata: %w", err)
	}
	var meta config.Metadata
	if err := json.Unmarshal(md, &meta); err != nil {
		return "", fmt.Errorf("failed to parse metadata file: %w", err)
	}
	return meta.Location, nil
}

func (s Service) FetchConfig(ctx context.Context, bucket string, req FetchSecretRequest) ([]byte, error) {
	fileName := req.Name
	prefix, err := s.getObjectPrefix(ctx, req.Environment, bucket)
	if err != nil {
		return nil, err
	}
	if prefix != "" {
		fileName = fmt.Sprintf("%s/%s", prefix, fileName)
	}
	data, err := s.store.ReadObject(ctx, bucket, fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to read config %s with error: %w", fileName, err)
	}
	return data, nil
}

func (s Service) readMetadata(ctx context.Context, bucket, mdf string) ([]byte, error) {
	log.Trace().Msgf("reading metadata from bucket:%s file:%s", bucket, mdf)
	return s.store.ReadObject(ctx, bucket, mdf)
}

func (s Service) SaveObject(ctx context.Context, bucket, fname string, md any) error {
	data, err := json.Marshal(md)
	if err != nil {
		return err
	}
	return s.store.WriteToObject(ctx, bucket, fname, data)
}
