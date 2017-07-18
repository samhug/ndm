package auth

import (
	"fmt"
	"github.com/go-errors/errors"
	"github.com/tobischo/gokeepasslib"
	"golang.org/x/crypto/ssh"
	"os"
	"strings"
)

var UsernameNotFound = errors.Errorf("Unable to find a username in KeePass entry")
var PasswordNotFound = errors.Errorf("Unable to find a password in KeePass entry")
var AttributeNotFound = errors.Errorf("Unable to find attribute in KeePass entry")

//var PrivateKeyNotFound = errors.Errorf("Unable to find a private key in KeePass entry")

type KeePassAuth struct {
	entry *gokeepasslib.Entry
}

func (t *KeePassAuth) GetSSHClientConfig() (*ssh.ClientConfig, error) {
	usernameVal := t.entry.Get("UserName")
	if usernameVal == nil {
		return nil, errors.New(UsernameNotFound)
	}

	passwordVal := t.entry.Get("Password")
	if usernameVal == nil {
		return nil, errors.New(PasswordNotFound)
	}

	/*
		// If there is an SSH key in the entry, we will use that for authentication. Otherwise fallback to user/pass method
		sshKey := t.entry.Get("SSHPrivateKey")
		if sshKey == nil {
			return nil, errors.New(PrivateKeyNotFound)
		}
		// Create the Signer for this private key.
		signer, err := ssh.ParsePrivateKey([]byte(sshKey.Value.Content))
		if err != nil {
			return nil, errors.Wrap("unable to parse private key", err)
		}
	*/

	config := GetDefaultClientConfig()
	config.User = usernameVal.Value.Content
	config.Auth = []ssh.AuthMethod{
		// TODO: Use the PublicKeys method for remote authentication.
		//ssh.PublicKeys(signer),
		ssh.Password(passwordVal.Value.Content),
	}
	return config, nil
}
func (t *KeePassAuth) GetAttribute(attr_name string) (string, error) {
	attrVal := t.entry.Get(attr_name)
	if attrVal == nil {
		return "", errors.New(AttributeNotFound)
	}

	return attrVal.Value.Content, nil
}

var _ Auth = &KeePassAuth{}

func NewKeePassProvider(dbPath string, unlockPassword string) (*KeePassProvider, error) {

	file, err := os.Open(dbPath)
	if err != nil {
		return nil, errors.Errorf("Unable to open KeePass Db: %s", err)
	}

	db := gokeepasslib.NewDatabase()
	db.Credentials = gokeepasslib.NewPasswordCredentials(unlockPassword)
	err = gokeepasslib.NewDecoder(file).Decode(db)
	if err != nil {
		return nil, errors.Errorf("Error decoding KeePass Db: %s", err)
	}

	if err := db.UnlockProtectedEntries(); err != nil {
		return nil, errors.Errorf("Error Unlocking KeePass Db: %s", err)
	}

	return &KeePassProvider{
		db: db,
	}, nil
}

type KeePassProvider struct {
	db *gokeepasslib.Database
}

func (t *KeePassProvider) Init() error {
	return nil
}

func (t *KeePassProvider) Lookup(path string) (Auth, error) {

	entry, err := t.ResolveEntryPath(path)
	if err != nil {
		return nil, err
	}

	usernameVal := entry.Get("UserName")
	if usernameVal == nil {
		return nil, errors.New("Unable to find a username in entry")
	}

	passwordVal := entry.Get("Password")
	if passwordVal == nil {
		return nil, errors.New("Unable to find a password in entry")
	}

	return &KeePassAuth{
		entry: entry,
	}, nil
}

func (t *KeePassProvider) ResolveEntryPath(path string) (*gokeepasslib.Entry, error) {
	parts := strings.Split(path, "/")

	// Entries are required to be in a group, so their path will have at least two parts
	if len(parts) < 2 {
		return nil, fmt.Errorf("Path '%s' is too short to be a velid entry", path)
	}

	// Resolve the base group
	groupName := parts[0]
	group, err := resolveGroupName(t.db.Content.Root.Groups, groupName)
	if err != nil {
		return nil, fmt.Errorf("Unable to resolve entry path, no root group found named: %s", groupName)
	}

	// Loop through and resolve all the middle parts
	dbgPath := groupName
	for _, groupName = range parts[1 : len(parts)-1] {
		dbgPath += "/" + groupName
		group, err = resolveGroupName(group.Groups, groupName)
		if err != nil {
			return nil, fmt.Errorf("Unable to resolve entry path, no group found at: %s", dbgPath)
		}
	}

	// Resolve the entry
	entryName := parts[len(parts)-1]
	entry, err := resolveEntryName(group.Entries, entryName)
	if err != nil {
		return nil, fmt.Errorf("Unable to resolve entry path, no entry found at: %s", path)
	}

	return entry, nil
}

func resolveGroupName(groups []gokeepasslib.Group, name string) (*gokeepasslib.Group, error) {
	for _, group := range groups {
		if group.Name == name {
			return &group, nil
		}
	}
	return nil, errors.New("Group not found")
}

func resolveEntryName(entries []gokeepasslib.Entry, name string) (*gokeepasslib.Entry, error) {
	for _, entry := range entries {
		if entry.GetTitle() == name {
			return &entry, nil
		}
	}
	return nil, errors.New("Entry not found")
}

// Assert interface compatibility
var _ Provider = &KeePassProvider{}
