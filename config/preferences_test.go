package config

import (
	//"strings"
	"github.com/samuelhug/ndm/config/utilities"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestPreferencesBasic(t *testing.T) {

	config_str := `
preferences {
	backup_dir = "./router-configs/"
	host_ip = "10.10.10.10"
}
	`
	c, err := utilities.LoadStringHcl(config_str)
	require.NoError(t, err)

	list, ok := utilities.GetObjectList(c)
	require.True(t, ok)

	result := &PreferencesConfig{}

	err = loadPreferencesHcl(list.Filter("preferences"), result)
	require.NoError(t, err)

	require.Equal(t, &PreferencesConfig{BackupDir: "./router-configs/", HostIP: "10.10.10.10"}, result)

}

func TestPreferences_NoHostIP(t *testing.T) {

	config_str := `
preferences {
	backup_dir = "./router-configs/"
}
	`
	c, err := utilities.LoadStringHcl(config_str)
	require.NoError(t, err)

	list, ok := utilities.GetObjectList(c)
	require.True(t, ok)

	result := &PreferencesConfig{}

	err = loadPreferencesHcl(list.Filter("preferences"), result)
	require.NoError(t, err)

	require.Equal(t, &PreferencesConfig{BackupDir: "./router-configs/", HostIP: ""}, result)

}
