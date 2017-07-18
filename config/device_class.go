package config

import (
	"github.com/go-errors/errors"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/hcl"
	"github.com/hashicorp/hcl/hcl/ast"
	"github.com/mitchellh/mapstructure"
	"github.com/samuelhug/cfgbak/config/utilities"
)

// BackupTargetConfig represents a target configuration block
type BackupTargetConfig struct {
	Macro string `mapstructure:"macro,"`
}

type DeviceClassConfig struct {
	BackupTargets map[string]*BackupTargetConfig
}

func loadDeviceClassConfigsHcl(list *ast.ObjectList, deviceClassCfgs *map[string]*DeviceClassConfig) error {
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
			return errors.Errorf("device_class '%s': backup_target should be an object", name)
		}

		backupTargets, err := loadBackupTargetConfigHcl(listVal.Filter("backup_target"))
		if err != nil {
			return err
		}
		device_class := &DeviceClassConfig{BackupTargets: backupTargets}

		if _, ok := (*deviceClassCfgs)[name]; ok {
			return errors.Errorf("device_class '%s': device_class already exists with that name", name)
		}

		// Append the result
		(*deviceClassCfgs)[name] = device_class
	}

	return nil
}

func loadBackupTargetConfigHcl(list *ast.ObjectList) (map[string]*BackupTargetConfig, error) {
	list = list.Children()
	if len(list.Items) == 0 {
		return nil, nil
	}

	// Where all the results will go
	results := make(map[string]*BackupTargetConfig, len(list.Items))

	var errorAccum *multierror.Error

	for _, item := range list.Items {
		name := item.Keys[0].Token.Value().(string)

		var result BackupTargetConfig
		var metadata mapstructure.Metadata

		// Decode the parse tree into an object map
		var parsed map[string]interface{}
		if err := hcl.DecodeObject(&parsed, item.Val); err != nil {
			return nil, err
		}

		decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
			Metadata: &metadata,
			Result:   &result,
		})
		if err != nil {
			return nil, errors.New("Failed constructing Decoder")
		}

		// Decode the object map into our structure
		if err := decoder.Decode(parsed); err != nil {
			errorAccum = multierror.Append(errorAccum, errors.Errorf("backup_target '%s': %s", name, err))
		}

		if err = utilities.CheckForRequiredFields(&metadata, []string{"macro"}); err != nil {
			errorAccum = multierror.Append(errorAccum, errors.Errorf("backup_target '%s': %s", name, err))
		}

		// Append the result
		results[name] = &result
	}

	if errorAccum.ErrorOrNil() != nil {
		return nil, errors.Wrap(errorAccum, 0)
	}

	return results, nil
}
