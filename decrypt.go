package dolores

import (
	"bytes"
	"fmt"
	"io"

	"filippo.io/age"
	"filippo.io/age/armor"
)

type DecryptConfig struct {
	Key     string
	KeyFile string
}

func (c *DecryptConfig) Identities() ([]age.Identity, error) {
	if c.Key != "" {
		id, err := age.ParseX25519Identity(c.Key)
		if err != nil {
			return nil, fmt.Errorf("failed to get identity: %w", err)
		}
		return []age.Identity{id}, nil
	}
	if c.KeyFile == "" {
		return nil, ErrInvalidKeyFile
	}
	return ParseIdentities(c.KeyFile)
}

func (c *DecryptConfig) Valid() error {
	if c.KeyFile == "" && c.Key == "" {
		return ErrInvalidIdentity
	}
	return nil
}

type DecryptOpt func(cfg *DecryptConfig)

func WithKey(key string) DecryptOpt {
	return func(c *DecryptConfig) {
		c.Key = key
	}
}

func WithKeyFile(keyFile string) DecryptOpt {
	return func(c *DecryptConfig) {
		c.KeyFile = keyFile
	}
}

type Decryptor struct {
	cfg DecryptConfig
}

func (d Decryptor) Decrypt(data []byte) ([]byte, error) {
	if err := d.cfg.Valid(); err != nil {
		return nil, err
	}
	ids, err := d.cfg.Identities()
	if err != nil {
		return nil, err
	}
	return d.decryptWithIdentity(data, ids...)
}

func (d Decryptor) decryptWithIdentity(data []byte, ids ...age.Identity) ([]byte, error) {
	dr := armor.NewReader(bytes.NewReader(data))
	if len(ids) == 0 {
		return nil, ErrInvalidIdentity
	}
	rdr, err := age.Decrypt(dr, ids...)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt with age: %w", err)
	}
	result, err := io.ReadAll(rdr)
	if err != nil {
		return nil, fmt.Errorf("failed to read: %w", err)
	}
	return result, nil
}

func NewDecryptor(c *DecryptConfig, opts ...DecryptOpt) (Decryptor, error) {
	for _, opt := range opts {
		opt(c)
	}
	if err := c.Valid(); err != nil {
		return Decryptor{}, fmt.Errorf("invalid decrypt config: %w", err)
	}
	return Decryptor{cfg: *c}, nil
}
