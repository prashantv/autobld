package task

import (
	"time"

	"github.com/prashantv/autobld/config"
	"github.com/prashantv/autobld/log"
	"github.com/prashantv/autobld/sync"
)

const (
	// buildBufferTime is the buffer time after a change is detected before we Reload the task.
	buildBufferTime = time.Second

	// killBufferTime is how long after a close request to wait before killing the task.
	killBufferTime = time.Second
)

// SM is used to store state about the currently running task.
type SM struct {
	Task *Task
	c    *config.Config

	// reloadRequest is the time at which a Reload was requested.
	reloadRequest time.Time
	// Done is set to True once a task ends.
	Done sync.Bool

	// Reprocess is the channel the caller waits on to reprocess the state machine.
	Reprocess chan struct{}
	// ReloadEnded is written to when a Reload is complete (eg the task has ended).
	ReloadEnded chan struct{}
}

// NewSM returns the state maachine used to run tasks.
func NewSM(c *config.Config) *SM {
	return &SM{
		c:           c,
		Reprocess:   make(chan struct{}),
		ReloadEnded: make(chan struct{}),
	}
}

// Running returns whether the task is currently running.
func (t *SM) Running() bool {
	return t.Task != nil && !t.Done.Read()
}

// PendingClose returns whether there is a close request which hasn't yet been completed.
// Once the task is closed, PendingClose will return false, and Running will return false.
func (t *SM) PendingClose() bool {
	return !t.reloadRequest.IsZero()
}

// PastBuildBuffer returns whether we are past the buffer timeout.
func (t *SM) PastBuildBuffer() bool {
	return t.reloadRequest.Add(buildBufferTime).Before(time.Now())
}

// PastKillTime returns whether we are past the kill buffer timeout.
func (t *SM) PastKillTime() bool {
	return t.reloadRequest.Add(buildBufferTime + killBufferTime).Before(time.Now())
}

// Execute runs the state machine, and returns whether it needs to be rerun
func (t *SM) Execute() (bool, error) {
	switch {
	case t.Task == nil:
		if err := t.startTask(); err != nil {
			return false, err
		}
	case t.PendingClose() && t.PastBuildBuffer() && t.Done.Read():
		t.clear()
		return true, nil
	case t.PendingClose() && !t.Done.Read():
		t.closeTask()
	}

	return false, nil
}

func (t *SM) startTask() error {
	var err error
	t.Task, err = New(t.c.BaseDir, t.c.Action)
	if err != nil {
		return err
	}

	go func() {
		t.Task.process.Wait()
		t.Done.Write(true)
		t.Reprocess <- struct{}{}
		t.ReloadEnded <- struct{}{}
	}()

	return nil
}

func (t *SM) closeTask() {
	if !t.PastBuildBuffer() {
		return
	}

	var err error
	if !t.PastKillTime() {
		err = t.Task.Interrupt()
	} else {
		err = t.Task.Kill()
	}
	if err != nil {
		log.L("Failed to stop task: %v", err)
	}
}

// clear resets the SM once a task has completed running.
func (t *SM) clear() {
	t.Task = nil
	t.Done.Write(false)
	t.reloadRequest = time.Time{}
}

// Reload will stop the task if it's running.
// To make sure the task is closed, a goroutine is set up to reprocess every second.
func (t *SM) Reload() {
	if t.PendingClose() {
		log.Fatalf("Reload called while already waiting for a close")
	}

	t.reloadRequest = time.Now()
	if !t.Running() {
		log.L("Change detected, starting task (task is no longer running)")
		go func() {
			time.Sleep(buildBufferTime)
			t.Reprocess <- struct{}{}
		}()
		return
	}

	log.L("Change detected, restarting task")
	go t.reloadCheck()
}

// Close will try interrupt the task, and if it does not close in 500ms, it will kill it.
func (t *SM) Close() {
	if !t.Running() {
		return
	}

	t.Task.Interrupt()
	select {
	case <-t.Reprocess:
		return
	case <-t.ReloadEnded:
		return
	case <-time.After(500 * time.Millisecond):
		t.Task.Kill()
	}
}

// reloadCheck is a goroutine that triggers a reprocess every second till the reload has completed.
func (t *SM) reloadCheck() {
	for {
		select {
		case <-t.ReloadEnded:
			return
		case <-time.After(time.Second):
			t.Reprocess <- struct{}{}
		}
	}
}
