package config

import (
	"github.com/samuelhug/ndm/config/utilities"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestDeviceClassBasic(t *testing.T) {

	config_str := `
device_class "D_CLASS_A" {
	backup_target "TARGET_1" {
		macro = "MACRO_PLACEHOLDER_1"
	}
	backup_target "TARGET_2" {
		macro = "MACRO_PLACEHOLDER_2"
	}
}

device_class "D_CLASS_B" {
	backup_target "TARGET_1" {
		macro = "MACRO_PLACEHOLDER_1"
	}
}
	`
	c, err := utilities.LoadStringHcl(config_str)
	require.NoError(t, err)

	list, ok := utilities.GetObjectList(c)
	require.True(t, ok)

	results := map[string]*DeviceClassConfig{}

	err = loadDeviceClassConfigsHcl(list.Filter("device_class"), &results)
	require.NoError(t, err)

	expected := map[string]*DeviceClassConfig{
		"D_CLASS_A": {BackupTargets: map[string]*BackupTargetConfig{
			"TARGET_1": {Macro: "MACRO_PLACEHOLDER_1"},
			"TARGET_2": {Macro: "MACRO_PLACEHOLDER_2"},
		}},
		"D_CLASS_B": {BackupTargets: map[string]*BackupTargetConfig{
			"TARGET_1": {Macro: "MACRO_PLACEHOLDER_1"},
		}},
	}
	require.Equal(t, expected, results)
}
