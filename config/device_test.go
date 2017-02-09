package config

import (
	"github.com/hashicorp/hcl/hcl/ast"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDeviceBasic(t *testing.T) {

	buf := `

auth "test_auth" {
	username = "a"
	password = "b"
}

device_class "test_class" {
	script = "c"
}

// ===============

device "abc" {
	class = "test_class"
	addr = "ghi"
	auth = "test_auth"
}
	`
	c, err := loadStringHcl(buf)
	assert.NoError(t, err)

	list, ok := c.Root.Node.(*ast.ObjectList)
	assert.True(t, ok)

	auths, err := loadAuthsHcl(list.Filter("auth"))
	assert.NoError(t, err)

	deviceClasses, err := loadDeviceClassesHcl(list.Filter("device_class"))
	assert.NoError(t, err)

	result, err := loadDevicesHcl(list.Filter("device"), deviceClasses, auths)
	assert.NoError(t, err)

	o1 := Device{Name: "abc", Class: &DeviceClass{Name: "test_class", Script: "c"}, Auth: &Auth{Name: "test_auth", Username: "a", Password: "b"}, Addr: "ghi"}

	expected := map[string]*Device{"abc": &o1}
	assert.Equal(t, expected, result)
}
