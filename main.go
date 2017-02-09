package main

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/samuelhug/cfgbak/config"
	"log"
	"os"
)

const (
	CMDNAME = "cfgbak"
	VERSION = "0.0.1"
	AUTHOR  = "Sam Hug"
)

func main() {

	var configPath string

	app := cli.NewApp()
	app.Name = CMDNAME
	app.Usage = "A utility for backing up device configs via TFTP."
	app.Author = AUTHOR
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "config, c",
			Usage:       "Path to configuration file",
			Destination: &configPath,
		},
	}
	app.Action = func(c *cli.Context) error {

		fmt.Printf("%s v%s by %s\n\n", CMDNAME, VERSION, AUTHOR)

		//TODO:
		configPath = "config.conf"

		if configPath == "" {
			log.Fatal("You must specifiy a configuration file.\n")
		}

		cfg, err := config.Load(configPath)
		if err != nil {
			log.Fatalf("Unable to load configuration: %s\n", err)
		}

		return runApp(cfg)
	}

	app.Run(os.Args)
}

func runApp(cfg *config.Config) error {

	tftpReceiver := NewTFTPReceiver(cfg.Preferences.HostIP)
	tftpReceiver.Run()

	//spew.Dump(cfg.Devices)

	for _, device := range cfg.Devices {
		p := NewDeviceProcessor(device, cfg.Preferences.BackupDir)

		err := p.Process(tftpReceiver)
		if err != nil {
			log.Printf(err.Error())
		}
	}

	tftpReceiver.Stop()

	return nil
}
