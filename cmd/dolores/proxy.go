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
	m.log.Info().Msgf("I'm here")
	cfg := monitor.Config{Port: cctx.Int("port"), Host: cctx.String("host")}
	pr := monitor.NewProxy(cfg)
	pr.Start()
	return nil
}

func monitorCommand(action cli.ActionFunc) *cli.Command {
	return &cli.Command{
		Name:  "monitor",
		Usage: "daemon to monitor HTTP API",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "host",
				Value: "localhost",
			},
			&cli.IntFlag{
				Name:    "port",
				Aliases: []string{"p"},
				Value:   9980,
			},
			&cli.IntFlag{
				Name:     "watch",
				Aliases:  []string{"w"},
				Required: true,
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
