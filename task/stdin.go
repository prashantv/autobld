package task

import (
	"io"
	"os"
)

// This file is used to manage how os.Stdin is copied to tasks.
// Simply passing os.Stdin directly to a cmd.Stdin causes data to be lsot on reload.
// This is because the underlying copier calls Read, which blocks on Stdin, and then
// fails on the Write call, and the data read from stdin is lost.
// To avoid this, we set up a channel to which data is written to from Stdin
// and a task specific goroutine will use select to read/write, or will notice
// when the task has ended, and stop reading from Stdin.

var stdinChan = make(chan []byte)

func init() {
	go stdinLoop()
}

func stdinLoop() {
	const BUFFERSIZE = 4096
	bytes := make([]byte, BUFFERSIZE)
	bytes2 := make([]byte, BUFFERSIZE)
	for {
		n, err := os.Stdin.Read(bytes)
		if err != nil {
			stdinChan <- nil
			return
		}
		stdinChan <- bytes[:n]
		// The receiver of bytes will read it, and we do not want to overwrite the buffer
		// while they are reading it, so we swap buffers.
		bytes, bytes2 = bytes2, bytes
	}
}

// copyStdin takes data from os.Stdin sent over stdinChan and writes it to
// a task's Stdin.
func copyStdin(stdinPipe io.WriteCloser, closer <-chan struct{}) {
	for {
		select {
		case data := <-stdinChan:
			if len(data) == 0 {
				stdinPipe.Close()
				return
			}
			stdinPipe.Write(data)
		case <-closer:
			stdinPipe.Close()
			return
		}
	}
}
