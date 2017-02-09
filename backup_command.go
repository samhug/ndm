package main

import (
	"flag"
	"fmt"
	"github.com/mitchellh/cli"
	"github.com/ryanuber/go-glob"
	"github.com/samuelhug/cfgbak/config"
	"log"
	"strings"
)

type BackupCommand struct {
}

var _ cli.Command = &BackupCommand{}

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

	cfg, err := config.Load(cfgPath)
	if err != nil {
		log.Printf("Unable to load configuration: %s\n", err)
		return 1
	}

	deviceList := filterDevices(cfg.Devices, deviceFilter)
	if len(deviceList) == 0 {
		log.Print("No devices mached the specified filter")
		return 0
	}

	tftpReceiver := NewTFTPReceiver(cfg.Preferences.HostIP)
	tftpReceiver.Run()

	for _, device := range deviceList {
		p := NewDeviceProcessor(device, cfg.Preferences.BackupDir)

		err := p.Process(tftpReceiver)
		if err != nil {
			log.Printf(err.Error())
		}
	}

	tftpReceiver.Stop()

	return 0
}

func filterDevices(devices map[string]*config.Device, filter string) map[string]*config.Device {

	filteredDevices := make(map[string]*config.Device)

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
