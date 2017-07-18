package main

import (
	"github.com/go-errors/errors"
	"github.com/robertkrimen/otto"
	"github.com/samuelhug/cfgbak/config"
)

func initDeviceClassTargets(deviceClassTargetCfgs map[string]*config.BackupTargetConfig) (map[string]*DeviceClassTarget, error) {
	deviceClassTargets := make(map[string]*DeviceClassTarget)

	for name, deviceClassTargetCfg := range deviceClassTargetCfgs {
		target, err := NewDeviceClassTarget(name, deviceClassTargetCfg.Macro)
		if err != nil {
			return nil, errors.Errorf("Unable to initialize DeviceClassTarget(%s): %s", name, err)
		}

		deviceClassTargets[name] = target
	}

	return deviceClassTargets, nil
}

func NewDeviceClassTarget(name string, macroSrc string) (*DeviceClassTarget, error) {
	macro, err := otto.New().Compile("", macroSrc)
	if err != nil {
		return nil, errors.Errorf("Unable to compile JavaScript macro: %s", err)
	}

	return &DeviceClassTarget{
		Name:  name,
		Macro: macro,
	}, nil
}

type DeviceClassTarget struct {
	Name  string
	Macro *otto.Script
}

func initDeviceClasses(deviceClassCfgs map[string]*config.DeviceClassConfig) (map[string]*DeviceClass, error) {

	deviceClasses := make(map[string]*DeviceClass)

	for name, deviceClassCfg := range deviceClassCfgs {
		targets, err := initDeviceClassTargets(deviceClassCfg.BackupTargets)
		if err != nil {
			return nil, errors.Errorf("Unable to initialize DeviceClass(%s): %s", name, err)
		}
		if len(targets) == 0 {
			return nil, errors.Errorf("DeviceClass '%s': No BackupTargets defined", name)
		}
		deviceClasses[name] = &DeviceClass{
			Targets: targets,
		}
	}

	return deviceClasses, nil
}

type DeviceClass struct {
	Targets map[string]*DeviceClassTarget
}
