package sync

import "sync"

// Bool is a boolean that can be used by multiple goroutines.
type Bool struct {
	val  bool
	lock sync.RWMutex
}

// NewBool returns a new Bool with the given default value.
func NewBool(val bool) *Bool {
	return &Bool{val: val}
}

// Read returns the current value of the bool.
func (b *Bool) Read() bool {
	b.lock.RLock()
	defer b.lock.RUnlock()
	return b.val
}

// Write writes the given newVal to the bool.
func (b *Bool) Write(newVal bool) {
	b.lock.Lock()
	defer b.lock.Unlock()
	b.val = newVal
}
