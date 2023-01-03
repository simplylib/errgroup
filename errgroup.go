// Package errgroup implements a mostly drop in replacement for golang.org/x/sync/multierror
// with one major change, where golang.org/x/sync/errgroup returns the first non-nil error (if any)
// github.com/simplylib/errgroup returns all the errors using github.com/simplylib/multierror
package errgroup

import (
	"sync"

	"github.com/simplylib/multierror"
)

// Group of goroutines concurrently running tasks, optionally limited by SetLimit(int).
type Group struct {
	bucket chan struct{}

	wg sync.WaitGroup

	mu   sync.Mutex
	errs []error
}

// SetLimit of concurrently running goroutines to n, must be called before.
func (g *Group) SetLimit(n int) {
	g.bucket = make(chan struct{}, n)
}

// TryGo calls f in a separate goroutine, only if the goroutine limit set in SetLimit is not reached
// at moment of Calling TryGo. Acts like Group.Go if SetLimit was not run.
// Returns true if goroutine will/was started, false if not.
func (g *Group) TryGo(f func() error) bool {
	if g.bucket != nil {
		select {
		case g.bucket <- struct{}{}:
		default:
			return false
		}
	}

	g.wg.Add(1)
	// manually inline function for Go and TryGo
	go func() {
		defer g.wg.Done()

		if err := f(); err != nil {
			g.mu.Lock()
			g.errs = append(g.errs, err)
			g.mu.Unlock()
		}

		if g.bucket != nil {
			<-g.bucket
		}
	}()

	return true
}

// Go run a function in a separate Goroutine, unbounded by default unless SetLimit(int) is called.
func (g *Group) Go(f func() error) {
	if g.bucket != nil {
		g.bucket <- struct{}{}
	}

	g.wg.Add(1)
	go func() {
		defer g.wg.Done()

		if err := f(); err != nil {
			g.mu.Lock()
			g.errs = append(g.errs, err)
			g.mu.Unlock()
		}

		if g.bucket != nil {
			<-g.bucket
		}
	}()
}

// Wait until all goroutines are finished and returning a Multierror if len(errors) > 1 or the direct error if 1.
func (g *Group) Wait() error {
	g.wg.Wait()

	if len(g.errs) > 1 {
		return multierror.Errors(g.errs)
	}

	if len(g.errs) == 1 {
		return g.errs[0]
	}

	return nil
}
