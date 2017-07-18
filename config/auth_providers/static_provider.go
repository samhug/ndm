package auth_providers

import (
	"github.com/go-errors/errors"
	"github.com/hashicorp/hcl/hcl/ast"
)

type StaticAuthProviderConfig struct {
	Auths map[string]*StaticAuthConfig
}

func (t *StaticAuthProviderConfig) Type() string {
	return "static"
}

// Assert interface implementation
var _ AuthProviderConfig = &StaticAuthProviderConfig{}

func loadStaticAuthProviderConfigHcl(list *ast.ObjectList, providers *map[string]AuthProviderConfig) error {
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
			return errors.Errorf("auth_provoder '%s': auth should be an object", name)
		}

		auths, err := loadStaticAuthsHcl(listVal.Filter("auth"))
		if err != nil {
			return err
		}
		provider := &StaticAuthProviderConfig{Auths: auths}

		if _, ok := (*providers)[name]; ok {
			return errors.Errorf("auth_provoder '%s': auth_provider already exists with that name", name)
		}

		// Append the result
		(*providers)[name] = provider
	}

	return nil
}
