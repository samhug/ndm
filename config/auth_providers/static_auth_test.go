package auth_providers

import (
	"github.com/samhug/ndm/config/utilities"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestStaticAuthConfig_Basic(t *testing.T) {

	config_str := `
auth "abc" {
	username = "def"
	password = "ghi"
}

auth "jkl" {
	username = "mno"
	password = "pqr"
	attributes {
		enable_password = "test"
	}
}
	`
	c, err := utilities.LoadStringHcl(config_str)
	require.NoError(t, err)

	list, ok := utilities.GetObjectList(c)
	require.True(t, ok)

	result, err := loadStaticAuthsHcl(list.Filter("auth"))
	require.NoError(t, err)

	expected := map[string]*StaticAuthConfig{
		"abc": {Username: "def", Password: "ghi", Attributes: map[string]string{}},
		"jkl": {Username: "mno", Password: "pqr", Attributes: map[string]string{"enable_password": "test"}},
	}
	require.Equal(t, expected, result)
}
