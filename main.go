package main

import (
	"fmt"
	"github.com/mitchellh/cli"
	"log"
	"os"
)

const (
	CMD_NAME = "ndm"
	VERSION  = "v0.1.0"
	AUTHOR   = "Sam Hug"
)

func printHeader() {
	fmt.Printf("%s %s by %s\n", CMD_NAME, VERSION, AUTHOR)
}

func main() {
	c := cli.NewCLI(CMD_NAME, VERSION)
	c.Args = os.Args[1:]
	c.Commands = map[string]cli.CommandFactory{
		"backup": backupCommandFactory,
	}

	printHeader()

	exitStatus, err := c.Run()
	if err != nil {
		log.Println(err)
	}

	os.Exit(exitStatus)
}

func backupCommandFactory() (cli.Command, error) {
	return &BackupCommand{}, nil
}
