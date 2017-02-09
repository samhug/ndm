package config

import (
	"github.com/hashicorp/hcl/hcl/ast"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDeviceClassBasic(t *testing.T) {

	buf := `
device_class "abc" {
	script = "def"
}

device_class "jkl" {
	script = "mno"
}
	`
	c, err := loadStringHcl(buf)
	assert.NoError(t, err)

	list, ok := c.Root.Node.(*ast.ObjectList)
	assert.True(t, ok)

	result, err := loadDeviceClassesHcl(list.Filter("device_class"))
	assert.NoError(t, err)

	o1 := DeviceClass{Name: "abc", Script: "def"}
	o2 := DeviceClass{Name: "jkl", Script: "mno"}

	expected := map[string]*DeviceClass{"abc": &o1, "jkl": &o2}
	assert.Equal(t, expected, result)
}
