package dolores

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"

	"filippo.io/age"
)

func ParseIdentities(keyFile string) ([]age.Identity, error) {
	// process identity from keyfile
	f, err := os.Open(keyFile)
	if err != nil {
		return nil, fmt.Errorf("error opening keyfile %s: %w", keyFile, err)
	}
	defer f.Close()
	ids, err := age.ParseIdentities(f)
	if err != nil {
		return nil, fmt.Errorf("failed to parse identity: %w", err)
	}
	return ids, nil
}

func ReadPublicKey(fname string) (string, error) {
	keyFile, err := os.Open(fname)
	if err != nil {
		return "", fmt.Errorf("error opening keyfile %s: %w", fname, err)
	}
	const recipientFileSizeLimit = 1 << 24 // 16 MiB
	scanner := bufio.NewScanner(io.LimitReader(keyFile, recipientFileSizeLimit))
	var n int
	re := regexp.MustCompile(`^#\s+public key.*(age1.*)`)
	for scanner.Scan() {
		n++
		line := scanner.Text()
		if line == "" {
			continue
		}
		match := re.FindStringSubmatch(line)
		if len(match) > 1 {
			r, err := age.ParseX25519Recipient(match[1])
			if err != nil {
				return "", fmt.Errorf("malformed recipient at line %d", n)
			}
			return r.String(), nil
		}
	}
	return "", fmt.Errorf("unable to extract public key")
}
