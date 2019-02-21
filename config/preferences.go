package config

import (
	"github.com/go-errors/errors"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/hcl"
	"github.com/hashicorp/hcl/hcl/ast"
	"github.com/mitchellh/mapstructure"
	"github.com/samhug/ndm/config/utilities"
)

// PreferencesConfig represents a preferences configuration block
type PreferencesConfig struct {
	BackupDir string `mapstructure:"backup_dir,"`
	HostIP    string `mapstructure:"host_ip,"`
}

// loadPreferencesHcl
func loadPreferencesHcl(list *ast.ObjectList, preferencesCfg *PreferencesConfig) error {
	if len(list.Items) == 0 {
		return nil
	}
	if len(list.Items) > 1 {
		return errors.New(
			"only one 'preferences' block may be specified")
	}

	var errorAccum *multierror.Error
	var metadata mapstructure.Metadata

	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Metadata: &metadata,
		Result:   preferencesCfg,
	})
	if err != nil {
		return errors.New("Failed constructing Decoder")
	}

	item := list.Items[0]
	var parsed map[string]interface{}
	if err := hcl.DecodeObject(&parsed, item.Val); err != nil {
		return err
	}

	if err := decoder.Decode(parsed); err != nil {
		errorAccum = multierror.Append(errorAccum, errors.Errorf("preferences: %s", err))
	}

	if err = utilities.CheckForRequiredFields(&metadata, []string{"backup_dir"}); err != nil {
		errorAccum = multierror.Append(errorAccum, errors.Errorf("preferences: %s", err))
	}

	if errorAccum.ErrorOrNil() != nil {
		return errors.Wrap(errorAccum, 0)
	}

	// Check for invalid keys
	validKeys := map[string]struct{}{
		"backup_dir": struct{}{},
		"host_ip":    struct{}{},
	}
	for _, item := range list.Items {
		if len(item.Keys) == 0 {
			continue
		}

		k := item.Keys[0].Token.Value().(string)
		if _, ok := validKeys[k]; ok {
			continue
		}

		return errors.Errorf("preferences: Unrecognized key '%s'", k)
	}

	return nil
}
