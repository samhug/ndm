package main

import (
	"github.com/mitchellh/cli"
	"log"
	"os"
)

const (
	CMDNAME = "cfgbak"
	VERSION = "0.0.1"
)

func main() {
	c := cli.NewCLI(CMDNAME, VERSION)
	c.Args = os.Args[1:]
	c.Commands = map[string]cli.CommandFactory{
		"backup": backupCommandFactory,
	}

	exitStatus, err := c.Run()
	if err != nil {
		log.Println(err)
	}

	os.Exit(exitStatus)
}

func backupCommandFactory() (cli.Command, error) {
	return &BackupCommand{}, nil
}
