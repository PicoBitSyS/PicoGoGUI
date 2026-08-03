// Package binding provides reactive value wrappers for two-way UI sync.
//
// Prefer Var[T] over plain struct fields. Assigning to a plain field
// (e.g. settings.Server = "x") cannot notify the UI in idiomatic Go without
// code generation.
package binding

import (
	"sync"
	"sync/atomic"
)

// Origin identifies the producer of a bound value change.
//
// A subscriber registered with the same non-zero origin is skipped when
// SetFrom is used. This prevents a control from echoing its own UI change
// back to itself while still notifying every other control and subscriber.
type Origin uint64

var originSeq atomic.Uint64

// NewOrigin returns a process-unique binding origin.
func NewOrigin() Origin {
	return Origin(originSeq.Add(1))
}

type subscriber[T any] struct {
	origin Origin
	fn     func(T)
}

// Var is a thread-safe observable value container.
type Var[T any] struct {
	mu    sync.RWMutex
	value T
	subs  []*subscriber[T]
}

// New creates a Var with an initial value.
//
// Example:
//
//	host := binding.New("localhost")
func New[T any](v T) *Var[T] {
	return &Var[T]{value: v}
}

// Get returns the current value.
func (v *Var[T]) Get() T {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.value
}

// Set updates the value and notifies subscribers.
func (v *Var[T]) Set(val T) {
	v.SetFrom(0, val)
}

// SetFrom updates the value and notifies subscribers other than the one
// registered with origin. Use a non-zero origin for changes produced by a
// bound control; use Set for ordinary model changes.
func (v *Var[T]) SetFrom(origin Origin, val T) {
	v.mu.Lock()
	v.value = val
	subs := append([]*subscriber[T](nil), v.subs...)
	v.mu.Unlock()
	for _, s := range subs {
		if origin != 0 && s.origin == origin {
			continue
		}
		s.fn(val)
	}
}

// SetSilent updates the value without notifying subscribers.
// Use when applying a UI change that should not echo back as a patch.
func (v *Var[T]) SetSilent(val T) {
	v.mu.Lock()
	v.value = val
	v.mu.Unlock()
}

// Subscribe registers a change listener. The listener is not called immediately.
// Returns an unsubscribe function.
func (v *Var[T]) Subscribe(fn func(T)) (unsubscribe func()) {
	return v.SubscribeFrom(0, fn)
}

// SubscribeFrom registers a listener associated with an origin.
// Changes sent through SetFrom with the same non-zero origin skip this
// listener while continuing to notify all other listeners.
func (v *Var[T]) SubscribeFrom(origin Origin, fn func(T)) (unsubscribe func()) {
	if fn == nil {
		return func() {}
	}
	sub := &subscriber[T]{origin: origin, fn: fn}
	v.mu.Lock()
	v.subs = append(v.subs, sub)
	v.mu.Unlock()
	return func() {
		v.mu.Lock()
		defer v.mu.Unlock()
		out := v.subs[:0]
		for _, s := range v.subs {
			if s != sub {
				out = append(out, s)
			}
		}
		v.subs = out
	}
}
