package config

import (
	"github.com/samhug/ndm/config/auth_providers"
	"github.com/samhug/ndm/config/utilities"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestParseDeviceAuthStr(t *testing.T) {
	var provider, path string
	var err error

	provider, path, err = parseDeviceAuthStr("a:b")
	require.NoError(t, err)
	require.Equal(t, "a", provider)
	require.Equal(t, "b", path)
}

func TestDeviceConfig_Basic(t *testing.T) {

	config_str := `
device "deviceClassA" "deviceA" {
	address = "10.10.10.10:22"
	auth = "providerA:auth1"
}
	`
	c, err := utilities.LoadStringHcl(config_str)
	require.NoError(t, err)

	list, ok := utilities.GetObjectList(c)
	require.True(t, ok)

	providers := map[string]auth_providers.AuthProviderConfig{
		"providerA": &auth_providers.StaticAuthProviderConfig{Auths: map[string]*auth_providers.StaticAuthConfig{
			"auth1": {Username: "bob", Password: "secret"},
		}},
	}
	deviceClasses := map[string]*DeviceClassConfig{
		"deviceClassA": {BackupTargets: map[string]*BackupTargetConfig{
			"targetA": {Macro: ""},
		}},
	}

	results := map[string]*DeviceConfig{}

	err = loadDeviceConfigsHcl(list.Filter("device"), &results, &deviceClasses, &providers)
	require.NoError(t, err)

	expected := map[string]*DeviceConfig{
		"deviceA": {Name: "deviceA", ClassName: "deviceClassA", AuthProvider: "providerA", AuthPath: "auth1", Address: "10.10.10.10:22"},
	}
	require.Equal(t, expected, results)
}
