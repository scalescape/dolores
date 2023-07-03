package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"filippo.io/age"
	"github.com/AlecAivazis/survey/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/scalescape/dolores/client"
	"github.com/scalescape/dolores/config"
	"github.com/urfave/cli/v2"
)

var ErrInvalidEnvironment = errors.New("invalid environment")

type InitCommand struct {
	*cli.Command
	log  zerolog.Logger
	rcli func(context.Context) *client.Client
}

type GetClient func(context.Context) *client.Client

func NewInitCommand(newCli GetClient) *cli.Command {
	ic := &InitCommand{
		log:  log.With().Str("cmd", "init").Logger(),
		rcli: newCli,
	}
	cmd := &cli.Command{
		Name:   "init",
		Usage:  "bootstrap with settings",
		Action: ic.initialize,
	}
	return cmd
}

type Input struct {
	UserID                 string `survey:"user_id"`
	Bucket                 string
	Location               string
	ApplicationCredentials string `survey:"google_creds"`
}

func (inp Input) ToMetadata(env string) config.Metadata {
	return config.Metadata{
		Bucket:      inp.Bucket,
		Location:    inp.Location,
		CreatedAt:   time.Now(),
		Environment: env,
	}
}

func (c *InitCommand) getData(env string) (*Input, error) {
	credFile := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	qs := []*survey.Question{
		{
			Name: "bucket",
			Prompt: &survey.Input{
				Message: "Enter the GCS bucket name:",
			},
			Validate: survey.Required,
		},
		{
			Name:     "location",
			Validate: survey.Required,
			Prompt: &survey.Input{
				Message: "Enter the object location to store the secrets:",
			},
		},
		{
			Name:     "user_id",
			Validate: survey.Required,
			Prompt: &survey.Input{
				Message: "Enter your unique name/id",
				Default: os.Getenv("USER"),
			},
		},
	}
	if credFile != "" {
		qs = append(qs, &survey.Question{
			Name:     "google_creds",
			Validate: survey.Required,
			Prompt: &survey.Select{
				Message: "Use GOOGLE_APPLICATION_CREDENTIALS env as credentials file",
				Options: []string{credFile},
			},
		})
	} else {
		qs = append(qs, &survey.Question{
			Name: "google_creds",
			Prompt: &survey.Input{
				Message: "Enter google service account file path",
			},
			Validate: survey.Required,
		})
	}
	res := new(Input)
	if err := survey.Ask(qs, res); err != nil {
		return nil, fmt.Errorf("failed to get appropriate input: %w", err)
	}
	return res, nil
}

func (c *InitCommand) createConfig(configDir, keyFile string) error {
	if err := os.MkdirAll(configDir, 0o770); err != nil {
		return fmt.Errorf("failed to create dir: %w", err)
	}
	return nil
}

func (c *InitCommand) generateKey(fname string) (string, error) {
	f, err := os.OpenFile(fname, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return "", fmt.Errorf("failed to create file %s with error %w", fname, err)
	}
	defer f.Close()
	k, err := age.GenerateX25519Identity()
	if err != nil {
		return "", fmt.Errorf("failed to generate key: %w", err)
	}
	pubKey := k.Recipient().String()
	fmt.Fprintf(f, "# created: %s\n", time.Now().Format(time.RFC3339))
	fmt.Fprintf(f, "# public key: %s\n", pubKey)
	fmt.Fprintf(f, "%s\n", k)
	log.Info().Msgf("successfully generated asymmetric key")
	return pubKey, nil
}

func (c *InitCommand) initialize(ctx *cli.Context) error {
	env := ctx.String("environment")
	if env == "" {
		return fmt.Errorf("invalid environment: %w", ErrInvalidEnvironment)
	}
	inp, err := c.getData(env)
	if err != nil {
		return err
	}
	keyFilePath := filepath.Join(config.Dir, env+".key")
	if err := c.createConfig(config.Dir, keyFilePath); err != nil {
		return err
	}
	_, err = os.Stat(keyFilePath)
	var publicKey string
	if os.IsNotExist(err) {
		publicKey, err = c.generateKey(keyFilePath)
		if err != nil {
			return err
		}
	} else {
		log.Info().Msgf("asymmetric key already exists at %s", keyFilePath)
	}
	d := &config.Dolores{}
	md := inp.ToMetadata(env)
	d.AddEnvironment(env, keyFilePath, md)
	if err := d.SaveToDisk(); err != nil {
		return fmt.Errorf("error saving dolores config: %w", err)
	}
	cli := c.rcli(ctx.Context)
	cfg := client.Configuration{PublicKey: publicKey, Metadata: md, UserID: inp.UserID}
	if err := cli.Init(ctx.Context, md.Bucket, cfg); err != nil {
		return err
	}
	log.Info().Msgf("successfully initialized dolores")
	return nil
}
