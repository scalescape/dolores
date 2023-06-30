package main

import (
	"fmt"
	"os"

	"github.com/rs/zerolog/log"
	"github.com/scalescape/dolores/secrets"
	"github.com/urfave/cli/v2"
)

func parseKeyConfig(ctx *cli.Context, cfg *secrets.DecryptConfig) {
	log.Trace().Msgf("parsing configuration required to decrypt config")
	key := ctx.String("key")
	keyFile := ctx.String("key-file")
	if keyFile == "" {
		keyFile = os.Getenv("DOLORES_SECRETS_KEY_FILE")
	}
	if key == "" {
		key = os.Getenv("DOLORES_SECRETS_KEY")
	}
	cfg.KeyFile = keyFile
	cfg.Key = key
}

func parseDecryptConfig(ctx *cli.Context) (secrets.DecryptConfig, error) {
	env := ctx.String("environment")
	name := ctx.String("name")
	req := secrets.DecryptConfig{
		Environment: env,
		Name:        name,
		Out:         os.Stdout,
	}
	parseKeyConfig(ctx, &req)
	if err := req.Valid(); err != nil {
		return secrets.DecryptConfig{}, fmt.Errorf("pass appropriate key or key-file to decrypt: %w", err)
	}

	return req, nil
}
