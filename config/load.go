package config

import (
	"errors"
	"github.com/hashicorp/hcl/hcl/ast"
	"github.com/samuelhug/cfgbak/config/auth_providers"
	"github.com/samuelhug/cfgbak/config/utilities"
	"path"
)

func getConfigDefaults() *Config {
	return &Config{
		ConfigDir:     "",
		ConfigName:    "",
		Preferences:   &PreferencesConfig{BackupDir: "", HostIP: ""},
		AuthProviders: map[string]auth_providers.AuthProviderConfig{},
		DeviceClasses: map[string]*DeviceClassConfig{},
		DeviceGroups:  map[string]*DeviceGroupConfig{
			"": { Devices: map[string]*DeviceConfig{}},
		},
	}
}

// LoadFile reads the contents of a file and parses it into a Config object
func LoadFile(filePath string) (*Config, error) {
	f, err := utilities.LoadFileHcl(filePath)
	if err != nil {
		return nil, err
	}

	fileDir, fileName := path.Split(filePath)
	cfg := getConfigDefaults()
	cfg.ConfigDir = fileDir
	cfg.ConfigName = fileName

	// Top-level item should be the object list
	list, ok := f.Node.(*ast.ObjectList)
	if !ok {
		return nil, errors.New("error parsing: config doesn't contain a root object")
	}

	if err = loadConfigHcl(list, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// LoadString reads the contents of cfg_str and parses it into a Config object
func LoadString(cfg_str string) (*Config, error) {
	f, err := utilities.LoadStringHcl(cfg_str)
	if err != nil {
		return nil, err
	}

	cfg := getConfigDefaults()

	// Top-level item should be the object list
	list, ok := f.Node.(*ast.ObjectList)
	if !ok {
		return nil, errors.New("error parsing: config doesn't contain a root object")
	}

	if err = loadConfigHcl(list, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
