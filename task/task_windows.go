package task

import (
	"fmt"
	"os/exec"
	"syscall"
	"unsafe"

	"github.com/prashantv/autobld/log"
)

var (
	modkernel32                  = syscall.NewLazyDLL("kernel32.dll")
	procGenerateConsoleCtrlEvent = modkernel32.NewProc("GenerateConsoleCtrlEvent")
)

func getSysProcAttrs() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
	}
}

func getPgID(cmd *exec.Cmd) (int, error) {
	return cmd.Process.Pid, nil
}

// Interrupt sends Ctrl-Break to the task's process group.
func (t *Task) Interrupt() error {
	log.VV("Requested Ctrl-Break on task")
	r, _, err := procGenerateConsoleCtrlEvent.Call(syscall.CTRL_BREAK_EVENT, uintptr(t.pgid))
	if r != 0 {
		return fmt.Errorf("r = %v err: %v", r, err)
	}
	return nil
}

// Kill kills all child processes of pgid, since there is no way to kill a process
// group in Windows.
func (t *Task) Kill() error {
	log.V("Kill task")

	// In case killing all child processes fails.
	defer t.process.Kill()

	parentMap, err := loadProcessMap()
	if err != nil {
		return err
	}
	if errs := killChildProcesses(parentMap, uint32(t.process.Pid)); len(errs) > 0 {
		log.L("killChildProcesses errors: %v", errs)
		return fmt.Errorf("failed to kill child processes: %v", errs)
	}
	return nil
}

// killChildProcesses will kill pid, then recurse through all children.
// It does not return on error, but continues and returns all encountered errors.
func killChildProcesses(parentMap map[uint32][]uint32, pid uint32) []error {
	// Kill the given process ID.
	var errors []error
	if err := killProcess(pid); err != nil {
		errors = append(errors, err)
	}

	// Kill all the child processes if any.
	children, ok := parentMap[pid]
	delete(parentMap, pid)
	if !ok {
		return errors
	}

	for _, c := range children {
		if err := killChildProcesses(parentMap, c); err != nil {
			errors = append(errors, err...)
		}
	}
	return errors
}

// killProcess terminates a given pid using TerminateProcess.
func killProcess(pid uint32) error {
	handle, err := syscall.OpenProcess(syscall.PROCESS_TERMINATE, false /* inheritHandle */, pid)
	if err != nil {
		return err
	}
	defer syscall.CloseHandle(handle)

	log.VV("TerminateProcess(%v) with handle %v", pid, handle)
	return syscall.TerminateProcess(handle, 1)
}

// loadProcessMap returns a map of processes and their children.
func loadProcessMap() (map[uint32][]uint32, error) {
	snapshot, err := syscall.CreateToolhelp32Snapshot(syscall.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return nil, err
	}
	defer syscall.CloseHandle(snapshot)

	var proc syscall.ProcessEntry32
	proc.Size = uint32(unsafe.Sizeof(proc))
	if err := syscall.Process32First(snapshot, &proc); err != nil {
		return nil, err
	}

	parentMap := make(map[uint32][]uint32)
	for {
		parentMap[proc.ParentProcessID] = append(parentMap[proc.ParentProcessID], proc.ProcessID)
		if err := syscall.Process32Next(snapshot, &proc); err != nil {
			if errNo, ok := err.(syscall.Errno); ok && errNo == syscall.ERROR_NO_MORE_FILES {
				return parentMap, nil
			}
			return nil, err
		}
	}
}
