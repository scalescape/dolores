package dolores

import (
	"bytes"
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog/log"
)

type Variable struct {
	Key   []byte
	Value []byte
}

func (v Variable) Data() []byte {
	return append(v.Key, v.Value...)
}

type EnvFile struct {
	data      []byte
	Variables []Variable
	CreatedAt time.Time
}

func (ef *EnvFile) Parse() error {
	for i, line := range bytes.Split(ef.data, []byte("\n")) {
		line = bytes.TrimSpace(line)
		if line == nil {
			continue
		}
		if line[0] == '#' {
			log.Debug().Msgf("parsing comment: %s", line)
			continue
		}
		split := bytes.Split(line, []byte("="))
		if len(split) != 2 {
			return fmt.Errorf("error parsing line: %d %w", i, ErrInvalidFormat)
		}
		ef.Variables = append(ef.Variables, Variable{Key: split[0], Value: split[1]})
	}
	return nil
}

func LoadEnvFile(fn string) (*EnvFile, error) {
	data, err := os.ReadFile(fn)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %s %w", fn, err)
	}
	envFile := &EnvFile{
		data:      data,
		CreatedAt: time.Now().UTC(),
	}
	if err := envFile.Parse(); err != nil {
		return nil, err
	}
	return envFile, nil
}
