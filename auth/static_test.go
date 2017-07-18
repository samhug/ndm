package auth

import (
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
	"testing"
)

func TestStaticProvider(t *testing.T) {

	provider := NewStaticProvider()

	provider.AddAuth("SampleAuth", "User Name", "Password", map[string]string{})

	a, err := provider.Lookup("SampleAuth")
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

	_, err = provider.Lookup("MissingAuth")
	require.Error(t, err)

	_, err = provider.Lookup("MissingAuth2")
	require.Error(t, err)
}
