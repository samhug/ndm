package auth_providers

import (
	"github.com/go-errors/errors"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/hcl"
	"github.com/hashicorp/hcl/hcl/ast"
	"github.com/mitchellh/mapstructure"
	"github.com/samuelhug/cfgbak/config/utilities"
)

type KeePassAuthProviderConfig struct {
	DbPath           string `mapstructure:"db_path,"`
	UnlockCredential string `mapstructure:"unlock_credential,"`
}

func (t *KeePassAuthProviderConfig) Type() string {
	return "keepass"
}

// Assert interface implementation
var _ AuthProviderConfig = &KeePassAuthProviderConfig{}

func loadKeePassAuthProviderConfigHcl(list *ast.ObjectList, providers *map[string]AuthProviderConfig) error {
	list = list.Children()
	if len(list.Items) == 0 {
		return nil
	}

	var errorAccum *multierror.Error

	for _, item := range list.Items {
		name := item.Keys[0].Token.Value().(string)

		var result = KeePassAuthProviderConfig{
			UnlockCredential: "",
		}
		var metadata mapstructure.Metadata

		// Decode the parse tree into an object map
		var parsed map[string]interface{}
		if err := hcl.DecodeObject(&parsed, item.Val); err != nil {
			return err
		}

		decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
			Metadata: &metadata,
			Result:   &result,
		})
		if err != nil {
			return errors.New("Failed constructing Decoder")
		}

		// Decode the object map into our structure
		if err := decoder.Decode(parsed); err != nil {
			errorAccum = multierror.Append(errorAccum, errors.Errorf("auth_provider 'keepass' '%s': %s", name, err))
		}

		if err = utilities.CheckForRequiredFields(&metadata, []string{"db_path"}); err != nil {
			errorAccum = multierror.Append(errorAccum, errors.Errorf("auth_provider 'keepass' '%s': %s", name, err))
		}

		// Append the result
		(*providers)[name] = &result
	}

	if errorAccum.ErrorOrNil() != nil {
		return errorAccum
	}

	return nil
}
