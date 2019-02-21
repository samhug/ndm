package cmd

import (
	"fmt"
	"github.com/go-errors/errors"
	"github.com/ryanuber/go-glob"
	"github.com/samhug/ndm/auth"
	"github.com/samhug/ndm/config"
	"github.com/samhug/ndm/config/auth_providers"
	"github.com/samhug/ndm/device_processor"
	"github.com/samhug/ndm/devices"
	"github.com/segmentio/go-prompt"
	"github.com/spf13/cobra"
	"log"
	"net"
)

func init() {
	rootCmd.AddCommand(backupCmd)

	backupCmd.Flags().StringVar(&cfgPath, "config", "config.hcl", "config file path")
}

var cfgPath string

var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Backup device configs",
	Run:   backupMain,
}

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
			if cfg.UnlockCredential == "" {
				fmt.Printf("Please provide the unlock credential for the KeePass database '%s'\n", cfg.DbPath)
				cfg.UnlockCredential = prompt.PasswordMasked("Password")
			}
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

func backupMain(cmd *cobra.Command, args []string) {

	deviceFilter := "*"
	if len(args) > 0 {
		deviceFilter = args[0]
	}

	cfg, err := config.LoadFile(cfgPath)
	if err != nil {
		log.Fatalln("Unable to load configuration:", err)
	}

	hostIP := cfg.Preferences.HostIP
	if hostIP == "" {
		log.Println("No external IP was specified, please choose an IP address for the TFTP server to listen on")
		hostIP, err = getExternalIPAddr()
		if err != nil {
			log.Fatalln("Unable to detect external interface IP address and no HostIP was specified:", err)
		}
		log.Printf("IP %s was selected\n", hostIP)
	}

	authProviderPool, err := initAuthProviderPool(cfg.AuthProviders)
	if err != nil {
		log.Fatalln("Unable to initialize the auth provider pool:", err)
	}

	deviceClasses, err := devices.LoadDeviceClasses(cfg.DeviceClasses)
	if err != nil {
		log.Fatalln("Error initializing device classes:", err)
	}

	devices, err := devices.LoadDevices(cfg.DeviceGroups, deviceClasses, authProviderPool)
	if err != nil {
		log.Fatalln("Error initializing devices:", err)
	}

	deviceList := filterDevices(devices, deviceFilter)
	if len(deviceList) == 0 {
		log.Fatalln("No devices mached the given filter")
	}

	tftpReceiver := device_processor.NewTFTPReceiver(hostIP)
	tftpReceiver.Run()

	for _, device := range deviceList {
		p := device_processor.NewDeviceProcessor(device, authProviderPool, cfg.Preferences.BackupDir)

		err := p.Process(tftpReceiver)
		if err != nil {
			log.Printf("Device Processing Error '%s': %s", device.Name, err)
		}
	}

	tftpReceiver.Stop()

}

func filterDevices(_devices map[string]*devices.Device, filter string) map[string]*devices.Device {

	filteredDevices := make(map[string]*devices.Device)

	for k, v := range _devices {
		if glob.Glob(filter, k) {
			filteredDevices[k] = v
		}
	}

	return filteredDevices
}

func getExternalIPAddr() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	ipAddrs := []string{}

	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return "", err
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			ipAddrs = append(ipAddrs, ip.String())
		}
	}
	if len(ipAddrs) == 0 {
		return "", errors.New("Unable to determine external interface address. Are you connected to the network?")
	}

	i := prompt.Choose("Select External IP Address", ipAddrs)
	return ipAddrs[i], nil
}
