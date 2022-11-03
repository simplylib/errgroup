// package errgroup implements a drop in replacement for golang.org/x/sync/multierror
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

// Go run a function, unbounded by default unless SetLimit(int) is called.
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
