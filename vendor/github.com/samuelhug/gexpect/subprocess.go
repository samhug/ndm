// +build !windows

package gexpect

import (
	"bufio"
	"errors"
	"io"
	"os"
	"os/exec"

	shell "github.com/kballard/go-shellquote"
	"github.com/kr/pty"
)

type ExpectSubprocess struct {
	ExpectIO
	Cmd    *exec.Cmd
	closer io.Closer
}

func SpawnAtDirectory(command string, directory string) (*ExpectSubprocess, error) {
	expect, err := _spawn(command)
	if err != nil {
		return nil, err
	}
	expect.Cmd.Dir = directory
	return _start(expect)
}

func Command(command string) (*ExpectSubprocess, error) {
	expect, err := _spawn(command)
	if err != nil {
		return nil, err
	}
	return expect, nil
}

func (expect *ExpectSubprocess) Start() error {
	_, err := _start(expect)
	return err
}

func Spawn(command string) (*ExpectSubprocess, error) {
	expect, err := _spawn(command)
	if err != nil {
		return nil, err
	}
	return _start(expect)
}

func (expect *ExpectSubprocess) Close() error {
	if err := expect.Cmd.Process.Kill(); err != nil {
		return err
	}
	if err := expect.closer.Close(); err != nil {
		return err
	}
	return nil
}

func (expect *ExpectSubprocess) Interact() {
	defer expect.Cmd.Wait()
	io.Copy(os.Stdout, &expect.ExpectIO.buf.b)
	go io.Copy(os.Stdout, expect.ExpectIO.buf.rw)
	go io.Copy(expect.ExpectIO.buf.rw, os.Stdin)
}

func (expect *ExpectSubprocess) Wait() error {
	return expect.Cmd.Wait()
}

func _start(expect *ExpectSubprocess) (*ExpectSubprocess, error) {
	f, err := pty.Start(expect.Cmd)
	if err != nil {
		return nil, err
	}
	expect.ExpectIO.buf.rw = bufio.NewReadWriter(bufio.NewReader(f), bufio.NewWriter(f))
	expect.closer = f

	return expect, nil
}

func _spawn(command string) (*ExpectSubprocess, error) {
	wrapper := new(ExpectSubprocess)

	wrapper.ExpectIO.outputBuffer = nil

	splitArgs, err := shell.Split(command)
	if err != nil {
		return nil, err
	}
	numArguments := len(splitArgs) - 1
	if numArguments < 0 {
		return nil, errors.New("gexpect: No command given to spawn")
	}
	path, err := exec.LookPath(splitArgs[0])
	if err != nil {
		return nil, err
	}

	if numArguments >= 1 {
		wrapper.Cmd = exec.Command(path, splitArgs[1:]...)
	} else {
		wrapper.Cmd = exec.Command(path)
	}
	wrapper.ExpectIO.buf = new(buffer)

	return wrapper, nil
}
