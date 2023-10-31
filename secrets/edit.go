package secrets

import (
	"bytes"
	"fmt"
	"os"

	"github.com/scalescape/dolores"
	"github.com/scalescape/dolores/client"
	"github.com/scalescape/dolores/lib"
)

type EditConfig struct {
	DecryptConfig
}

func (sm SecretManager) Edit(cfg EditConfig) error {
	req := client.FetchSecretRequest{Name: cfg.Name, Environment: cfg.Environment}
	data, err := sm.client.FetchSecrets(req)
	if err != nil {
		return err
	}
	dc := &dolores.DecryptConfig{KeyFile: cfg.KeyFile, Key: cfg.Key}
	dec, err := dolores.NewDecryptor(dc)
	if err != nil {
		return err
	}
	result, err := dec.Decrypt(data)
	if err != nil {
		return err
	}
	f, err := lib.CreateTempFile(cfg.Name)
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}
	defer f.Close()
	if _, err := f.Write(result); err != nil {
		return err
	}
	if err := sm.editFile(f.Name(), cfg); err != nil {
		return err
	}
	if err := os.Remove(f.Name()); err != nil {
		return err
	}
	return nil
}

func (sm SecretManager) editFile(fname string, cfg EditConfig) error {
	before, err := lib.Hash(fname)
	if err != nil {
		return err
	}
	sm.log.Trace().Msgf("editing config with temp file: %s", fname)
	if err := lib.OpenEditor(fname); err != nil {
		return err
	}
	after, err := lib.Hash(fname)
	if err != nil {
		return err
	}
	if !bytes.Equal(before, after) {
		sm.log.Debug().Msgf("Updating changes to remote")
		ereq := EncryptConfig{Environment: cfg.Environment, FileName: fname, Name: cfg.Name}
		if err := sm.Encrypt(ereq); err != nil {
			return fmt.Errorf("error uploading changes to remote: %w", err)
		}
	}
	return nil
}
