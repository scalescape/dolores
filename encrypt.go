package dolores

import (
	"bytes"
	"fmt"

	"filippo.io/age"
	"filippo.io/age/armor"
)

type Encryptor struct {
	recipients []age.Recipient
}

func (e *Encryptor) Encrypt(vars []Variable) ([]byte, error) {
	buf := &bytes.Buffer{}
	arm := armor.NewWriter(buf)
	w, err := age.Encrypt(arm, e.recipients...)
	if err != nil {
		return nil, fmt.Errorf("error encrypting: %w", err)
	}
	for _, v := range vars {
		if _, err := w.Write(v.Data()); err != nil {
			return nil, fmt.Errorf("error writing data: %w", err)
		}
	}
	if err := w.Close(); err != nil {
		return nil, fmt.Errorf("error closing writer: %w", err)
	}
	if err := arm.Close(); err != nil {
		return nil, fmt.Errorf("error closing arm writer: %w", err)
	}
	return buf.Bytes(), nil
}

func NewEncryptor(keys ...string) (*Encryptor, error) {
	recps := make([]age.Recipient, len(keys))
	for i, key := range keys {
		recp, err := age.ParseX25519Recipient(key)
		if err != nil {
			return nil, fmt.Errorf("error parsing %d key: %w", i, err)
		}
		recps[i] = recp
	}
	return &Encryptor{recipients: recps}, nil
}
