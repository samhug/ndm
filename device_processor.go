package main

import (
	"encoding/json"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/go-errors/errors"
	"github.com/hashicorp/go-uuid"
	"github.com/robertkrimen/otto"
	"github.com/samuelhug/gexpect"
	"github.com/samuelhug/ndm/auth"
	"golang.org/x/crypto/ssh"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"time"
)

// NewDeviceProcessor: Initializes a new DeviceProcessor object
func NewDeviceProcessor(device *Device, authProviders *auth.ProviderPool, backupDir string) *DeviceProcessor {
	return &DeviceProcessor{
		device:        device,
		authProviders: authProviders,
		configDir:     backupDir,
	}
}

type vmCtx struct {
	TFTPHost     string
	TFTPFilename string
}

func (ctx *vmCtx) Serialize() (string, error) {
	b, err := json.Marshal(ctx)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

type DeviceProcessor struct {
	authProviders *auth.ProviderPool
	device        *Device
	configDir     string
	vm            *otto.Otto
}

func (t *DeviceProcessor) connect() (*ssh.Client, error) {

	sshClientConfig, err := t.device.Auth.GetSSHClientConfig()
	if err != nil {
		return nil, errors.Errorf("Failed to construct SSHClientConfig from Auth(%s): %s",
			t.device.AuthPath, err)
	}

	// Enable the use of this insecure cypher so we can interact with crappy legacy devices
	sshClientConfig.Config.Ciphers = append(sshClientConfig.Config.Ciphers, "3des-cbc")

	client, err := ssh.Dial("tcp", t.device.Address, sshClientConfig)
	if err != nil {
		return nil, errors.Errorf("Failed to connect to Device(%s): %s", t.device.Name, err)
	}

	return client, nil
}

func (t *DeviceProcessor) startShell(client *ssh.Client) (*ssh.Session, io.WriteCloser, io.Reader, error) {

	session, err := client.NewSession()
	if err != nil {
		return nil, nil, nil, errors.Errorf("Failed to create terminal session via SSH: %s", err)
	}

	stdOut, err := session.StdoutPipe()
	if err != nil {
		return nil, nil, nil, errors.Errorf("Failed to attach to Stdout: %s", err)
	}

	stdIn, err := session.StdinPipe()
	if err != nil {
		return nil, nil, nil, errors.Errorf("Failed to attach to Stdin: %s", err)
	}

	modes := ssh.TerminalModes{
		ssh.ECHO:          0,     // disable echoing
		ssh.TTY_OP_ISPEED: 38400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 38400, // output speed = 14.4kbaud
	}

	err = session.RequestPty("xterm", 0, 200, modes)
	if err != nil {
		return nil, nil, nil, errors.Errorf("Request for pty failed: %s", err)
	}

	err = session.Shell()
	if err != nil {
		return nil, nil, nil, errors.Errorf("Request for shell failed: %s", err)
	}

	return session, stdIn, stdOut, nil
}

func (t *DeviceProcessor) initVM(stdIn io.WriteCloser, stdOut io.Reader, ctx vmCtx) (*otto.Otto, error) {

	expect := gexpect.NewExpectIO(stdOut, stdIn)
	expect.Capture()

	vm := otto.New()

	var err error

	vm.SetDebuggerHandler(func(vm *otto.Otto) {
		spew.Printf("vm.Context: %v", vm.Context())
	})

	// Initialize the Expect library
	if err = ottoExpect(vm, expect); err != nil {
		return nil, errors.Errorf("Failed to initialize the expect library: %s", err)
	}

	err = vm.Set("getAuthAttr", func(call otto.FunctionCall) otto.Value {
		attrName := call.Argument(0).String()

		val, err := t.device.Auth.GetAttribute(attrName)
		if err != nil {
			return vm.MakeCustomError("AttrError", fmt.Sprintf("Unable to find auth attribute '%s': %s", attrName, err))
		}

		ottoVal, err := vm.ToValue(val)
		if err != nil {
			return vm.MakeCustomError("TypeError", "Unable to convert AuthAttr to a string")
		}

		return ottoVal
	})

	// Set the context variables
	if err = vm.Set("device", t.device); err != nil {
		return nil, err
	}

	ctxRaw, err := ctx.Serialize()
	if err != nil {
		return nil, errors.Errorf("Failed to serialize context variable: %s", err)
	}
	if err = vm.Set("_ctxRaw", ctxRaw); err != nil {
		return nil, errors.Errorf("Failed to set _ctxRaw variable: %s", err)
	}
	_, err = vm.Run(`ctx = JSON.parse(_ctxRaw);`)
	if err != nil {
		return nil, errors.Errorf("Failed to unserialize context variable: %s", err)
	}

	return vm, nil
}

func (t *DeviceProcessor) saveFile(backupTarget *DeviceClassTarget, file *ReceivedFile) error {

	dstPath := path.Join(t.configDir, t.device.Name, fmt.Sprintf("%s.conf", backupTarget.Name))

	dirPath, _ := path.Split(dstPath)

	// Create the parent directory structure if needed
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		err = os.MkdirAll(dirPath, os.ModePerm)
		if err != nil {
			return errors.Errorf("Parent directory '%s' doesn't exist and an error occurred while trying to create it: %s", dirPath, err)
		}
	}

	err := ioutil.WriteFile(dstPath, file.Data.Bytes(), 0644)
	if err != nil {
		return errors.Errorf("Unable to write to file '%s': %s", dstPath, err)
	}

	return nil
}

func (t *DeviceProcessor) Process(reciever *TFTPReceiver) error {
	for target_name, _ := range t.device.Class.Targets {
		log.Printf("Processing backup target '%s':'%s'", t.device.Name, target_name)

		err := t.ProcessTarget(target_name, reciever)
		if err != nil {
			return errors.Errorf("target %s: %s", target_name, err)
		}
	}

	return nil
}

func (t *DeviceProcessor) ProcessTarget(target_name string, reciever *TFTPReceiver) error {

	backupTarget := t.device.Class.Targets[target_name]

	// Generate a unique filename to use during the TFTP upload
	filename, err := uuid.GenerateUUID()
	if err != nil {
		return errors.Errorf("Failed to generate UUID: %s", err)
	}

	// Create channel to recieve the file on
	recvChan := make(chan ReceivedFile, 3)

	// Register the filename and channel with the TFTP receiver
	reciever.ExpectFile(filename, recvChan)

	// Connect to the device
	client, err := t.connect()
	if err != nil {
		return errors.Errorf("Unable to connect: %s", err)
	}

	session, stdIn, stdOut, err := t.startShell(client)
	if err != nil {
		return errors.Errorf("Failed to start shell: %s", err)
	}
	defer session.Close()

	ctx := vmCtx{
		TFTPHost:     reciever.PublicAddr,
		TFTPFilename: filename,
	}

	vm, err := t.initVM(stdIn, stdOut, ctx)
	if err != nil {
		return errors.Errorf("Failed to init JavaScript VM: %s", err)
	}

	_, err = vm.Run(backupTarget.Macro)
	if err != nil {
		return errors.Errorf("JavaScript VM Runtime Error: %s", err)
	}

	var recvdFile ReceivedFile

	// Wait for a maximum of 60 seconds for the file on the receive channel
	select {
	case err = <-reciever.GetErrorChannel():
		log.Printf("TFTP Receiver error: %s\n", err)
	case recvdFile = <-recvChan:
	case <-time.After(60 * time.Second):
		return errors.Errorf("Timed out waiting to receive file over TFTP")
	}

	// Save the received file
	if err = t.saveFile(backupTarget, &recvdFile); err != nil {
		return err
	}

	return nil
}

func ottoExpect(vm *otto.Otto, expect *gexpect.ExpectIO) error {

	var err error

	err = vm.Set("dbgDump", func(call otto.FunctionCall) otto.Value {

		v, _ := call.Argument(0).Export()
		fmt.Printf(">>> dbgDump >>>:\n%v<<< dbgDump <<<\n", spew.Sdump(v))

		return otto.Value{}
	})
	if err != nil {
		return err
	}

	err = vm.Set("dbgLog", func(call otto.FunctionCall) otto.Value {

		fmt.Printf("dbgLog: %s\n", call.Argument(0).String())

		return otto.Value{}
	})
	if err != nil {
		return err
	}

	// function expect(val string) string {}
	err = vm.Set("expect", func(call otto.FunctionCall) otto.Value {

		// TODO: Make timeout configurable
		err := expect.ExpectTimeout(call.Argument(0).String(), 15*time.Second)
		if err != nil {
			//fmt.Printf(">>> ExpectCollected >>>:\n%v<<< ExpectCollected <<<\n", spew.Sdump(expect.Collect()))
			panic(vm.MakeCustomError("ExpectError", err.Error()))
		}

		return otto.Value{}
	})
	if err != nil {
		return err
	}

	err = vm.Set("readLine", func(call otto.FunctionCall) otto.Value {

		line, err := expect.ReadLine()
		if err != nil {
			panic(vm.MakeCustomError("ExpectError", err.Error()))
		}

		v, err := call.Otto.ToValue(line)
		if err != nil {
			panic(err.Error())
		}

		return v
	})
	if err != nil {
		return err
	}

	err = vm.Set("sendLine", func(call otto.FunctionCall) otto.Value {

		err := expect.SendLine(call.Argument(0).String())
		if err != nil {
			panic(vm.MakeCustomError("ExpectError", err.Error()))
		}

		return otto.Value{}
	})
	if err != nil {
		return err
	}

	return nil
}
