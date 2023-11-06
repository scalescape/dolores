package client

import (
	"context"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/scalescape/dolores/config"
	"github.com/scalescape/dolores/store/google"
)

type Client struct {
	Service Service
	bucket  string
	prefix  string
	ctx     context.Context //nolint:containedctx
	log     zerolog.Logger
}

type EncryptedConfig struct {
	Environment string `json:"environment"`
	Name        string `json:"name"`
	Data        string `json:"data"`
}

func (c *Client) Init(ctx context.Context, bucket string, cfg Configuration) error {
	return c.Service.Init(ctx, bucket, cfg)
}

func (c *Client) UploadSecrets(req EncryptedConfig) error {
	c.log.Trace().Msgf("uploading to %s name: %s", c.bucket, req.Name)
	return c.Service.Upload(c.ctx, req, c.bucket)
}

type FetchSecretRequest struct {
	Environment string `json:"environment"`
	Name        string `json:"name"`
}
type FetchSecretResponse struct {
	Data string `json:"data"`
}

func (c *Client) FetchSecrets(req FetchSecretRequest) ([]byte, error) {
	data, err := c.Service.FetchConfig(c.ctx, c.bucket, req)
	if err != nil {
		return nil, err
	}
	return data, nil
}

type Recipient struct {
	PublicKey string `json:"public_key"`
}

type OrgPublicKeys struct {
	Recipients []Recipient `json:"recipients"`
}

func (pk OrgPublicKeys) RecipientList() []string {
	result := make([]string, len(pk.Recipients))
	for i, k := range pk.Recipients {
		result[i] = k.PublicKey
	}
	return result
}

func (c *Client) GetOrgPublicKeys(env string) (OrgPublicKeys, error) {
	c.log.Debug().Msgf("fetching public keys for env: %s", env)
	keys, err := c.Service.GetOrgPublicKeys(c.ctx, env, c.bucket, c.prefix+"/keys")
	if err != nil || len(keys) == 0 {
		return OrgPublicKeys{}, err
	}
	recps := make([]Recipient, len(keys))
	for i, k := range keys {
		recps[i].PublicKey = k
	}
	return OrgPublicKeys{Recipients: recps}, nil
}

type GetSecretListConfig struct {
	Environment string `json:"environment"`
}
type GetSecretListResponse struct {
	Secrets []google.SecretObject `json:"secrets"`
}

func (c *Client) GetSecretList(cfg GetSecretListConfig) ([]google.SecretObject, error) {
	return c.Service.GetObjList(c.ctx, c.bucket, c.prefix)
}

func New(ctx context.Context, cfg config.Client) (*Client, error) {
	if err := cfg.Valid(); err != nil {
		return nil, err
	}
	gcfg := google.Config{ServiceAccountFile: cfg.Google.ApplicationCredentials}
	st, err := google.NewStore(ctx, gcfg)
	if err != nil {
		return nil, err
	}
	cli := &Client{
		ctx:     ctx,
		Service: Service{store: st},
		bucket:  cfg.BucketName(),
		prefix:  cfg.StoragePrefix,
		log:     log.With().Str("bucket", cfg.BucketName()).Str("prefix", cfg.StoragePrefix).Logger(),
	}
	return cli, nil
}
