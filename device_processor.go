package main

import (
	//"bufio"
	"encoding/json"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/hashicorp/go-uuid"
	"github.com/robertkrimen/otto"
	"github.com/samuelhug/cfgbak/config"
	"github.com/samuelhug/gexpect"
	"golang.org/x/crypto/ssh"
	"io"
	"io/ioutil"
	"path"
	"time"
)

func NewDeviceProcessor(device *config.Device, configDir string) *DeviceProcessor {
	return &DeviceProcessor{
		device:    device,
		configDir: configDir,
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
	device    *config.Device
	configDir string
	vm        *otto.Otto
}

func (t *DeviceProcessor) connect() (*ssh.Client, error) {

	sshClientConfig := &ssh.ClientConfig{
		User: t.device.Auth.Username,
		Auth: []ssh.AuthMethod{
			ssh.Password(t.device.Auth.Password),
		},

		// Enable the use of this insecure cypher so we can interact with crappy legacy devices
		Config: ssh.Config{
			Ciphers: []string{"3des-cbc"},
		},
	}

	client, err := ssh.Dial("tcp", t.device.Addr, sshClientConfig)
	if err != nil {
		return nil, fmt.Errorf("Failed to dial: %s", err)
	}

	return client, nil
}

func (t *DeviceProcessor) startShell(client *ssh.Client) (*ssh.Session, io.WriteCloser, io.Reader, error) {

	session, err := client.NewSession()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("Failed to create session: %s", err)
	}

	stdOut, err := session.StdoutPipe()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("Failed to attach to Stdout: %s", err)
	}

	stdIn, err := session.StdinPipe()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("Failed to attach to Stdin: %s", err)
	}

	modes := ssh.TerminalModes{
		ssh.ECHO:          0,     // disable echoing
		ssh.TTY_OP_ISPEED: 38400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 38400, // output speed = 14.4kbaud
	}

	err = session.RequestPty("xterm", 0, 200, modes)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("Request for pty failed: %s", err)
	}

	err = session.Shell()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("Request for shell failed: %s", err)
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
		return nil, fmt.Errorf("Failed to initialize the expect library: %s", err)
	}

	// Set the context variables
	if err = vm.Set("device", t.device); err != nil {
		return nil, err
	}

	ctxRaw, err := ctx.Serialize()
	if err != nil {
		return nil, fmt.Errorf("Failed to serialize context variable: %s", err)
	}
	if err = vm.Set("_ctxRaw", ctxRaw); err != nil {
		return nil, fmt.Errorf("Failed to set _ctxRaw variable: %s", err)
	}
	_, err = vm.Run(`ctx = JSON.parse(_ctxRaw);`)
	if err != nil {
		return nil, fmt.Errorf("Failed to unserialize context variable: %s", err)
	}

	return vm, nil
}

func (t *DeviceProcessor) saveFile(file *ReceivedFile) error {

	dstFile := path.Join(t.configDir, fmt.Sprintf("%s.conf", t.device.Name))

	err := ioutil.WriteFile(dstFile, file.Data.Bytes(), 0644)
	if err != nil {
		return fmt.Errorf("File to save received file: %s", err)
	}

	return nil
}

func (t *DeviceProcessor) Process(reciever *TFTPReceiver) error {

	// Generate a unique filename to use during the TFTP upload
	filename, err := uuid.GenerateUUID()
	if err != nil {
		return fmt.Errorf("Device Processing Error [%s]: Failed to generate UUID: %s", t.device.Name, err)
	}

	// Create channel to recieve the file on
	recvChan := make(chan ReceivedFile, 3)

	// Register the filename and channel with the TFTP receiver
	reciever.ExpectFile(filename, recvChan)

	// Connect to the device
	client, err := t.connect()
	if err != nil {
		return fmt.Errorf("Device Processing Error [%s]: Unable to connect: %s", t.device.Name, err)
	}

	session, stdIn, stdOut, err := t.startShell(client)
	if err != nil {
		return fmt.Errorf("Device Processing Error [%s]: %s", t.device.Name, err)
	}
	defer session.Close()

	ctx := vmCtx{
		TFTPHost:     reciever.PublicAddr,
		TFTPFilename: filename,
	}

	vm, err := t.initVM(stdIn, stdOut, ctx)
	if err != nil {
		return fmt.Errorf("Device Processing Error [%s]: Failed to init VM: %s", t.device.Name, err)
	}

	//TODO
	//time.Sleep(5*time.Second)

	val, err := vm.Run(t.device.Class.Script)
	if err != nil {
		return fmt.Errorf("Device Processing Error [%s]: VM Runtime Error: %s", t.device.Name, err)
	}

	_ = val

	var recvdFile ReceivedFile

	// Wait for a maximum of 10 seconds for the file on the receive channel
	select {
	case recvdFile = <-recvChan:
	case <-time.After(10 * time.Second):
		return fmt.Errorf("Device Processing Error [%s]: Timed out waiting to receive file", t.device.Name)
	}

	// Save the received file
	if err = t.saveFile(&recvdFile); err != nil {
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
