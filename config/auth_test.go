package config

import (
	"github.com/hashicorp/hcl/hcl/ast"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAuthBasic(t *testing.T) {

	buf := `
auth "abc" {
	username = "def"
	password = "ghi"
}

auth "jkl" {
	username = "mno"
	password = "pqr"
}
	`
	c, err := loadStringHcl(buf)
	assert.NoError(t, err)

	list, ok := c.Root.Node.(*ast.ObjectList)
	assert.True(t, ok)

	result, err := loadAuthsHcl(list.Filter("auth"))
	assert.NoError(t, err)

	a1 := Auth{Name: "abc", Username: "def", Password: "ghi"}
	a2 := Auth{Name: "jkl", Username: "mno", Password: "pqr"}

	expected := map[string]*Auth{"abc": &a1, "jkl": &a2}
	assert.Equal(t, expected, result)
}
