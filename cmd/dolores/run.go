package main

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/scalescape/dolores/lib"
	"github.com/scalescape/dolores/secrets"
	"github.com/urfave/cli/v2"
)

type OutputType string

var ErrInvalidCommand = errors.New("invalid command")

const (
	Stdout OutputType = "stdout"
	Stderr OutputType = "stderr"
)

func handleReader(wg *sync.WaitGroup, reader io.Reader, mode OutputType) {
	defer wg.Done()
	buf := make([]byte, 1024)
	for {
		n, err := reader.Read(buf)
		if err != nil && err != io.EOF {
			log.Error().Msgf("Error reading from reader: %v", err)
			return
		}
		if n == 0 {
			break
		}
		if mode == Stdout {
			os.Stdout.Write(buf[:n])
		} else if mode == Stderr {
			log.Error().Msgf("%s", string(buf[:n]))
		}
	}
}

func (c *Runner) environ(ctx context.Context, name string) ([]string, error) {
	envs := os.Environ()
	log := log.With().Str("cmd", "run").Str("environment", c.environment).Logger()
	log.Debug().Msgf("loading configuration %s before running", name)
	sec := secrets.NewSecretsManager(log, c.rcli(ctx))
	if err := sec.Decrypt(c.DecryptConfig); err != nil {
		return nil, err
	}
	scanner := bufio.NewScanner(c.configBuffer)
	for scanner.Scan() {
		v := scanner.Text()
		envs = append(envs, v)
	}
	return envs, nil
}

func (c *Runner) runScript(ctx context.Context, cmdName string, args []string) error {
	log.Trace().Msgf("executing cmd: %s with args: %s", cmdName, args)
	cmd := exec.CommandContext(ctx, cmdName, args...)
	if c.configName != "" {
		vars, err := c.environ(ctx, c.configName)
		if err != nil {
			return err
		}
		cmd.Env = vars
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("error creating stdout pipe: %w", err)
	}
	defer stdout.Close()
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("error creating stderr pipe: %w", err)
	}
	defer stderr.Close()
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("error starting command: %w", err)
	}
	c.wg.Add(2)
	go handleReader(c.wg, stdout, Stdout)
	go handleReader(c.wg, stderr, Stderr)
	c.wg.Wait()

	if err := cmd.Wait(); err != nil {
		if err, ok := err.(*exec.ExitError); ok { // nolint:errorlint
			if status, ok := err.Sys().(syscall.WaitStatus); ok {
				// TODO: Report this error exit status
				c.exitStatus = status.ExitStatus()
			}
		}
		return fmt.Errorf("exited with error: %w", err)
	}
	return nil
}

type Runner struct {
	*cli.Command
	rcli GetClient
	execCommand
	exitStatus   int
	wg           *sync.WaitGroup
	configName   string
	configBuffer *bytes.Buffer
	environment  string
	secrets.DecryptConfig
}

type execCommand struct {
	Script  string
	Command string
	Args    []string
}

func (c *execCommand) Valid() error {
	if c.Script == "" && c.Command == "" {
		return fmt.Errorf("please pass script flag or a valid command to run: %w", ErrInvalidCommand)
	}
	if c.Script != "" {
		if _, err := os.Stat(c.Script); err != nil {
			return fmt.Errorf("invalid script file: %s %w", c.Script, err)
		}
	}
	return nil
}

func (c *Runner) parse(ctx *cli.Context) error {
	req := execCommand{}
	if script := ctx.String("script"); script == "" {
		req.Command = ctx.Args().First()
		req.Args = ctx.Args().Tail()
	} else {
		req.Args = append(req.Args, lib.AbsPath(script))
		req.Args = append(req.Args, ctx.Args().Slice()...)
		req.Command = "/bin/bash"
	}
	c.configName = ctx.String("with-config")
	c.environment = ctx.String("environment")
	if c.configName != "" && c.environment == "" {
		return fmt.Errorf("pass environment: %w", ErrInvalidEnvironment)
	}
	if c.configName != "" && c.environment != "" {
		c.DecryptConfig = secrets.DecryptConfig{
			Name:        c.configName,
			Environment: c.environment,
			Out:         c.configBuffer,
		}
		if err := c.DecryptConfig.Valid(); err != nil {
			return err
		}
	}
	if err := req.Valid(); err != nil {
		return err
	}
	c.execCommand = req
	return nil
}

func (c *Runner) runAction(ctx *cli.Context) error {
	startT := time.Now()
	defer func() {
		log.Info().Msgf("total elapsed time: %s", time.Since(startT))
	}()
	if err := c.parse(ctx); err != nil {
		return err
	}
	if err := c.runScript(ctx.Context, c.execCommand.Command, c.execCommand.Args); err != nil {
		return err
	}
	return nil
}

func NewRunner(client GetClient) Runner {
	cmd := Runner{
		rcli:         client,
		wg:           new(sync.WaitGroup),
		configBuffer: bytes.NewBuffer(make([]byte, 0)),
		Command: &cli.Command{
			Name:  "run",
			Usage: "execute a script or command with secrets loaded",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name: "script",
				},
				&cli.StringFlag{
					Name:  "with-config",
					Usage: "load the secrets config before running",
				},
				&cli.StringFlag{
					Name: "key-file",
				},
				&cli.StringFlag{
					Name: "key",
				},
			},
		},
	}
	cmd.Action = cmd.runAction
	return cmd
}
