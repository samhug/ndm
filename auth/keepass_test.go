package auth

import (
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
	"testing"
)

var TEST_DB_PATH = "testdata/test.kdbx"
var TEST_DB_PASS = "abcdefg12345678"

func TestKeePassBasic(t *testing.T) {

	kp, err := NewKeePassProvider(TEST_DB_PATH, TEST_DB_PASS)
	require.NoError(t, err)

	a, err := kp.Lookup("test/Sample Entry")
	require.NoError(t, err)

	expected_config := &ssh.ClientConfig{
		User: "User Name",
		Auth: []ssh.AuthMethod{
			ssh.Password("Password"),
		},
	}

	config, err := a.GetSSHClientConfig()
	require.NoError(t, err)
	require.Equal(t, expected_config.User, config.User)

	attributeVal, err := a.GetAttribute("test_attribute")
	require.NoError(t, err)
	require.Equal(t, "Test Value", attributeVal)

	_, err = kp.Lookup("test/General/Nested Entry")
	require.NoError(t, err)

	_, err = kp.Lookup("test2/Sample Entry")
	require.Error(t, err)

	_, err = kp.Lookup("test/Fake Entry")
	require.Error(t, err)
}
