package auth

import (
	//"fmt"
	"github.com/go-errors/errors"
	//"github.com/samuelhug/ndm/config"
)

// NewProviderPool: Constructs a new ProviderPool
func NewProviderPool() *ProviderPool {
	return &ProviderPool{
		registeredProviders:  make(map[string]Provider),
		initializedProviders: make(map[string]Provider),
	}
}

// ProviderPool: Instantiates and manages AuthProviders
type ProviderPool struct {
	registeredProviders  map[string]Provider
	initializedProviders map[string]Provider
}

// RegisterProvider: Registers an auth provider with the pool
func (t *ProviderPool) RegisterProvider(name string, provider Provider) error {
	if _, exists := t.registeredProviders[name]; exists {
		return errors.Errorf("Provider already registered with the name '%s'", name)
	}

	t.registeredProviders[name] = provider
	return nil
}

// GetProvider: Retrieves a provider by name from the pool
func (t *ProviderPool) GetProvider(name string) (Provider, error) {

	// If a provider by that name has already been initialized, return it
	if provider, exists := t.initializedProviders[name]; exists {
		return provider, nil
	}

	var provider Provider
	var registered bool

	// Unsure that a provider by that name has been registered
	if provider, registered = t.registeredProviders[name]; !registered {
		return nil, errors.Errorf("No provider registered with the name '%s'", name)
	}

	// Otherwise, initialize the provider
	if err := provider.Init(); err != nil {
		return nil, errors.WrapPrefix(err, "Unable to initialize Provider", 1)
	}
	t.initializedProviders[name] = provider

	return provider, nil
}

/*
func (t *ProviderPool) instantiateProvider(name string) (Provider, error) {
	providerCfg, ok := t.configs[name]
	if !ok {
		return nil, fmt.Errorf("Unable to locate auth provider with the name '%s'", name)
	}

	switch providerCfg.Type() {
	case "keepass":
		keepassCfg, ok := providerCfg.(config.KeepassProvider)
		if !ok {
			panic("Why?")
		}

		p, err := NewKeePassProvider(keepassCfg.Path, keepassCfg.Credential)
		if err != nil {
			return nil, fmt.Errorf("Unable to initialize auth provider: %s", err)
		}

		return nil, p
	default:
		return nil, fmt.Errorf("Unimplemented Provider type: '%s'", providerCfg.Type())
	}
}
*/
