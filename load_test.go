package dolores

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShouldLoadPlainEnvSuccessfully(t *testing.T) {
	ef, err := LoadEnvFile("./testdata/sample.env")

	require.NoError(t, err, "should not have any errors")
	require.Equal(t, 4, len(ef.Variables))
	assert.Equal(t, Variable{[]byte("key1"), []byte("value1")}, ef.Variables[0])
	assert.Equal(t, Variable{[]byte("key2"), []byte("valuetwo")}, ef.Variables[1])
	assert.Equal(t, Variable{[]byte("key3"), []byte("random")}, ef.Variables[2])
	assert.Equal(t, Variable{[]byte("key4"), []byte("something")}, ef.Variables[3])
	assert.NotEmpty(t, ef.CreatedAt)
}
