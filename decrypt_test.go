package dolores

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	alicePubKey  = "age1t5nvyz07qh3ud07sar7ftw3qlm93lmprgtc23frx47d7k6ml2dnqju0yr0"
	bobPubKey    = "age1rgqud79me4jpkh2ly74h6ye42pdj2lqjdnh5sy6s4ln3gzee6y6s5rq0a0"
	aliceKeyFile = "./testdata/alice.key"
	bobKeyFile   = "./testdata/bob.key"
)

func TestShouldDecryptSuccessfully(t *testing.T) {
	data, err := os.ReadFile("./testdata/sample.age")
	require.NoError(t, err)

	result, err := Decrypt(aliceKeyFile, data)

	expectedPlain, err := os.ReadFile("./testdata/sample.env")
	require.NoError(t, err)
	require.NoError(t, err, "error decrypting")
	assert.Equal(t, expectedPlain, result)
}

func TestShouldEncryptLoadedFileSuccessfully(t *testing.T) {
	enc, err := NewEncryptor(alicePubKey)
	require.NoError(t, err, "creating encryptor failed")
	data, err := LoadEnvFile("./testdata/sample.env")
	require.NoError(t, err, "failed to load file")

	d, err := enc.Encrypt(data.Variables)

	require.NoError(t, err, "failed to encrypt")
	result := string(d)
	lines := strings.Split(result, "\n")
	assert.NotEmpty(t, result)
	assert.Equal(t, "-----BEGIN AGE ENCRYPTED FILE-----", lines[0])
	assert.Equal(t, "-----END AGE ENCRYPTED FILE-----", lines[len(lines)-2])
}

func TestShouldDecryptEncryptedFileWithMultipleKeys(t *testing.T) {
	enc, err := NewEncryptor(alicePubKey, bobPubKey)
	require.NoError(t, err, "creating encryptor failed")
	data, err := LoadEnvFile("./testdata/sample.env")
	require.NoError(t, err, "failed to load file")
	d, err := enc.Encrypt(data.Variables)
	require.NoError(t, err, "failed to encrypt")

	result, err := Decrypt(aliceKeyFile, d)

	require.NoError(t, err)
	expectedPlain := `key1=value1
key2=valuetwo
key3=random
key4=something
`
	assert.Equal(t, string(expectedPlain), string(result))
}
