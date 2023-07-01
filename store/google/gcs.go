package google

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"cloud.google.com/go/storage"
	"github.com/rs/zerolog/log"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

var ErrInvalidServiceAccount = errors.New("invalid service account")

type StorageClient struct {
	*storage.Client
	projectID string
}

type Config struct {
	ServiceAccountFile string
}

type ServiceAccount struct {
	Type                    string `json:"type"`
	ProjectID               string `json:"project_id"`
	PrivateKeyID            string `json:"private_key_id"`
	PrivateKey              string `json:"private_key"`
	ClientEmail             string `json:"client_email"`
	ClientID                string `json:"client_id"`
	AuthURI                 string `json:"auth_uri"`
	TokenURI                string `json:"token_uri"`
	AuthProviderX509CertURL string `json:"auth_provider_x509_cert_url"`
	ClientX509CertURL       string `json:"client_x509_cert_url"`
}

func (s StorageClient) CreateBucket(ctx context.Context, bucketName string) error {
	bucket := s.Client.Bucket(bucketName)
	_, err := bucket.Attrs(ctx)
	if errors.Is(err, storage.ErrBucketNotExist) {
		log.Info().Msgf("creating storage bucket: %s", bucketName)
		if err := s.createNewBucket(ctx, bucketName); err != nil {
			return fmt.Errorf("error creating new bucket: %w", err)
		}
		return nil
	} else if err != nil {
		return fmt.Errorf("error creating bucket: %w", err)
	}
	return nil
}

func (s StorageClient) WriteToObject(ctx context.Context, bucketName, fileName string, data []byte) error {
	log.Debug().Msgf("writing to %s/%s", bucketName, fileName)
	bucket := s.Client.Bucket(bucketName)
	_, err := bucket.Attrs(ctx)
	if errors.Is(err, storage.ErrBucketNotExist) {
		if err := s.createNewBucket(ctx, bucketName); err != nil {
			return fmt.Errorf("error creating new bucket: %w", err)
		}
	}
	if err != nil {
		return fmt.Errorf("failed to check bucket: %w", err)
	}
	w := bucket.Object(fileName).NewWriter(ctx)
	defer w.Close()
	if _, err := w.Write(data); err != nil {
		return fmt.Errorf("error writing data to file: %w", err)
	}
	return nil
}

func (s StorageClient) ReadObject(ctx context.Context, bucketName, fileName string) ([]byte, error) {
	bucket := s.Client.Bucket(bucketName)
	if _, err := bucket.Attrs(ctx); err != nil {
		return nil, fmt.Errorf("failed to get bucket: %w", err)
	}
	obj := bucket.Object(fileName)
	if _, err := obj.Attrs(ctx); err != nil {
		return nil, fmt.Errorf("failed to verify bucket attributes: %w", err)
	}
	rdr, err := obj.NewReader(ctx)
	if err != nil {
		return nil, err
	}
	data, err := io.ReadAll(rdr)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (s StorageClient) ListOjbect(ctx context.Context, bucketName, path string) ([]string, error) {
	bucket := s.Client.Bucket(bucketName)
	if _, err := bucket.Attrs(ctx); err != nil {
		return nil, fmt.Errorf("failed to get bucket: %w", err)
	}
	objs := make([]string, 0)
	iter := bucket.Objects(ctx, &storage.Query{Prefix: path})
	for {
		attrs, err := iter.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate object list: %w", err)
		}
		objs = append(objs, attrs.Name)
	}
	log.Trace().Msgf("list of objects from path: %s %+v", path, objs)
	return objs, nil
}

func (s StorageClient) createNewBucket(ctx context.Context, name string) error {
	bucket := s.Client.Bucket(name)
	attrs := &storage.BucketAttrs{PublicAccessPrevention: storage.PublicAccessPreventionEnforced}
	err := bucket.Create(ctx, s.projectID, attrs)
	if err != nil {
		return err
	}
	return nil
}

func NewStore(ctx context.Context, cfg Config) (StorageClient, error) {
	data, err := os.ReadFile(cfg.ServiceAccountFile)
	if err != nil {
		return StorageClient{}, fmt.Errorf("failed to read service account file with error %v %w", err, ErrInvalidServiceAccount)
	}
	sa := new(ServiceAccount)
	if err := json.Unmarshal(data, sa); err != nil {
		return StorageClient{}, fmt.Errorf("unable to parse service account file: %w", err)
	}
	client, err := storage.NewClient(ctx, option.WithCredentialsJSON(data))
	if err != nil {
		return StorageClient{}, fmt.Errorf("error creating gcp storage client: %w", err)
	}
	return StorageClient{Client: client, projectID: sa.ProjectID}, nil
}
