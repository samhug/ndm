package auth

import (
	"fmt"
	"github.com/go-errors/errors"
	"golang.org/x/crypto/ssh"
)

type StaticAuth struct {
	username   string
	password   string
	attributes map[string]string
}

func (t *StaticAuth) GetSSHClientConfig() (*ssh.ClientConfig, error) {
	config := GetDefaultClientConfig()
	config.User = t.username
	config.Auth = []ssh.AuthMethod{
		// TODO: Use the PublicKeys method for remote authentication.
		//ssh.PublicKeys(signer),
		ssh.Password(t.password),
	}
	return config, nil
}
func (t *StaticAuth) GetAttribute(attr_name string) (string, error) {
	val, ok := t.attributes[attr_name]
	if !ok {
		keys := make([]string, 0, len(t.attributes))
		for k := range t.attributes {
			keys = append(keys, k)
		}
		return "", errors.Errorf("Attribute not found. Available attributes are %s", keys)
	}
	return val, nil
}

var _ Auth = &KeePassAuth{}

func NewStaticProvider() *StaticProvider {
	return &StaticProvider{
		auths: make(map[string]*StaticAuth),
	}
}

type StaticProvider struct {
	auths map[string]*StaticAuth
}

func (t *StaticProvider) Init() error {
	return nil
}

func (t *StaticProvider) Lookup(path string) (Auth, error) {

	if a, ok := t.auths[path]; ok {
		return a, nil
	}
	return nil, errors.WrapPrefix(AuthNotFound, fmt.Sprintf("Unknown auth %s", path), 0)
}

func (t *StaticProvider) AddAuth(path string, username string, password string, attributes map[string]string) error {
	if _, exists := t.auths[path]; exists {
		return errors.Errorf("An auth with the path '%s' as already added", path)
	}

	t.auths[path] = &StaticAuth{
		username:   username,
		password:   password,
		attributes: attributes,
	}

	return nil
}

// Assert interface compatibility
var _ Provider = &StaticProvider{}
