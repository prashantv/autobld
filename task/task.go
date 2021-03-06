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
	// stdinPipe is a pipe to write Stdin to.
	stdinPipe io.WriteCloser
}

func getOutFile(confFile string, defaultFile *os.File) (*os.File, error) {
	if confFile == "" {
		return defaultFile, nil
	}
	return os.Create(confFile)
}

// New starts the binary specified in args, and returns a Task for the process.
func New(baseDir string, outFile string, errFile string, args []string) (*Task, error) {
	if !log.V("Starting task: %v", args) {
		log.L("Starting task")
	}
	cmd := exec.Command(args[0], args[1:]...)

	// Use a separate process group so we can kill the whole group.
	cmd.Dir = baseDir
	cmd.SysProcAttr = getSysProcAttrs()
	var err error
	if cmd.Stdout, err = getOutFile(outFile, os.Stdout); err != nil {
		return nil, err
	}
	if cmd.Stderr, err = getOutFile(errFile, os.Stderr); err != nil {
		return nil, err
	}
	stdinPipe, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}

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
		process:   cmd.Process,
		pgid:      pgid,
		stdinPipe: stdinPipe,
	}, nil
}
