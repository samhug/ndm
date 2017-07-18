package main

import (
	"flag"
	"fmt"
	"github.com/go-errors/errors"
	"github.com/mitchellh/cli"
	"github.com/ryanuber/go-glob"
	"github.com/samuelhug/cfgbak/auth"
	"github.com/samuelhug/cfgbak/config"
	"github.com/samuelhug/cfgbak/config/auth_providers"
	"log"
	"strings"
)

type BackupCommand struct {
}

var _ cli.Command = &BackupCommand{}

func initAuthProviderPool(providersCfg map[string]auth_providers.AuthProviderConfig) (*auth.ProviderPool, error) {
	pool := auth.NewProviderPool()

	for providerName, providerCfg := range providersCfg {
		var provider auth.Provider
		var err error
		switch providerCfg.Type() {
		case "static":
			cfg := providerCfg.(*auth_providers.StaticAuthProviderConfig)
			static_provider := auth.NewStaticProvider()
			for path, auth := range cfg.Auths {
				err = static_provider.AddAuth(path, auth.Username, auth.Password, auth.Attributes)
				if err != nil {
					return nil, errors.Errorf("Unable to add Auth(%s) to StaticAuthProvider(%s): %s", path, providerName, err)
				}
			}
			provider = static_provider
		case "keepass":
			cfg := providerCfg.(*auth_providers.KeePassAuthProviderConfig)
			provider, err = auth.NewKeePassProvider(cfg.DbPath, cfg.UnlockCredential)
			if err != nil {
				return nil, errors.Errorf("Unable to initialise KeePassAuthProvider(%s): %s", providerName, err)
			}
		default:
			return nil, errors.Errorf("Unsupported AuthProvider type (%s)", providerCfg.Type())
		}
		err = pool.RegisterProvider(providerName, provider)
		if err != nil {
			return nil, errors.Errorf("Unable to register provider '%s': %s", providerName, err)
		}
	}

	return pool, nil
}

func (t *BackupCommand) Run(args []string) int {

	cmdname := "backup"

	var cfgPath string

	cmdFlags := flag.NewFlagSet(cmdname, flag.ContinueOnError)
	cmdFlags.StringVar(&cfgPath, "config", "config.conf", "path")
	cmdFlags.Usage = func() { fmt.Printf(t.Help()) }
	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}
	args = cmdFlags.Args()

	if len(args) > 1 {
		log.Print("Too many command line arguments. Configuration path expected.")
		fmt.Printf(t.Help())
		return 1
	}

	deviceFilter := "*"
	if len(args) > 0 {
		deviceFilter = args[0]
	}

	cfg, err := config.LoadFile(cfgPath)
	if err != nil {
		log.Printf("Unable to load configuration: %s\n", err)
		return 1
	}

	hostIP := cfg.Preferences.HostIP
	if hostIP == "" {
		hostIP, err = getExternalIPAddr()
		if err != nil {
			log.Printf("Unable to detect external interface IP address and no HostIP was specified: %s\n", err)
			return 1
		}
		log.Printf("No host IP was specified, auto detected %s\n", hostIP)
	}

	authProviderPool, err := initAuthProviderPool(cfg.AuthProviders)
	if err != nil {
		log.Printf("Unable to initialize the auth provider pool: %s\n", err)
		return 1
	}

	deviceClasses, err := initDeviceClasses(cfg.DeviceClasses)
	if err != nil {
		log.Printf("Error initializing device classes: %s\n", err)
		return 1
	}

	devices, err := initDevices(cfg.DeviceGroups, deviceClasses, authProviderPool)
	if err != nil {
		log.Printf("Error initializing devices: %s\n", err)
		return 1
	}

	deviceList := filterDevices(devices, deviceFilter)
	if len(deviceList) == 0 {
		log.Print("No devices mached the specified filter")
		return 0
	}

	tftpReceiver := NewTFTPReceiver(hostIP)
	tftpReceiver.Run()

	for _, device := range deviceList {
		p := NewDeviceProcessor(device, authProviderPool, cfg.Preferences.BackupDir)

		err := p.Process(tftpReceiver)
		if err != nil {
			log.Printf("Device Processing Error '%s': %s", device.Name, err)
		}
	}

	tftpReceiver.Stop()

	return 0
}

func filterDevices(devices map[string]*Device, filter string) map[string]*Device {

	filteredDevices := make(map[string]*Device)

	for k, v := range devices {
		if glob.Glob(filter, k) {
			filteredDevices[k] = v
		}
	}

	return filteredDevices
}

func (t *BackupCommand) Help() string {
	helpText := `
Usage: cfgbak backup [options] [device-filter]
	Backs up the configuration for device matching the filter. If no filter is specified,
	will backup all devices.
Options:
	--config <path>		Path to the configuation file.
`
	return strings.TrimSpace(helpText)
}

func (t *BackupCommand) Synopsis() string {
	return "Backs up the configuration for the specified devices"
}
