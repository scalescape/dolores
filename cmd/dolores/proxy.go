package main

import (
	"fmt"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/scalescape/dolores/monitor"
	"github.com/scalescape/go-metrics"
	"github.com/scalescape/go-metrics/common"
	"github.com/urfave/cli/v2"
)

type baseCommand struct {
	*cli.Command
	log zerolog.Logger
}

const DefaultPort = 9980

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
	metricsStoreAddress := cctx.String("statsd")
	obs, err := metrics.Setup(
		metrics.WithAddress(metricsStoreAddress),
		metrics.WithKind(common.Statsd),
	)
	if err != nil {
		return fmt.Errorf("failed to create observer: %w", err)
	}
	pr, err := monitor.NewProxy(cfg, backend, obs)
	if err != nil {
		return fmt.Errorf("failed to instantiate proxy: %w", err)
	}
	return pr.Start()
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
				Value:   DefaultPort,
			},
			&cli.StringFlag{
				Name:     "statsd",
				Usage:    "statsd address host:8125",
				Value:    "localhost:8125",
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
