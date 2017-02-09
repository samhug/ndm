package config

import (
	//"strings"
	"github.com/hashicorp/hcl/hcl/ast"
	"github.com/stretchr/testify/assert"
	"testing"
	//"github.com/stretchr/testify/suite"
)

func TestPreferencesBasic(t *testing.T) {

	buf := `
preferences {
	backup_dir = "./router-configs/"
	host_ip = "10.10.10.10"
}
	`
	c, err := loadStringHcl(buf)
	assert.NoError(t, err)

	list, ok := c.Root.Node.(*ast.ObjectList)
	assert.True(t, ok)

	result, err := loadPreferencesHcl(list.Filter("preferences"))
	assert.NoError(t, err)

	assert.Equal(t, &Preferences{BackupDir: "./router-configs/", HostIP:"10.10.10.10"}, result)

}
