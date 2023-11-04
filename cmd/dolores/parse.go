package main

import (
	"fmt"
	"net/url"
	"os"

	"github.com/rs/zerolog/log"
	"github.com/scalescape/dolores"
	"github.com/scalescape/dolores/config"
	"github.com/scalescape/dolores/secrets"
	"github.com/urfave/cli/v2"
)

func parseKeyConfig(ctx *cli.Context, cfg *secrets.DecryptConfig) error {
	log.Trace().Msgf("parsing configuration required to decrypt config")
	key := ctx.String("key")
	keyFile := ctx.String("key-file")
	if keyFile == "" {
		d, err := config.LoadFromDisk()
		if err != nil {
			return fmt.Errorf("dolores not initialized yet: %w", err)
		}
		keyFile = d.Environments[cfg.Environment].KeyFile
	}
	if keyFile == "" {
		keyFile = os.Getenv("DOLORES_SECRETS_KEY_FILE")
	}
	if key == "" {
		key = os.Getenv("DOLORES_SECRETS_KEY")
	}
	cfg.KeyFile = keyFile
	cfg.Key = key

	return nil
}

func parseDecryptConfig(ctx *cli.Context) (secrets.DecryptConfig, error) {
	env := ctx.String("environment")
	name := ctx.String("name")
	req := secrets.DecryptConfig{
		Environment: env,
		Name:        name,
		Out:         os.Stdout,
	}
	if err := parseKeyConfig(ctx, &req); err != nil {
		return secrets.DecryptConfig{}, fmt.Errorf("unable to load key-file from config: %w", err)
	}
	if err := req.Valid(); err != nil {
		return secrets.DecryptConfig{}, fmt.Errorf("pass appropriate key or key-file to decrypt: %w", err)
	}

	return req, nil
}

func parseServerConfig(cctx *cli.Context) (config.Server, error) {
	cfg := config.Server{
		Port: cctx.Int("port"), Host: cctx.String("host"),
	}
	if cfg.Port == 0 {
		return config.Server{}, fmt.Errorf("%w: port", dolores.ErrInvalidConfiguraion)
	}
	return cfg, nil
}

func parseBackendConfig(cctx *cli.Context) (config.Backend, error) {
	burl, err := url.ParseRequestURI(cctx.String("backend"))
	if err != nil {
		return config.Backend{}, err
	}
	backend := config.Backend{
		URL: burl,
	}
	return backend, nil
}
