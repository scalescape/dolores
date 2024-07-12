package aws

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/rs/zerolog/log"
	"github.com/scalescape/dolores/server/cloud/cld"
)

type StorageClient struct {
	client *s3.Client
	region string
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
	existsErr := new(types.BucketAlreadyOwnedByYou)
	if errors.As(err, &existsErr) {
		log.Debug().Msgf("bucket %s already exists", bucketName)
		return nil
	}
	if err != nil {
		return fmt.Errorf("error creating bucket: %s at region %s: %w", bucketName, s.region, err)
	}
	return nil
}

func (s StorageClient) ListObject(ctx context.Context, bucket, path string) ([]cld.Object, error) {
	resp, err := s.client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
		Prefix: aws.String(path),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get object list: %w", err)
	}

	items := resp.Contents
	objs := make([]cld.Object, len(items))
	for i, item := range items {
		o := cld.Object{Name: *item.Key, UpdatedAt: *item.LastModified, Bucket: bucket}
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

func NewStorageClient(ctx context.Context, acfg Config) (StorageClient, error) {
	cp := credentials.NewStaticCredentialsProvider(acfg.AccessKeyID, acfg.SecretAccessKey, acfg.Token)
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(acfg.Region), config.WithCredentialsProvider(cp))
	if err != nil {
		return StorageClient{}, err
	}
	cli := s3.NewFromConfig(cfg)
	return StorageClient{client: cli, region: acfg.Region}, nil
}
