package spinlog

import (
	"io"
	"sync"
)

// Router is an io.Writer whose output stream can be modified dynamically.
type Router struct {
	mutex  sync.Mutex
	target *io.Writer
}

// NewRouter creates a new router with no target.
func NewRouter() *Router {
	return &Router{sync.Mutex{}, nil}
}

// GetTarget returns the router's current target.
func (r *Router) GetTarget() *io.Writer {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	return r.target
}

// SetTarget sets the router's current target.
func (r *Router) SetTarget(writer io.Writer) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.target = &writer
}

// Write writes the bytes to the router's current target.
// If the router has no target, (len(p), nil) is returned.
func (r *Router) Write(p []byte) (n int, err error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if r.target == nil {
		return len(p), nil
	}
	return (*r.target).Write(p)
}
