package main

import (
	"context"
	"fmt"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/scalescape/dolores/client"
	"github.com/scalescape/dolores/secrets"
	"github.com/urfave/cli/v2"
)

type ConfigCommand struct {
	*cli.Command
	rcli func(context.Context) *client.Client
	log  zerolog.Logger
}

func NewConfig(client func(context.Context) *client.Client) *ConfigCommand {
	log := log.With().Str("cmd", "config").Logger()
	cmd := &cli.Command{
		Name:  "config",
		Usage: "secrets management",
		Flags: []cli.Flag{},
	}
	cfg := &ConfigCommand{
		Command: cmd,
		log:     log,
		rcli:    client,
	}
	cfg.Subcommands = append(cfg.Subcommands, EncryptCommand(cfg.encryptAction))
	cfg.Subcommands = append(cfg.Subcommands, DecryptCommand(cfg.decryptAction))
	cfg.Subcommands = append(cfg.Subcommands, EditCommand(cfg.editAction))
	return cfg
}

func (c *ConfigCommand) editAction(ctx *cli.Context) error {
	env := ctx.String("environment")
	log := c.log.With().Str("cmd", "config.edit").Str("environment", env).Logger()
	dcfg, err := parseDecryptConfig(ctx)
	if err != nil {
		return err
	}
	sec := secrets.NewSecertsManager(log, c.rcli(ctx.Context))
	cfg := secrets.EditConfig{DecryptConfig: dcfg}
	if err := sec.Edit(cfg); err != nil {
		log.Error().Msgf("error editing file: %v", err)
		return err
	}
	return nil
}

func (c *ConfigCommand) encryptAction(ctx *cli.Context) error {
	env := ctx.String("environment")
	file := ctx.String("file")
	name := ctx.String("name")
	log := c.log.With().Str("cmd", "config.encrypt").Str("environment", env).Logger()
	secMan := secrets.NewSecertsManager(log, c.rcli(ctx.Context))
	req := secrets.EncryptConfig{Environment: env, FileName: file, Name: name}
	if err := secMan.Encrypt(req); err != nil {
		return err
	}
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
	parseKeyConfig(ctx, &req)
	if err := req.Valid(); err != nil {
		return secrets.DecryptConfig{}, fmt.Errorf("pass appropriate key or key-file to decrypt: %w", err)
	}

	return req, nil
}

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

func (c *ConfigCommand) decryptAction(ctx *cli.Context) error {
	req, err := parseDecryptConfig(ctx)
	if err != nil {
		return err
	}
	log := c.log.With().Str("cmd", "config.dencrypt").Str("environment", req.Environment).Logger()
	log.Trace().Str("cmd", "config.decrypt").Msgf("running decryption")
	sec := secrets.NewSecertsManager(log, c.rcli(ctx.Context))
	return sec.Decrypt(req)
}

func EncryptCommand(action cli.ActionFunc) *cli.Command {
	return &cli.Command{
		Name:  "encrypt",
		Usage: "encrypt the passed file",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "file",
				Aliases:  []string{"f"},
				Required: true,
			},
			&cli.StringFlag{
				Name:     "name",
				Required: true,
			},
		},
		Action: action,
	}
}

func DecryptCommand(action cli.ActionFunc) *cli.Command {
	return &cli.Command{
		Name:  "decrypt",
		Usage: "decrypt the remote configuration",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "name",
				Required: true,
			},
			&cli.StringFlag{
				Name:    "key",
				Aliases: []string{"k"},
			},
			&cli.StringFlag{
				Name: "key-file",
			},
		},
		Action: action,
	}
}

func EditCommand(action cli.ActionFunc) *cli.Command {
	return &cli.Command{
		Name:  "edit",
		Usage: "edit the secrets configuration",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "name",
				Required: true,
			},
			&cli.StringFlag{
				Name:    "key",
				Aliases: []string{"k"},
			},
			&cli.StringFlag{
				Name: "key-file",
			},
		},
		Action: action,
	}
}
