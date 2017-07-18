package auth_providers

import (
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/hcl/hcl/ast"
)

type AuthProviderConfig interface {
	Type() string
}

func LoadAuthProviderConfigHcl(list *ast.ObjectList, providers *map[string]AuthProviderConfig) error {
	list = list.Children()
	if len(list.Items) == 0 {
		return nil
	}

	var errorAccum *multierror.Error

	// - Static
	if o := list.Filter("static"); len(o.Items) > 0 {
		err := loadStaticAuthProviderConfigHcl(o, providers)
		if err != nil {
			multierror.Append(errorAccum, err)
		}
	}

	// - Keepass
	if o := list.Filter("keepass"); len(o.Items) > 0 {
		err := loadKeePassAuthProviderConfigHcl(o, providers)
		if err != nil {
			multierror.Append(errorAccum, err)
		}
	}

	if errorAccum.ErrorOrNil() != nil {
		return errorAccum
	}

	return nil
}
