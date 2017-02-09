package config

import (
	"testing"
	//"github.com/hashicorp/hcl/hcl/ast"
	"github.com/stretchr/testify/assert"
)

func TestConfigBasic(t *testing.T) {

	buf := `

preferences {
	backup_dir = "/some/path"
}

auth "test" {
	username = "john.doe"
	password = "secret"
}

device_class "abc" {
	script = "def"
}

device_class "jkl" {
	script = "mno"
}
	`
	c, err := loadStringHcl(buf)
	assert.NoError(t, err)

	_, err = c.Config()
	assert.NoError(t, err)

	//expected := map[string]*DeviceClass{ "abc": &o1, "jkl": &o2 }
	//assert.Equal(t, expected, result)
}
