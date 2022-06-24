package encryption

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_encryptor(t *testing.T) {
	e := New()
	data := []byte("testDataToEncrypt")

	encrypted, err := e.EncryptData(data)
	require.NoError(t, err)
	require.NotEmpty(t, encrypted)

	decrypted, err := e.DecryptData(encrypted)
	require.NoError(t, err)
	require.Equal(t, data, decrypted)
}

func Test_encryptor_GenerateRandom(t *testing.T) {
	e := New()
	got, err := e.GenerateRandom(16)
	require.NoError(t, err)
	require.Len(t, got, 16)
	require.NotEmpty(t, got)
}
