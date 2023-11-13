package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
)

var ErrInvalidDoloresConfig = errors.New("invalid dolores config")

type Environment struct {
	Metadata      `json:"metadata"`
	KeyFile       string `json:"key_file"`
	CloudProvider string `json:"cloud_provider"`
}

type Dolores struct {
	Environments map[string]Environment `json:"environments"`
}

func (d *Dolores) AddEnvironment(env string, keyFile string, md Metadata) {
	if d.Environments == nil {
		d.Environments = make(map[string]Environment)
	}
	d.Environments[env] = Environment{
		Metadata:      md,
		KeyFile:       keyFile,
		CloudProvider: AWS, // Temp adding for testing
	}
}

func (d *Dolores) valid() error {
	if len(d.Environments) == 0 {
		log.Trace().Msgf("no environments configured")
		return ErrInvalidDoloresConfig
	}
	return nil
}

func (d Dolores) SaveToDisk() error {
	err := os.MkdirAll(filepath.Dir(File), os.ModePerm)
	if err != nil {
		return err
	}
	f, err := os.Create(File)
	if err != nil {
		return err
	}
	defer f.Close()
	err = json.NewEncoder(f).Encode(d)
	if err != nil {
		return err
	}
	log.Debug().Msgf("created default configuration %s", File)
	return nil
}

func LoadFromDisk() (*Dolores, error) {
	data, err := os.ReadFile(File)
	if err != nil {
		return nil, err
	}
	d := new(Dolores)
	if err := json.Unmarshal(data, d); err != nil {
		return nil, err
	}
	return d, d.valid()
}
