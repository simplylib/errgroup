package errgroup

import (
	"context"
	"errors"
	"io"
	"runtime"
	"sync"
	"testing"
)

func TestWithContextTryGo(t *testing.T) {
	t.Parallel()

	eg, ctx := WithContext(context.Background())
	eg.SetLimit(1)

	ok := eg.TryGo(func() error {
		return io.EOF
	})
	if !ok {
		t.Fatal("expected ok = true instead ok = false")
	}

	err := eg.Wait()
	if err != io.EOF {
		t.Fatalf("expected io.EOF instead got (%T)", err)
	}

	select {
	case <-ctx.Done():
	default:
		t.Fatal("expected cancelled context")
	}
}

func TestWithContextGo(t *testing.T) {
	t.Parallel()

	eg, ctx := WithContext(context.Background())
	eg.SetLimit(1)

	eg.Go(func() error {
		return io.EOF
	})

	err := eg.Wait()
	if err != io.EOF {
		t.Fatalf("expected io.EOF instead got (%T)", err)
	}

	select {
	case <-ctx.Done():
	default:
		t.Fatal("expected cancelled context")
	}
}

func TestNoError(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

	eg := Group{}
	eg.SetLimit(1)

	stop := make(chan struct{})
	i := 0

	eg.Go(func() error {
		<-stop
		i++
		return nil
	})

	ok := eg.TryGo(func() error {
		<-stop
		i++
		return nil
	})
	if ok {
		t.Fatalf("expected ok = false instead ok = true")
	}

	close(stop)

	if err := eg.Wait(); err != nil {
		t.Fatal(err)
	}

	if i != 1 {
		t.Fatalf("expect i = 1 intead i = %v", i)
	}
}

func TestTryGo(t *testing.T) {
	t.Parallel()

	eg := Group{}
	eg.SetLimit(runtime.NumCPU())

	closeChan := make(chan struct{})

	var run bool
	for i := 0; i < runtime.NumCPU(); i++ {
		run = eg.TryGo(func() error {
			<-closeChan
			return io.EOF
		})

		if !run {
			close(closeChan)
			t.Fatal("could not run all goroutines")
		}
	}

	if eg.TryGo(func() error { return nil }) {
		t.Fatal("TryGo returned true when limit is currently hit")
	}

	close(closeChan)

	err := eg.Wait()
	multierror, ok := err.(interface{ Unwrap() []error })
	if !ok {
		t.Fatalf("Expected a multierror got (%t)\n", multierror)
	}

	me := multierror.Unwrap()

	if len(me) != runtime.NumCPU() {
		t.Fatalf("expected (%v) errors got (%v)\n", runtime.NumCPU(), len(me))
	}

	for i := range me {
		if me[i] != io.EOF {
			t.Fatalf("expected (io.EOF) got (%v)\n", err)
		}
	}
}

func TestSingleError(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

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

	me, ok := err.(interface{ Unwrap() []error })
	if !ok {
		t.Fatalf("expected a multierror.Errors got (%T)\n", me)
	}

	errs := me.Unwrap()

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
