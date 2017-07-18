package auth

import (
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
	"testing"
)

func TestProviderPool_New(t *testing.T) {

	var err error

	pool := NewProviderPool()

	staticProvider := NewStaticProvider()

	err = staticProvider.AddAuth("test", "User Name", "Password", map[string]string{})
	require.NoError(t, err)

	err = pool.RegisterProvider("staticTest", staticProvider)
	require.NoError(t, err)

	// Attempt to re-register it (Should fail)
	err = pool.RegisterProvider("staticTest", staticProvider)
	require.Error(t, err)

	p, err := pool.GetProvider("staticTest")
	require.NoError(t, err)

	require.Equal(t, staticProvider, p)
}

func TestProviderPool_GetProvider(t *testing.T) {

	kp, err := NewKeePassProvider(TEST_DB_PATH, TEST_DB_PASS)
	require.NoError(t, err)

	auth, err := kp.Lookup("test/Sample Entry")
	require.NoError(t, err)

	expected_config := ssh.ClientConfig{
		User: "User Name",
		Auth: []ssh.AuthMethod{
			ssh.Password("Password"),
		},
	}

	config, err := auth.GetSSHClientConfig()
	require.NoError(t, err)
	require.Equal(t, expected_config.User, config.User)

	_, err = kp.Lookup("test2/Sample Entry")
	require.Error(t, err)

	_, err = kp.Lookup("test/Fake Entry")
	require.Error(t, err)
}
