package secrets

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/scalescape/dolores"
	"github.com/scalescape/dolores/client"
	"github.com/scalescape/dolores/store/google"
)

type secClient interface {
	FetchSecrets(req client.FetchSecretRequest) ([]byte, error)
	UploadSecrets(req client.EncryptedConfig) error
	GetOrgPublicKeys(env string) (client.OrgPublicKeys, error)
	GetSecretList() ([]google.SecretObject, error)
}

type EncryptConfig struct {
	Environment string
	FileName    string
	Name        string
}

var (
	ErrInvalidDecryptConfig = errors.New("invalid decrypt configuration")
	ErrInvalidKeyFile       = errors.New("invalid key file")
	ErrInvalidConfigName    = errors.New("invalid config name")
	ErrInvalidEnvironment   = errors.New("invalid environment")
	ErrInvalidOutput        = errors.New("invalid output writer")
)

type SecretManager struct {
	client secClient
	log    zerolog.Logger
}

func (sm SecretManager) Encrypt(req EncryptConfig) error {
	env, file, name := req.Environment, req.FileName, req.Name
	log := sm.log.With().Str("cmd", "config.encrypt").Str("environment",
		env).Logger()
	log.Debug().Msgf("running encryption")
	envFile, err := dolores.LoadEnvFile(file)
	if err != nil {
		return fmt.Errorf("failed to load file: %w", err)
	}
	resp, err := sm.client.GetOrgPublicKeys(env)
	if err != nil {
		return fmt.Errorf("failed to get keys: %w", err)
	}
	recps := make([]string, len(resp.Recipients))
	for i, r := range resp.Recipients {
		recps[i] = r.PublicKey
	}
	enc, err := dolores.NewEncryptor(recps...)
	if err != nil {
		return fmt.Errorf("error creating encryptor: %w", err)
	}
	data, err := enc.Encrypt(envFile.Variables)
	if err != nil {
		return fmt.Errorf("error encrypting: %w", err)
	}
	log.Debug().Msgf("uploading encrypted file to server: %s", name)
	ureq := client.EncryptedConfig{
		Environment: env,
		Name:        name,
		Data:        base64.StdEncoding.EncodeToString(data),
	}
	if err := sm.client.UploadSecrets(ureq); err != nil {
		return err
	}
	return nil
}

type DecryptConfig struct {
	Name        string
	Environment string
	KeyFile     string
	Key         string
	Out         io.Writer
}

func (c DecryptConfig) Output() io.Writer {
	if c.Out == nil {
		return os.Stdout
	}
	return c.Out
}

func (c DecryptConfig) Valid() error {
	if c.KeyFile == "" && c.Key == "" {
		return ErrInvalidKeyFile
	}
	if c.Name == "" {
		return ErrInvalidConfigName
	}
	if strings.ToLower(c.Environment) != "production" && strings.ToLower(c.Environment) != "staging" {
		return ErrInvalidEnvironment
	}
	if c.Out == nil {
		return ErrInvalidOutput
	}
	return nil
}

func (sm SecretManager) Decrypt(cfg DecryptConfig) error {
	if err := cfg.Valid(); err != nil {
		return fmt.Errorf("invalid config: %w: %v", ErrInvalidDecryptConfig, err)
	}
	req := client.FetchSecretRequest{Name: cfg.Name, Environment: cfg.Environment}
	data, err := sm.client.FetchSecrets(req)
	if err != nil {
		return err
	}
	dc := &dolores.DecryptConfig{KeyFile: cfg.KeyFile, Key: cfg.Key}
	dec, err := dolores.NewDecryptor(dc)
	if err != nil {
		return err
	}
	result, err := dec.Decrypt(data)
	if err != nil {
		return err
	}
	if _, err := cfg.Output().Write(result); err != nil {
		return err
	}
	return nil
}

type ListSecretConfig struct {
	Environment string
	Out         io.Writer
}

func (c ListSecretConfig) Output() io.Writer {
	if c.Out == nil {
		return os.Stdout
	}
	return c.Out
}

func (c ListSecretConfig) Valid() error {
	if strings.ToLower(c.Environment) != "production" && strings.ToLower(c.Environment) != "staging" {
		return ErrInvalidEnvironment
	}
	return nil
}

func (sm SecretManager) ListSecret(cfg ListSecretConfig) error {
	if err := cfg.Valid(); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}
	resp, err := sm.client.GetSecretList()
	if err != nil {
		return fmt.Errorf("failed to get secrets: %w", err)
	}
	if _, err := cfg.Output().Write([]byte(fmt.Sprintf("%-20s %-40s %-40s %-40s\n", "Name", "Bucket", "Create At", "Updated At"))); err != nil {
		return err
	}
	for _, obj := range resp {
		if !strings.HasSuffix(obj.Name, ".key") && !strings.HasSuffix(obj.Name, "/") {
			if _, err := cfg.Output().Write([]byte(fmt.Sprintf("%-20s %-40s %-40s %-40s\n", obj.Name, obj.Bucket, obj.CreatedAt, obj.UpdatedAt))); err != nil {
				return err
			}
		}
	}
	return nil
}

func NewSecretsManager(log zerolog.Logger, rcli secClient) SecretManager {
	return SecretManager{client: rcli, log: log}
}
