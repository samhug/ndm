package config

import (
	"github.com/go-errors/errors"
	"github.com/hashicorp/hcl/hcl/ast"
	"github.com/samuelhug/cfgbak/config/auth_providers"
)

type DeviceGroupConfig struct {
	Devices map[string]*DeviceConfig
}

func loadDeviceGroupConfigsHcl(list *ast.ObjectList, deviceGroupCfgs *map[string]*DeviceGroupConfig, deviceClassCfgs *map[string]*DeviceClassConfig, authProviderCfgs *map[string]auth_providers.AuthProviderConfig) error {
	list = list.Children()
	if len(list.Items) == 0 {
		return nil
	}

	for _, item := range list.Items {
		name := item.Keys[0].Token.Value().(string)

		var listVal *ast.ObjectList
		if ot, ok := item.Val.(*ast.ObjectType); ok {
			listVal = ot.List
		} else {
			return errors.Errorf("device_group '%s': device should be an object", name)
		}

		childDeviceCfgs := make(map[string]*DeviceConfig)

		err := loadDeviceConfigsHcl(listVal.Filter("device"), &childDeviceCfgs, deviceClassCfgs, authProviderCfgs)
		if err != nil {
			return errors.Errorf("device_group '%s': %s", name, err)
		}
		device_group := &DeviceGroupConfig{Devices: childDeviceCfgs}

		if _, ok := (*deviceGroupCfgs)[name]; ok {
			return errors.Errorf("device_group '%s': device_group already exists with that name", name)
		}

		// Append the result
		(*deviceGroupCfgs)[name] = device_group
	}

	return nil
}
