package auth_providers

import (
	"github.com/samuelhug/cfgbak/config/utilities"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestKeePassAuthProviderConfig_Basic(t *testing.T) {

	config_str := `
	auth_provider "keepass" "test_kp" {
		db_path = "/path/to/kb_db"
		unlock_credential = "secret"
	}
	`

	c, err := utilities.LoadStringHcl(config_str)
	require.NoError(t, err)

	list, ok := utilities.GetObjectList(c)
	require.True(t, ok)

	result := make(map[string]AuthProviderConfig)
	err = loadKeePassAuthProviderConfigHcl(list.Filter("auth_provider", "keepass"), &result)
	require.NoError(t, err)

	expected := map[string]AuthProviderConfig{
		"test_kp": &KeePassAuthProviderConfig{DbPath: "/path/to/kb_db", UnlockCredential: "secret"},
	}

	require.Equal(t, expected, result)

}
