package auth_providers

import (
	"github.com/go-errors/errors"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/hcl"
	"github.com/hashicorp/hcl/hcl/ast"
	"github.com/mitchellh/mapstructure"
	"github.com/samuelhug/cfgbak/config/utilities"
)

// StaticAuthConfig represents a auth configuration block within a auth_provider "static" "..." {} block.
type StaticAuthConfig struct {
	Username   string `mapstructure:"username,"`
	Password   string `mapstructure:"password,"`
	Attributes map[string]string  `mapstructure:"attributes,"`
}

func loadStaticAuthsHcl(list *ast.ObjectList) (map[string]*StaticAuthConfig, error) {
	//fmt.Println("list: %v", list)
	list = list.Children()
	if len(list.Items) == 0 {
		return nil, nil
	}

	// Where all the results will go
	results := make(map[string]*StaticAuthConfig, len(list.Items))

	var errorAccum *multierror.Error

	for _, item := range list.Items {
		name := item.Keys[0].Token.Value().(string)

		result := StaticAuthConfig{
			Attributes: make(map[string]string),
		}
		var metadata mapstructure.Metadata

		// Decode the parse tree into an object map
		var parsed map[string]interface{}
		if err := hcl.DecodeObject(&parsed, item.Val); err != nil {
			return nil, err
		}

		decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
			Metadata: &metadata,
			Result:   &result,
			WeaklyTypedInput: true, // Needed to for attributes
			ErrorUnused: true,
		})
		if err != nil {
			return nil, errors.New("Failed constructing Decoder")
		}

		// Decode the object map into our structure
		if err := decoder.Decode(parsed); err != nil {
			errorAccum = multierror.Append(errorAccum, errors.Errorf("auth '%s': %s", name, err))
		}

		if err = utilities.CheckForRequiredFields(&metadata, []string{"username", "password"}); err != nil {
			errorAccum = multierror.Append(errorAccum, errors.Errorf("auth '%s': %s", name, err))
		}

		// Append the result
		results[name] = &result
	}

	if errorAccum.ErrorOrNil() != nil {
		return nil, errorAccum
	}

	return results, nil
}
