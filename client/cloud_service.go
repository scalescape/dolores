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
	cloud "github.com/scalescape/dolores/store/cld"
)

var ErrInvalidPublicKeys = errors.New("invalid public keys")

const metadataFile = "dolores.md"

type Service struct {
	store storeI
}

type storeI interface {
	WriteToObject(ctx context.Context, bucketName, fileName string, data []byte) error
	ReadObject(ctx context.Context, bucketName, fileName string) ([]byte, error)
	ListObject(ctx context.Context, bucketName, path string) ([]cloud.Object, error)
	ExistsObject(ctx context.Context, bucketName, fileName string) (bool, error)
}

type Configuration struct {
	Metadata  config.Metadata
	PublicKey string
	UserID    string
}

func (s Service) Init(ctx context.Context, bucket string, cfg Configuration) error {
	// saving metadata and append key to google cloud storage
	if cfg.PublicKey != "" {
		pubKey := fmt.Sprintf("%s/keys/%s.key", cfg.Metadata.Location, cfg.UserID)
		if err := s.uploadPubKey(ctx, bucket, pubKey, cfg.PublicKey); err != nil {
			return fmt.Errorf("error writing public key: %w", err)
		}
	}
	exists, err := s.store.ExistsObject(ctx, bucket, metadataFile)
	if err != nil {
		return err
	}
	if exists {
		log.Info().Msgf("metadata already configured in remote")
		return nil
	}
	if err := s.saveObject(ctx, bucket, metadataFile, cfg.Metadata); err != nil {
		log.Error().Msgf("error writing metadta: %v", err)
		return err
	}
	return nil
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

func (s Service) GetOrgPublicKeys(ctx context.Context, env, bucketName, path string) ([]string, error) {
	pubKey := os.Getenv("DOLORES_PUBLIC_KEY")
	if pubKey != "" {
		return []string{pubKey}, nil
	}
	resp, err := s.ListObject(ctx, bucketName, path)
	if err != nil {
		return nil, fmt.Errorf("error listing objects: %w", err)
	}
	keys := make([]string, len(resp))
	for i, obj := range resp {
		key, err := s.store.ReadObject(ctx, bucketName, obj.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to read object %s: %w", obj, err)
		}
		keys[i] = string(key)
	}
	return keys, nil
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

func (s Service) uploadPubKey(ctx context.Context, bucket string, path, key string) error {
	return s.store.WriteToObject(ctx, bucket, path, []byte(key))
}

func (s Service) readMetadata(ctx context.Context, bucket, mdf string) ([]byte, error) {
	log.Trace().Msgf("reading metadata from bucket:%s file:%s", bucket, mdf)
	return s.store.ReadObject(ctx, bucket, mdf)
}

func (s Service) saveObject(ctx context.Context, bucket, fname string, md any) error {
	data, err := json.Marshal(md)
	if err != nil {
		return err
	}
	return s.store.WriteToObject(ctx, bucket, fname, data)
}

func (s Service) ListObject(ctx context.Context, bucket, path string) ([]cloud.Object, error) {
	resp, err := s.store.ListObject(ctx, bucket, path)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func NewService(st storeI) Service {
	return Service{store: st}
}
