package utilities

import (
	"fmt"
	"github.com/go-errors/errors"
	"github.com/hashicorp/hcl"
	"github.com/hashicorp/hcl/hcl/ast"
	"github.com/mitchellh/mapstructure"
	"io/ioutil"
)

// LoadFileHcl reads a file into a Configurable object
func LoadFileHcl(root string) (*ast.File, error) {
	// Read the HCL file and prepare for parsing
	d, err := ioutil.ReadFile(root)
	if err != nil {
		return nil, fmt.Errorf(
			"Error reading %s: %s", root, err)
	}

	f, err := LoadStringHcl(string(d))
	if err != nil {
		return nil, fmt.Errorf(
			"Error parsing %s: %s", root, err)
	}

	return f, nil
}

// LoadFileHcl reads a string into a Configurable object
func LoadStringHcl(d string) (*ast.File, error) {
	// Parse it
	hclRoot, err := hcl.Parse(d)
	if err != nil {
		return nil, err
	}

	return hclRoot, nil
}

func GetObjectList(f *ast.File) (*ast.ObjectList, bool) {
	list, ok := f.Node.(*ast.ObjectList)
	return list, ok
}

// CheckForRequiredFields checks that there exists a field for each field in requiredFields.
func CheckForRequiredFields(metadata *mapstructure.Metadata, requiredFields []string) error {
	fieldsPresent := make(map[string]struct{}, len(metadata.Keys))
	var present struct{}
	for _, fieldName := range metadata.Keys {
		fieldsPresent[fieldName] = present
	}

	for _, fieldName := range requiredFields {
		if _, ok := fieldsPresent[fieldName]; !ok {
			return errors.Errorf("Missing required field '%s'", fieldName)
		}
	}

	return nil
}
