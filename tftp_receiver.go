package main

import (
	"bytes"
	"fmt"
	"github.com/pin/tftp"
	"io"
	"log"
	"os"
	"strings"
	"sync"
	"time"
	"github.com/pkg/errors"
)

type ReceivedFile struct {
	Name string
	Data bytes.Buffer
}

func NewTFTPReceiver(publicAddr string) *TFTPReceiver {

	// Create the error channel
	errChannel := make(chan error, 3)

	return &TFTPReceiver{
		PublicAddr: publicAddr,
		recvHooks:  make(map[string]chan ReceivedFile),
		mutex:      &sync.Mutex{},
		errChannel: errChannel,
	}
}

type TFTPReceiver struct {
	PublicAddr string
	server     *tftp.Server
	recvHooks  map[string]chan ReceivedFile
	mutex      *sync.Mutex
	errChannel chan error
}

func (r *TFTPReceiver) GetErrorChannel() chan error {
	return r.errChannel
}

func (r *TFTPReceiver) ExpectFile(name string, ch chan ReceivedFile) {
	r.mutex.Lock()
	r.recvHooks[name] = ch
	r.mutex.Unlock()
}

func (r *TFTPReceiver) Run() {
	// Launch a TFTP server to recieve the incoming file
	r.server = tftp.NewServer(nil, r.tftpRecvHandler)
	r.server.SetTimeout(5 * time.Second)

	log.Print("Starting TFTP Server...")

	go func(server *tftp.Server, errChannel chan error) {
		err := server.ListenAndServe(":69") // blocks until s.Shutdown() is called
		if err != nil {
			errChannel <- errors.Errorf("TFTP Server: %s", err)
		}
	}(r.server, r.errChannel)
}

func (r *TFTPReceiver) Stop() {
	r.server.Shutdown()
}

func (r *TFTPReceiver) tftpRecvHandler(filename string, wt io.WriterTo) error {
	log.Print("Recieving File on TFTP Server...")

	var destChan chan ReceivedFile
	found := false

	// Ensure that the incoming file is one we're expecting
	// We only check for a prefix match as some devices require a file extension to be specified.
	r.mutex.Lock()
	for hookFilename, ch := range r.recvHooks {
		if strings.HasPrefix(filename, hookFilename) {
			destChan = ch
			found = true
			break
		}
	}
	r.mutex.Unlock()
	if !found {
		// TODO: This message will never be seen. Need to add logging to github.com/pin/tftp
		// TODO: See: https://github.com/pin/tftp/blob/master/server.go#L94
		return fmt.Errorf("unexpected incoming file (%s)", filename)
	}

	var buf bytes.Buffer

	n, err := wt.WriteTo(&buf)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return err
	}

	destChan <- ReceivedFile{
		Name: filename,
		Data: buf,
	}

	log.Printf("TFTP: %d bytes received\n", n)
	return nil
}
