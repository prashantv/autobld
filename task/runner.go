package task

import (
	"sync"
	"time"

	"github.com/prashantv/autobld/config"
	"github.com/prashantv/autobld/log"
	"github.com/prashantv/autobld/syncv"
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
	// done is set to True once a task ends.
	done syncv.Bool
	// blockRequests is used to block all proxy port requests after a Reload is requested.
	blockRequests *sync.WaitGroup

	// Reprocess is the channel the caller waits on to reprocess the state machine.
	Reprocess chan struct{}
	// reloadEnded is written to when a Reload is complete (eg the task has ended).
	reloadEnded chan struct{}
}

// NewSM returns the state maachine used to run tasks.
func NewSM(c *config.Config, blockRequests *sync.WaitGroup) *SM {
	return &SM{
		c:             c,
		blockRequests: blockRequests,
		Reprocess:     make(chan struct{}),
		reloadEnded:   make(chan struct{}),
	}
}

// Running returns whether the task is currently running.
func (t *SM) Running() bool {
	return t.Task != nil && !t.done.Read()
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
	case t.PendingClose() && t.PastBuildBuffer():
		if !t.done.Read() {
			t.closeTask()
			return false, nil
		}

		t.clear()
		return true, nil
	}

	return false, nil
}

func (t *SM) startTask() error {
	var err error
	t.Task, err = New(t.c.BaseDir, t.c.Action)
	if err != nil {
		return err
	}

	t.blockRequests.Done()

	go func() {
		t.Task.process.Wait()
		log.V("Task is no longer running")
		t.done.Write(true)
		t.Reprocess <- struct{}{}
		t.reloadEnded <- struct{}{}
	}()

	return nil
}

func (t *SM) closeTask() {
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
	t.done.Write(false)
	t.reloadRequest = time.Time{}
}

// Reload will stop the task if it's running.
// To make sure the task is closed, a goroutine is set up to reprocess every second.
func (t *SM) Reload() {
	if t.PendingClose() {
		log.Fatalf("Reload called while already waiting for a close")
	}

	t.blockRequests.Add(1)
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
	case <-t.reloadEnded:
		return
	case <-time.After(500 * time.Millisecond):
		t.Task.Kill()
	}
}

// reloadCheck is a goroutine that triggers a reprocess every second till the reload has completed.
func (t *SM) reloadCheck() {
	for {
		select {
		case <-t.reloadEnded:
			return
		case <-time.After(time.Second):
			t.Reprocess <- struct{}{}
		}
	}
}
