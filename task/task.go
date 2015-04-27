package task

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/prashantv/autobld/log"
)

// Task is used to run and close/kill an external process.
type Task struct {
	process *os.Process
	// pgid is the process group ID, used when killing the task.
	pgid int
}

// New starts the binary specified in args, and returns a Task for the process.
func New(baseDir string, args []string) (*Task, error) {
	if !log.V("Starting task: %v", args) {
		log.L("Starting task")
	}
	cmd := exec.Command(args[0], args[1:]...)

	// Use a separate process group so we can kill the whole group.
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Dir = baseDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("error starting command: %v", err)
	}
	pgid, err := syscall.Getpgid(cmd.Process.Pid)
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

// Interrupt sends the same signal as a Ctrl-C to the task.
func (t *Task) Interrupt() error {
	log.VV("Requested Ctrl-C on task")
	return syscall.Kill(-t.pgid, syscall.SIGINT)
}

// Kill sends a KILL signal to the task.
func (t *Task) Kill() error {
	log.V("Kill task")
	return syscall.Kill(-t.pgid, syscall.SIGKILL)
}
