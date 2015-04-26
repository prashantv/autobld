package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/fsnotify.v1"
)

func setupListener(c *Config, m *Matcher, watcher *fsnotify.Watcher) error {
	if len(m.Dirs) == 0 {
		m.Dirs = []string{""}
	}

	for _, d := range m.Dirs {
		dir := c.BaseDir + "/" + d
		if err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return fmt.Errorf("Walk directories failed: %v", err)
			}
			if m.excludeDirMap[filepath.Base(path)] {
				return filepath.SkipDir
			}
			if info.IsDir() {
				c.configsMap[filepath.Clean(path)] = m
				watcher.Add(path)
			}
			return nil
		}); err != nil {
			return err
		}
	}
	return nil
}

// wrapErr wraps some known errors with more information.
func wrapErr(err error) error {
	eMsg := err.Error()
	if strings.Contains(eMsg, "too many open files") {
		return fmt.Errorf("%v\nTo increase the limit for files that can be watched no OSX, run ulimit -n 512. The default limit is 256", err)
	}
	return err
}

// SetupWatcher sets up fsnotify to watch all the directories specified in the config.
func SetupWatcher(c *Config) (*fsnotify.Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, wrapErr(err)
	}

	for _, m := range c.Matchers {
		if err := setupListener(c, &m, watcher); err != nil {
			return nil, wrapErr(err)
		}
	}
	return watcher, nil
}

// IsMatch checks whether an event on the given path should cause a reload.
func IsMatch(c *Config, path string) bool {
	dir, file := filepath.Split(path)
	if len(dir) == 0 {
		dir = "./"
	}
	//	log.Printf("check for match %v in %+v", path, c.configsMap)
	dc := c.configsMap[filepath.Clean(dir)]
	if dc == nil {
		return false
	}

	// If there are no patterns, then we treat it as a wildcard matching everything.
	if len(dc.Patterns) == 0 {
		return true
	}
	for _, p := range dc.Patterns {
		if match, err := filepath.Match(p, file); err == nil && match {
			return true
		}
	}
	return false
}
