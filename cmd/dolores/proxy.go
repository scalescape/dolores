package main

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/scalescape/dolores/monitor"
	"github.com/urfave/cli/v2"
)

type baseCommand struct {
	*cli.Command
	log zerolog.Logger
}

type Monitor struct {
	baseCommand
}

func (m Monitor) Daemon(cctx *cli.Context) error {
	cfg, err := parseServerConfig(cctx)
	if err != nil {
		return err
	}
	backend, err := parseBackendConfig(cctx)
	if err != nil {
		return err
	}
	pr := monitor.NewProxy(cfg, backend)
	if err := pr.Start(); err != nil {
		return err
	}
	return nil
}

func monitorCommand(action cli.ActionFunc) *cli.Command {
	return &cli.Command{
		Name:  "monitor",
		Usage: "daemon to monitor HTTP API",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "backend",
				Usage:    "backend url (http://localhost:8080) for which the requests should be forwarded",
				Required: true,
			},
			&cli.StringFlag{
				Name:  "host",
				Value: "localhost",
			},
			&cli.IntFlag{
				Name:    "port",
				Aliases: []string{"p"},
				Value:   9980,
			},
		},
		Action: action,
	}
}

func NewMonitor() *cli.Command {
	mon := Monitor{
		baseCommand: baseCommand{
			log: log.With().Str("cmd", "config").Logger(),
		},
	}
	return monitorCommand(mon.Daemon)
}
