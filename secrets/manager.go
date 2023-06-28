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
)

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
	*client.Client
	log zerolog.Logger
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
	resp, err := sm.Client.GetOrgPublicKeys(env)
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
	if err := sm.Client.UploadSecrets(ureq); err != nil {
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
	data, err := sm.FetchSecrets(req)
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

func NewSecertsManager(log zerolog.Logger, rcli *client.Client) SecretManager {
	return SecretManager{Client: rcli, log: log}
}
