package auth

import (
	"github.com/go-errors/errors"
	"golang.org/x/crypto/ssh"
)

type Auth interface {
	GetSSHClientConfig() (*ssh.ClientConfig, error)
	GetAttribute(attr_name string) (string, error)
}

var AuthNotFound = errors.Errorf("Unable to find auth")

type Provider interface {
	Init() error
	Lookup(key string) (Auth, error)
}

func GetDefaultClientConfig() *ssh.ClientConfig {
	config := &ssh.ClientConfig{}
	config.SetDefaults()
	// TODO: Allow specification of host keys
	config.HostKeyCallback = ssh.InsecureIgnoreHostKey()
	return config
}
