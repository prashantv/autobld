// +build !windows

package task

import (
	"os/exec"
	"syscall"

	"github.com/prashantv/autobld/log"
)

func getSysProcAttrs() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{Setpgid: true}
}

func getPgID(cmd *exec.Cmd) (int, error) {
	return syscall.Getpgid(cmd.Process.Pid)
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
