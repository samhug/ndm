package config

import (
	"github.com/go-errors/errors"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/hcl"
	"github.com/hashicorp/hcl/hcl/ast"
	"github.com/mitchellh/mapstructure"
	"github.com/samuelhug/cfgbak/config/auth_providers"
	"github.com/samuelhug/cfgbak/config/utilities"
	"strings"
)

// DeviceConfig represents a device configuration block
type DeviceConfig struct {
	Name         string
	ClassName    string
	Address      string
	AuthProvider string
	AuthPath     string
}

// parseDeviceAuthStr parses out the provider name and path given a device auth string of the form "provider:path".
func parseDeviceAuthStr(auth_str string) (string, string, error) {
	parts := strings.SplitN(auth_str, ":", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", errors.Errorf("Invalid device auth string (%s). Auth strings must be of the format \"provider_name:auth_path\"", auth_str)
	}
	return parts[0], parts[1], nil
}

// loadDeviceConfigsHcl constructs DeviceConfig objects representing each device configuration block
func loadDeviceConfigsHcl(list *ast.ObjectList, deviceCfgs *map[string]*DeviceConfig, deviceClassCfgs *map[string]*DeviceClassConfig, authProviderCfgs *map[string]auth_providers.AuthProviderConfig) error {

	type hclDevice struct {
		Address string `mapstructure:"address,"`
		AuthStr string `mapstructure:"auth,"`
	}

	list = list.Children()
	if len(list.Items) == 0 {
		return nil
	}

	var errorAccum *multierror.Error

	for _, item := range list.Items {
		if len(item.Keys) != 2 {
			return errors.New("device block must specify a class and a name")
		}

		className := item.Keys[0].Token.Value().(string)
		name := item.Keys[1].Token.Value().(string)

		var rawResult hclDevice
		var metadata mapstructure.Metadata

		// Decode the parse tree into an object map
		var parsed map[string]interface{}
		if err := hcl.DecodeObject(&parsed, item.Val); err != nil {
			return err
		}

		decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
			Metadata: &metadata,
			Result:   &rawResult,
		})
		if err != nil {
			return errors.New("Failed constructing Decoder")
		}

		// Decode the object map into our structure
		if err := decoder.Decode(parsed); err != nil {
			errorAccum = multierror.Append(errorAccum, err.(*mapstructure.Error).WrappedErrors()...)
		}

		if err = utilities.CheckForRequiredFields(&metadata, []string{"address", "auth"}); err != nil {
			errorAccum = multierror.Append(errorAccum, errors.Errorf("device '%s': %s", name, err))
		}

		if _, ok := (*deviceClassCfgs)[className]; !ok {
			errorAccum = multierror.Append(errorAccum, errors.Errorf("device '%s': device_class '%s' doesn't exist", name, className))
		}

		auth_provider, auth_path, err := parseDeviceAuthStr(rawResult.AuthStr)
		if err != nil {
			errorAccum = multierror.Append(errorAccum, errors.Errorf("device '%s': %s", name, err))
		} else if _, ok := (*authProviderCfgs)[auth_provider]; !ok {
			errorAccum = multierror.Append(errorAccum, errors.Errorf("device '%s': auth_provider '%s' doesn't exist", name, auth_provider))
		}

		if _, ok := (*deviceCfgs)[name]; ok {
			return errors.Errorf("device '%s': device already exists with that name", name)
		}

		// Append the result
		(*deviceCfgs)[name] = &DeviceConfig{
			Name:         name,
			Address:      rawResult.Address,
			ClassName:    className,
			AuthProvider: auth_provider,
			AuthPath:     auth_path,
		}
	}

	if errorAccum.ErrorOrNil() != nil {
		return errorAccum
	}

	return nil
}
