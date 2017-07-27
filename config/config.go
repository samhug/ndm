package config

import (
	"github.com/go-errors/errors"
	"github.com/hashicorp/hcl/hcl/ast"
	"github.com/samuelhug/ndm/config/auth_providers"
	"github.com/samuelhug/ndm/config/utilities"
	"path"
)

type Config struct {
	ConfigDir     string
	ConfigName    string
	Preferences   *PreferencesConfig
	AuthProviders map[string]auth_providers.AuthProviderConfig
	DeviceClasses map[string]*DeviceClassConfig
	DeviceGroups  map[string]*DeviceGroupConfig
}

func loadIncludes(list *ast.ObjectList, cfg *Config) error {
	list = list.Children()
	if len(list.Items) == 0 {
		return nil
	}

	for _, item := range list.Items {
		name := item.Keys[0].Token.Value().(string)

		includePath := path.Join(cfg.ConfigDir, name)

		f, err := utilities.LoadFileHcl(includePath)
		if err != nil {
			return errors.Errorf("include '%s': %s", name, err)
		}

		// Top-level item should be the object list
		list, ok := f.Node.(*ast.ObjectList)
		if !ok {
			return errors.New("error parsing: config doesn't contain a root object")
		}

		err = loadConfigHcl(list, cfg)
		if err != nil {
			return errors.Errorf("include '%s': %s", name, err)
		}
	}

	return nil
}

func loadConfigHcl(list *ast.ObjectList, cfg *Config) error {
	// Include
	if o := list.Filter("include"); len(o.Items) > 0 {
		err := loadIncludes(o, cfg)
		if err != nil {
			return err
		}
	}

	// Preferences
	if o := list.Filter("preferences"); len(o.Items) > 0 {
		if err := loadPreferencesHcl(o, cfg.Preferences); err != nil {
			return err
		}
	}

	if o := list.Filter("auth_provider"); len(o.Items) > 0 {
		if err := auth_providers.LoadAuthProviderConfigHcl(o, &cfg.AuthProviders); err != nil {
			return err
		}
	}

	// DeviceClasses
	if o := list.Filter("device_class"); len(o.Items) > 0 {
		if err := loadDeviceClassConfigsHcl(o, &cfg.DeviceClasses); err != nil {
			return err
		}
	}

	// Devices
	if o := list.Filter("device"); len(o.Items) > 0 {
		if err := loadDeviceConfigsHcl(o, &cfg.DeviceGroups[""].Devices, &cfg.DeviceClasses, &cfg.AuthProviders); err != nil {
			return err
		}
	}

	// Device Groups
	if o := list.Filter("device_group"); len(o.Items) > 0 {
		if err := loadDeviceGroupConfigsHcl(o, &cfg.DeviceGroups, &cfg.DeviceClasses, &cfg.AuthProviders); err != nil {
			return err
		}
	}

	// Check for invalid keys
	validKeys := map[string]struct{}{
		"include":       struct{}{},
		"preferences":   struct{}{},
		"auth_provider": struct{}{},
		"device_class":  struct{}{},
		"device_group":  struct{}{},
		"device":        struct{}{},
	}
	for _, item := range list.Items {
		if len(item.Keys) == 0 {
			continue
		}

		k := item.Keys[0].Token.Value().(string)
		if _, ok := validKeys[k]; ok {
			continue
		}

		return errors.Errorf("Unrecognized key '%s'", k)
	}

	return nil
}
