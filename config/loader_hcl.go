package config

import (
	//"bytes"
	"fmt"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/hcl"
	"github.com/hashicorp/hcl/hcl/ast"
	"github.com/mitchellh/mapstructure"
	//"github.com/davecgh/go-spew/spew"
	//"io"
	//"os"
	"io/ioutil"
)

type configurable struct {
	File string
	Root *ast.File
}

func Load(path string) (*Config, error) {
	c, err := loadFileHcl(path)
	if err != nil {
		return nil, err
	}

	config, err := c.Config()
	if err != nil {
		return nil, err
	}

	return config, nil
}

func loadFileHcl(root string) (*configurable, error) {
	// Read the HCL file and prepare for parsing
	d, err := ioutil.ReadFile(root)
	if err != nil {
		return nil, fmt.Errorf(
			"Error reading %s: %s", root, err)
	}

	c, err := loadStringHcl(string(d))
	if err != nil {
		return nil, fmt.Errorf(
			"Error parsing %s: %s", root, err)
	}

	c.File = root

	return c, nil
}

func loadStringHcl(d string) (*configurable, error) {
	// Parse it
	hclRoot, err := hcl.Parse(d)
	if err != nil {
		return nil, err
	}

	// Start building the result
	result := &configurable{
		Root: hclRoot,
	}

	return result, nil
}

func (t *configurable) Config() (*Config, error) {

	validKeys := map[string]struct{}{
		"preferences":  struct{}{},
		"auth":         struct{}{},
		"device_class": struct{}{},
		"device": struct{}{},
	}

	// Top-level item should be the object list
	list, ok := t.Root.Node.(*ast.ObjectList)
	if !ok {
		return nil, fmt.Errorf("error parsing: config doesn't contain a root object")
	}

	// Start building up the actual configuration.
	config := new(Config)

	// Preferences
	if o := list.Filter("preferences"); len(o.Items) > 0 {
		var err error
		config.Preferences, err = loadPreferencesHcl(o)
		if err != nil {
			return nil, err
		}
	}

	// Auth
	if o := list.Filter("auth"); len(o.Items) > 0 {
		var err error
		config.Auths, err = loadAuthsHcl(o)
		if err != nil {
			return nil, err
		}
	}

	// DeviceClasses
	if o := list.Filter("device_class"); len(o.Items) > 0 {
		var err error
		config.DeviceClasses, err = loadDeviceClassesHcl(o)
		if err != nil {
			return nil, err
		}
	}

	// Devices
	if o := list.Filter("device"); len(o.Items) > 0 {
		var err error
		config.Devices, err = loadDevicesHcl(o, config.DeviceClasses, config.Auths)
		if err != nil {
			return nil, err
		}
	}

	// Check for invalid keys
	for _, item := range list.Items {
		if len(item.Keys) == 0 {
			// Not sure how this would happen, but let's avoid a panic
			continue
		}

		k := item.Keys[0].Token.Value().(string)
		if _, ok := validKeys[k]; ok {
			continue
		}

		return nil, fmt.Errorf("unrecognized key '%s'", k)
		//config.unknownKeys = append(config.unknownKeys, k)
	}

	return config, nil

}

func loadPreferencesHcl(list *ast.ObjectList) (*Preferences, error) {
	if len(list.Items) == 0 {
		return nil, nil
	}
	if len(list.Items) > 1 {
		return nil, fmt.Errorf(
			"only one 'preferences' block may be specified")
	}

	var errorAccum *multierror.Error
	var result Preferences
	var metadata mapstructure.Metadata

	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Metadata: &metadata,
		Result:   &result,
	})
	if err != nil {
		return nil, fmt.Errorf("Failed constructing Decoder")
	}

	item := list.Items[0]
	var parsed map[string]interface{}
	if err := hcl.DecodeObject(&parsed, item.Val); err != nil {
		return nil, err
	}

	if err := decoder.Decode(parsed); err != nil {
		errorAccum = multierror.Append(errorAccum, err.(*mapstructure.Error).WrappedErrors()...)
	}

	if err = checkForRequiredFields(&metadata, []string{"backup_dir", "host_ip"}); err != nil {
		errorAccum = multierror.Append(errorAccum, err.(*multierror.Error).WrappedErrors()...)
	}

	if errorAccum.ErrorOrNil() != nil {
		return nil, errorAccum
	}

	return &result, nil
}

func loadAuthsHcl(list *ast.ObjectList) (map[string]*Auth, error) {
	list = list.Children()
	if len(list.Items) == 0 {
		return nil, nil
	}

	// Where all the results will go
	results := make(map[string]*Auth, len(list.Items))

	var errorAccum *multierror.Error

	for _, item := range list.Items {
		n := item.Keys[0].Token.Value().(string)

		var result Auth
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
			return nil, fmt.Errorf("Failed constructing Decoder")
		}

		// Decode the object map into our structure
		if err := decoder.Decode(parsed); err != nil {
			errorAccum = multierror.Append(errorAccum, err.(*mapstructure.Error).WrappedErrors()...)
		}

		// Set the name field
		result.Name = n

		if err = checkForRequiredFields(&metadata, []string{"username", "password"}); err != nil {
			errorAccum = multierror.Append(errorAccum, err.(*multierror.Error).WrappedErrors()...)
		}

		// Append the result
		results[n] = &result
	}

	if errorAccum.ErrorOrNil() != nil {
		return nil, errorAccum
	}

	return results, nil
}

func loadDeviceClassesHcl(list *ast.ObjectList) (map[string]*DeviceClass, error) {
	list = list.Children()
	if len(list.Items) == 0 {
		return nil, nil
	}

	var errorAccum *multierror.Error

	// Where all the results will go
	results := make(map[string]*DeviceClass, len(list.Items))

	for _, item := range list.Items {
		n := item.Keys[0].Token.Value().(string)

		var result DeviceClass
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
			return nil, fmt.Errorf("Failed constructing Decoder")
		}

		// Decode the object map into our structure
		if err := decoder.Decode(parsed); err != nil {
			errorAccum = multierror.Append(errorAccum, err.(*mapstructure.Error).WrappedErrors()...)
		}

		// Set the name field
		result.Name = n

		if err = checkForRequiredFields(&metadata, []string{"script"}); err != nil {
			errorAccum = multierror.Append(errorAccum, err.(*multierror.Error).WrappedErrors()...)
		}

		// Append the result
		results[n] = &result
	}

	if errorAccum.ErrorOrNil() != nil {
		return nil, errorAccum
	}

	return results, nil
}

func loadDevicesHcl(list *ast.ObjectList, deviceClasses map[string]*DeviceClass, auths map[string]*Auth) (map[string]*Device, error) {

	type hclDevice struct {
		Name      string `mapstructure:",key"`
		ClassName string `mapstructure:"class,"`
		Addr      string `mapstructure:"addr,"`
		AuthName  string `mapstructure:"auth,"`
	}

	list = list.Children()
	if len(list.Items) == 0 {
		return nil, nil
	}

	var errorAccum *multierror.Error

	// Where all the results will go
	results := make(map[string]*Device, len(list.Items))

	for _, item := range list.Items {
		n := item.Keys[0].Token.Value().(string)

		var rawResult hclDevice
		var metadata mapstructure.Metadata

		// Decode the parse tree into an object map
		var parsed map[string]interface{}
		if err := hcl.DecodeObject(&parsed, item.Val); err != nil {
			return nil, err
		}

		decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
			Metadata: &metadata,
			Result:   &rawResult,
		})
		if err != nil {
			return nil, fmt.Errorf("Failed constructing Decoder")
		}

		// Decode the object map into our structure
		if err := decoder.Decode(parsed); err != nil {
			errorAccum = multierror.Append(errorAccum, err.(*mapstructure.Error).WrappedErrors()...)
		}

		// Set the name field
		rawResult.Name = n

		if err = checkForRequiredFields(&metadata, []string{"class", "auth"}); err != nil {
			errorAccum = multierror.Append(errorAccum, err.(*multierror.Error).WrappedErrors()...)
		}

		class, ok := deviceClasses[rawResult.ClassName]
		if !ok {
			errorAccum = multierror.Append(errorAccum, fmt.Errorf("device_class '%s' doesn't exist", rawResult.ClassName))
		}

		auth, ok := auths[rawResult.AuthName]
		if !ok {
			errorAccum = multierror.Append(errorAccum, fmt.Errorf("auth '%s' doesn't exist", rawResult.AuthName))
		}

		// Append the result
		results[n] = &Device{
			Name:  rawResult.Name,
			Addr:  rawResult.Addr,
			Class: class,
			Auth:  auth,
		}
	}

	if errorAccum.ErrorOrNil() != nil {
		return nil, errorAccum
	}

	return results, nil
}

// checkForRequiredFields
func checkForRequiredFields(metadata *mapstructure.Metadata, requiredFields []string) error {
	fieldsPresent := make(map[string]struct{}, len(metadata.Keys))
	var present struct{}
	for _, fieldName := range metadata.Keys {
		fieldsPresent[fieldName] = present
	}

	var errorAccum *multierror.Error
	for _, fieldName := range requiredFields {
		if _, ok := fieldsPresent[fieldName]; !ok {
			errorAccum = multierror.Append(errorAccum, fmt.Errorf("'%s' was not specified", fieldName))
		}
	}

	if errorAccum.ErrorOrNil() != nil {
		return errorAccum
	}

	return nil
}
