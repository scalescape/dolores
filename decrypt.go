package dolores

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"filippo.io/age"
	"filippo.io/age/armor"
)

func Decrypt(keyFile string, d []byte) ([]byte, error) {
	if keyFile == "" {
		return nil, ErrInvalidKeyFile
	}
	f, err := os.Open(keyFile)
	if err != nil {
		return nil, fmt.Errorf("error opening keyfile %s: %w", keyFile, err)
	}
	defer f.Close()
	ids, err := age.ParseIdentities(f)
	if err != nil {
		return nil, fmt.Errorf("failed to parse identity: %w", err)
	}
	dr := armor.NewReader(bytes.NewReader(d))
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
