package main

import (
	"context"
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/scalescape/dolores/client"
	"github.com/scalescape/dolores/config"
	"github.com/urfave/cli/v2"
)

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
	Bucket                 string
	Environment            string
	Location               string
	ApplicationCredentials string `survey:"google_creds"`
}

func (c *InitCommand) getData() (*Input, error) {
	credFile := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	qs := []*survey.Question{
		{
			Name: "environment",
			Prompt: &survey.Select{
				Message: "Choose an environment:",
				Options: []string{"production", "staging"},
				Default: "production",
			},
		},
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
				Message: "Enter the object location to store the secrets",
			},
		},
	}
	if credFile != "" {
		qs = append(qs, &survey.Question{
			Name:     "google_creds",
			Validate: survey.Required,
			Prompt: &survey.Select{
				Message: "Confirm using GOOGLE_APPLICATION_CREDENTIALS env as credentials file",
				Options: []string{credFile, ""},
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
		return nil, err
	}
	return res, nil
}

func (c *InitCommand) initialize(ctx *cli.Context) error {
	inp, err := c.getData()
	if err != nil {
		return err
	}
	md := config.Metadata{
		Bucket: inp.Bucket,
		Locations: map[string]string{
			inp.Environment: inp.Location,
		},
	}
	if err := c.rcli(ctx.Context).SaveMetadata(ctx.Context, inp.Bucket, md); err != nil {
		c.log.Error().Msgf("error writing metadta: %v", err)
		return err
	}
	return nil
}
