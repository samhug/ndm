package auth_providers

import (
	"github.com/samuelhug/cfgbak/config/utilities"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestStaticAuthProviderConfig_Basic(t *testing.T) {

	config_str := `
auth_provider "static" "test" {
	auth "authA" {
		username = "bob.edwards"
		password = "secret1"
	}
	auth "authB" {
		username = "alice.jones"
		password = "secret2"
	}
}
	`
	c, err := utilities.LoadStringHcl(config_str)
	require.NoError(t, err)

	list, ok := utilities.GetObjectList(c)
	require.True(t, ok)

	result := make(map[string]AuthProviderConfig)
	err = loadStaticAuthProviderConfigHcl(list.Filter("auth_provider", "static"), &result)
	require.NoError(t, err)

	expected := map[string]AuthProviderConfig{
		"test": &StaticAuthProviderConfig{Auths: map[string]*StaticAuthConfig{
			"authA": {Username: "bob.edwards", Password: "secret1", Attributes: map[string]string{}},
			"authB": {Username: "alice.jones", Password: "secret2", Attributes: map[string]string{}},
		}},
	}
	require.Equal(t, expected, result)
}
