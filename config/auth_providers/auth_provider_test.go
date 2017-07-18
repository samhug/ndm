package auth_providers

import (
	"github.com/hashicorp/hcl/hcl/ast"
	"github.com/samuelhug/cfgbak/config/utilities"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestAuthProviderConfig_Basic(t *testing.T) {

	buf := `
	auth_provider "static" "test" {
		auth "authA" {
			username = "User Name"
			password = "secret stuff"
		}
	}
	auth_provider "keepass" "test_kp" {
		db_path = "/path/to/kb_db"
		unlock_credential = "secret"
	}
	`
	f, err := utilities.LoadStringHcl(buf)
	require.NoError(t, err)

	list, ok := f.Node.(*ast.ObjectList)
	require.True(t, ok)

	result := make(map[string]AuthProviderConfig)
	err = LoadAuthProviderConfigHcl(list.Filter("auth_provider"), &result)
	require.NoError(t, err)

	expected := map[string]AuthProviderConfig{
		"test": &StaticAuthProviderConfig{Auths: map[string]*StaticAuthConfig{
			"authA": {Username: "User Name", Password: "secret stuff", Attributes: map[string]string{}},
		}},
		"test_kp": &KeePassAuthProviderConfig{DbPath: "/path/to/kb_db", UnlockCredential: "secret"},
	}

	require.Equal(t, expected, result)
}
