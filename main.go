package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/prashantv/autobld/config"
	"github.com/prashantv/autobld/log"
	"github.com/prashantv/autobld/proxy"
	"github.com/prashantv/autobld/task"

	"gopkg.in/fsnotify.v1"
)

func main() {
	c, err := config.Parse()
	if err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	watcher, err := config.SetupWatcher(c)
	if err != nil {
		log.Fatalf("Change detection failed: %v", err)
	}

	var (
		// errC is used to report errors. Any error will cause a log.Fatal
		errC = make(chan error)
		// signnalC is used for signals to this process (ctrl+c, etc).
		signalC = make(chan os.Signal)
	)
	signal.Notify(signalC, syscall.SIGINT, syscall.SIGKILL)

	// Start all the proxy listeners.
	for _, pc := range c.ProxyConfigs {
		go proxy.Start(pc, errC)
	}

	if err := eventLoop(c, errC, signalC, watcher); err != nil {
		log.Fatalf("Error: %v", err)
	}
}

func eventLoop(c *config.Config, errC <-chan error, signalC <-chan os.Signal, watcher *fsnotify.Watcher) error {
	taskSM := task.NewSM(c)
	defer taskSM.Close()

	for {
		// Any change in the task state should make the proxy try reconnecting.
		proxy.TryConnect.Write(true)

		// Run the task state machine.
		for rerun := true; rerun; {
			var err error
			rerun, err = taskSM.Execute()
			if err != nil {
				return fmt.Errorf("task error: %v", err)
			}
		}

		select {
		case err := <-errC:
			return err
		case err := <-watcher.Errors:
			return fmt.Errorf("watcher error: %v", err)
		case <-signalC:
			return nil
		case event := <-watcher.Events:
			if config.IsMatch(c, event.Name) && !taskSM.PendingClose() {
				taskSM.Reload()
			}
		case <-taskSM.Reprocess:
			// Nothing needs to be done, just the standard reprocess.
		}
	}
	return nil
}
