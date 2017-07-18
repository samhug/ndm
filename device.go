package main

import (
	"github.com/go-errors/errors"
	"github.com/samuelhug/cfgbak/config"
	"path"
	"github.com/samuelhug/cfgbak/auth"
)

func initDevices(deviceGroupCfgs map[string]*config.DeviceGroupConfig, deviceClasses map[string]*DeviceClass, authProviders *auth.ProviderPool) (map[string]*Device, error) {

	devices := make(map[string]*Device)

	// Iterate through the Device Groups
	for groupName, deviceGroupCfg := range deviceGroupCfgs {

		// Iterate through the groups Devices
		for deviceName, deviceCfg := range deviceGroupCfg.Devices {

			var deviceFullName string
			if groupName != "" {
				deviceFullName = path.Join(groupName, deviceName)
			} else {
				deviceFullName = deviceName
			}

			deviceClass, ok := deviceClasses[deviceCfg.ClassName]
			if !ok {
				return nil, errors.Errorf("Unable to initialize Device(%s): DeviceClass(%s) not found", deviceName, deviceCfg.ClassName)
			}

			authProvider, err := authProviders.GetProvider(deviceCfg.AuthProvider)
			if err != nil {
				return nil, errors.Errorf("Failed to retrieve AuthProvider(%s) from the pool: %s",
					deviceCfg.AuthProvider, err)
			}

			auth, err := authProvider.Lookup(deviceCfg.AuthPath)
			if err != nil {
				return nil, errors.Errorf("Lookup failed for Auth(%s) in AuthProvider(%s): %s",
					deviceCfg.AuthPath, deviceCfg.AuthProvider, err)
			}

			devices[deviceFullName] = &Device{
				Name:             deviceFullName,
				Class:            deviceClass,
				Address:          deviceCfg.Address,
				AuthProviderName: deviceCfg.AuthProvider,
				AuthPath:         deviceCfg.AuthPath,
				Auth:             auth,
			}
		}
	}

	return devices, nil
}

type Device struct {
	Name             string
	Class            *DeviceClass
	Address          string
	AuthProviderName string
	AuthPath         string
	Auth             auth.Auth
}
