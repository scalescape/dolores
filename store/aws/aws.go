package aws

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/rs/zerolog/log"
	cloud "github.com/scalescape/dolores/store/cld"
)

var ErrInvalidServiceAccount = errors.New("invalid service account")

type StorageClient struct {
	client *s3.Client
	region string
}

type Config struct {
	ServiceAccountFile string
}

type ServiceAccount struct {
	AccessKeyID     string `json:"accessKey"`
	SecretAccessKey string `json:"secretKey"`
	Region          string `json:"region"`
}

func (s StorageClient) bucketExists(ctx context.Context, bucketName string) (bool, error) {
	_, err := s.client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		var notFoundType *types.NotFound
		if errors.As(err, &notFoundType) {
			return false, nil
		}
	}
	return true, err
}

func (s StorageClient) CreateBucket(ctx context.Context, bucketName string) error {
	lconst := types.BucketLocationConstraint(s.region)
	cbCfg := &types.CreateBucketConfiguration{LocationConstraint: lconst}
	bucket := &s3.CreateBucketInput{Bucket: aws.String(bucketName),
		CreateBucketConfiguration: cbCfg}
	_, err := s.client.CreateBucket(ctx, bucket)
	var existsErr *types.BucketAlreadyOwnedByYou = new(types.BucketAlreadyOwnedByYou)
	if errors.As(err, &existsErr) {
		log.Debug().Msgf("bucket %s already exists", bucketName)
		return nil
	}
	if err != nil {
		return fmt.Errorf("error creating bucket: %s at region %s: %w", bucketName, s.region, err)
	}
	return nil
}

func (s StorageClient) ListObject(ctx context.Context, bucket, path string) ([]cloud.Object, error) {
	resp, err := s.client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
		Prefix: aws.String(path),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get object list: %w", err)
	}

	items := resp.Contents
	objs := make([]cloud.Object, len(items))
	for i, item := range items {
		o := cloud.Object{Name: *item.Key, Updated: *item.LastModified, Bucket: bucket}
		objs[i] = o
	}
	log.Trace().Msgf("list of objects from path: %s length: %+v", path, len(objs))
	return objs, nil
}

func (s StorageClient) WriteToObject(ctx context.Context, bucketName, fileName string, data []byte) error {
	log.Debug().Msgf("writing to %s/%s", bucketName, fileName)
	bucketExist, err := s.bucketExists(ctx, bucketName)
	if err != nil {
		return fmt.Errorf("failed to fetch bucket: %w", err)
	}
	if !bucketExist {
		if err := s.CreateBucket(ctx, bucketName); err != nil {
			return err
		}
	}

	fileReader := bytes.NewReader(data)
	_, err = s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(fileName),
		Body:   fileReader,
	})

	if err != nil {
		return fmt.Errorf("failed to upload secret: %w", err)
	}
	return nil
}

func (s StorageClient) ReadObject(ctx context.Context, bucketName, fileName string) ([]byte, error) {
	resp, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(fileName),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to read object : %w", err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body : %w", err)
	}
	return data, nil
}

func (s StorageClient) ExistsObject(ctx context.Context, bucketName, fileName string) (bool, error) {
	_, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(fileName),
	})
	if err != nil {
		var notFoundType *types.NoSuchKey
		if errors.As(err, &notFoundType) {
			return false, nil
		} else {
			return false, err
		}
	}

	return true, nil
}

func NewStore(ctx context.Context, acfg Config) (StorageClient, error) {
	data, err := os.ReadFile(acfg.ServiceAccountFile)
	if err != nil {
		return StorageClient{}, fmt.Errorf("failed to read service account file with error %v %w", err, ErrInvalidServiceAccount)
	}
	sa := new(ServiceAccount)
	if err := json.Unmarshal(data, sa); err != nil {
		return StorageClient{}, fmt.Errorf("unable to parse service account file: %w", err)
	}
	cp := credentials.NewStaticCredentialsProvider(sa.AccessKeyID, sa.SecretAccessKey, "")
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(sa.Region), config.WithCredentialsProvider(cp))
	if err != nil {
		return StorageClient{}, err
	}
	cli := s3.NewFromConfig(cfg)
	return StorageClient{client: cli, region: sa.Region}, nil
}
