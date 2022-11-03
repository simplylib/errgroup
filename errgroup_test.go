package errgroup

import (
	"errors"
	"io"
	"runtime"
	"sync"
	"testing"

	"github.com/simplylib/multierror"
)

func TestGroupNoError(t *testing.T) {
	eg := Group{}

	var (
		count int
		mu    sync.Mutex
	)

	const countTarget = 10000

	for i := 0; i < countTarget; i++ {
		eg.Go(func() error {
			mu.Lock()
			count++
			mu.Unlock()
			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		t.Fatal(err)
	}

	if count != countTarget {
		t.Fatalf("count (%v) != countTarget (%v)\n", count, countTarget)
	}
}

func TestSetLimit(t *testing.T) {
	eg := Group{}

	eg.SetLimit(runtime.NumCPU())

	var (
		count int
		mu    sync.Mutex
	)

	const countTarget = 10000

	for i := 0; i < countTarget; i++ {
		eg.Go(func() error {
			mu.Lock()
			count++
			mu.Unlock()
			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		t.Fatal(err)
	}

	if count != countTarget {
		t.Fatalf("count (%v) != countTarget (%v)\n", count, countTarget)
	}
}
func TestGroupSingleError(t *testing.T) {
	eg := Group{}

	eg.SetLimit(runtime.NumCPU())

	const countTarget = 10000

	eg.Go(func() error {
		return io.EOF
	})

	for i := 0; i < countTarget-1; i++ {
		eg.Go(func() error {
			return nil
		})
	}

	err := eg.Wait()

	if err == nil {
		t.Fatal("expected err != nil")
	}

	if err != io.EOF {
		t.Fatalf("expected io.EOF got (%T)\n", err)
	}

}
func TestGroupMultiError(t *testing.T) {
	eg := Group{}

	eg.SetLimit(runtime.NumCPU())

	const countTarget = 10000

	for i := 0; i < countTarget/2; i++ {
		eg.Go(func() error {
			return io.EOF
		})
	}

	for i := 0; i < countTarget/2; i++ {
		eg.Go(func() error {
			return io.ErrClosedPipe
		})
	}

	err := eg.Wait()
	if err == nil {
		t.Fatal("expected err != nil")
	}

	me, ok := err.(multierror.Errors)
	if !ok {
		t.Fatalf("expected a multierror.Errors got (%T)\n", me)
	}

	errs := []error(me)

	if len(errs) != countTarget {
		t.Fatalf("len(errs) = (%v) expected (%v)\n", len(errs), countTarget)
	}

	for i := range errs {
		if errors.Is(errs[i], io.EOF) {
			continue
		}

		if errors.Is(errs[i], io.ErrClosedPipe) {
			continue
		}

		t.Fatalf("expected io.EOF or io.ErrClosedPipe got (%T)\n", errs[i])
	}
}
