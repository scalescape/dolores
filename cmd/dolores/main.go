package main

import (
	"context"
	"fmt"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/scalescape/dolores/client"
	"github.com/scalescape/dolores/config"
	"github.com/urfave/cli/v2"
)

var (
	version = "0.0.1"
	sha     = "undefined"
)

type CtxKey string

var EnvValue CtxKey = "environment"

func main() {
	log.Logger = log.Output(zerolog.NewConsoleWriter())
	cli.VersionPrinter = VersionDisplay

	app := &cli.App{
		Name:    "Dolores",
		Usage:   "service configuration management with your own cloud platform",
		Version: version,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name: "environment", Aliases: []string{"env"},
				Usage: "environment where you want to manage [staging|production]",
				Action: func(cctx *cli.Context, v string) error {
					cctx.Context = context.WithValue(cctx.Context, EnvValue, v)
					if v == "" {
						return fmt.Errorf("invalid flag: %w", ErrInvalidEnvironment)
					}
					return nil
				},
			},
		},
		Commands: []*cli.Command{
			NewConfig(newClient).Command,
			NewRunner(newClient).Command,
			NewInitCommand(newClient),
		},
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal().Msgf("error: %v", err)
	}
}

func VersionDisplay(cc *cli.Context) {
	fmt.Printf("rom %s (%s)", version, sha) //nolint
}

func newClient(ctx context.Context) *client.Client {
	env, ok := ctx.Value(EnvValue).(string)
	if !ok || env == "" {
		log.Fatal().Msgf("environment not passed properly")
	}
	cfg, err := config.LoadClient(env)
	if err != nil {
		log.Fatal().Msgf("error loading config: %v", err)
		return nil
	}
	client, err := client.New(ctx, cfg)
	if err != nil {
		log.Fatal().Msgf("error building client: %v", err)
		return nil
	}
	return client
}
