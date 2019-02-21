package config

import (
	"github.com/samhug/ndm/config/auth_providers"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestConfig_Basic(t *testing.T) {

	buf := `
preferences {
	backup_dir = "/backup_dir/path"
	host_ip = "10.10.10.10"
}

auth_provider "static" "basic" {
	auth "testA" {
		username = "john.doe"
		password = "secret"
	}
}

device_class "classA" {
	backup_target "target1" {
		macro = "MACRO"
	}
	backup_target "target2" {
		macro = "MACRO"
	}
}

device "classA" "deviceA" {
	address = "127.0.0.1:22"
	auth = "basic:testA"
}
	`
	result, err := LoadString(buf)
	require.NoError(t, err)

	expected := &Config{
		Preferences: &PreferencesConfig{BackupDir: "/backup_dir/path", HostIP: "10.10.10.10"},

		AuthProviders: map[string]auth_providers.AuthProviderConfig{
			"basic": &auth_providers.StaticAuthProviderConfig{
				Auths: map[string]*auth_providers.StaticAuthConfig{
					"testA": {Username: "john.doe", Password: "secret", Attributes: map[string]string{}},
				},
			},
		},

		DeviceClasses: map[string]*DeviceClassConfig{
			"classA": {BackupTargets: map[string]*BackupTargetConfig{"target1": {Macro: "MACRO"}, "target2": {Macro: "MACRO"}}},
		},

		DeviceGroups: map[string]*DeviceGroupConfig{
			"": {map[string]*DeviceConfig{
				"deviceA": {Name: "deviceA", ClassName: "classA", Address: "127.0.0.1:22", AuthProvider: "basic", AuthPath: "testA"},
			}},
		},
	}
	require.Equal(t, expected, result)
}

func TestConfig_Include(t *testing.T) {
	buf := `
include "test_data/test_include.conf" {}

preferences {
	backup_dir = "/backup_dir/path"
	host_ip = "10.10.10.10"
}

auth_provider "static" "basic" {
	auth "testA" {
		username = "john.doe"
		password = "secret"
	}
}

device "classA" "deviceA" {
	address = "127.0.0.1:22"
	auth = "basic:testA"
}
	`
	result, err := LoadString(buf)
	require.NoError(t, err)

	expected := &Config{
		Preferences: &PreferencesConfig{BackupDir: "/backup_dir/path", HostIP: "10.10.10.10"},

		AuthProviders: map[string]auth_providers.AuthProviderConfig{
			"basic": &auth_providers.StaticAuthProviderConfig{
				Auths: map[string]*auth_providers.StaticAuthConfig{
					"testA": {Username: "john.doe", Password: "secret", Attributes: map[string]string{}},
				},
			},
		},

		DeviceClasses: map[string]*DeviceClassConfig{
			"classA": {BackupTargets: map[string]*BackupTargetConfig{"target1": {Macro: "MACRO"}, "target2": {Macro: "MACRO"}}},
		},

		DeviceGroups: map[string]*DeviceGroupConfig{
			"": {map[string]*DeviceConfig{
				"deviceA": {Name: "deviceA", ClassName: "classA", Address: "127.0.0.1:22", AuthProvider: "basic", AuthPath: "testA"},
			}},
		},
	}
	require.Equal(t, expected, result)
}
