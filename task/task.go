package task

import (
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/prashantv/autobld/log"
)

// Task is used to run and close/kill an external process.
type Task struct {
	process *os.Process
	// pgid is the process group ID, used when killing the task.
	pgid int
}

// underlyingReader is used to force a separate file to be used to connect
// Stdin, as the same file cannot be shared when using a separate process group.
type underlyingReader struct {
	rdr io.Reader
}

func (u underlyingReader) Read(p []byte) (n int, err error) {
	return u.rdr.Read(p)
}

// New starts the binary specified in args, and returns a Task for the process.
func New(baseDir string, args []string) (*Task, error) {
	if !log.V("Starting task: %v", args) {
		log.L("Starting task")
	}
	cmd := exec.Command(args[0], args[1:]...)

	// Use a separate process group so we can kill the whole group.
	cmd.Dir = baseDir
	cmd.SysProcAttr = getSysProcAttrs()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = underlyingReader{os.Stdin}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("error starting command: %v", err)
	}
	pgid, err := getPgID(cmd)
	if err != nil {
		// If we cannot get the pgid, kill the process and return an error.
		cmd.Process.Kill()
		return nil, err
	}

	return &Task{
		process: cmd.Process,
		pgid:    pgid,
	}, nil
}
