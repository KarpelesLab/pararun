package pararun

import (
	"context"
	"runtime"
	"sync"
)

type Runner struct {
	ch chan func()
}

var MaxThreadCount = 256 // overall maximum number of threads

// New returns a new runner limited to the specified number of threads
func New(cnt int) *Runner {
	if cnt <= 0 {
		cnt = runtime.NumCPU() + 1
	}
	if cnt > MaxThreadCount {
		cnt = MaxThreadCount
	}

	r := &Runner{
		ch: make(chan func()),
	}

	for i := 0; i < cnt; i += 1 {
		go r.thread()
	}

	return r
}

// Run runs the specified function. If no thread is available, it will wait for one to become available
func (r *Runner) Run(f func()) {
	r.ch <- f
}

// TryRun attempts to run the specified function, but if no thread is available it will return false
func (r *Runner) TryRun(f func()) bool {
	select {
	case r.ch <- f:
		return true
	default:
		return false
	}
}

// RunCtx will run the function f, but if no thread is available it'll wait until the function can run, or
// until the context is cancelled.
func (r *Runner) RunCtx(ctx context.Context, f func()) error {
	select {
	case r.ch <- f:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// ForEach allows going through a given data array using the specified runner
func ForEach[T any](r *Runner, data []T, cb func(int, T) error) error {
	wg := &sync.WaitGroup{}
	wg.Add(len(data))
	ctx, cancel := context.WithCancelCause(context.Background())

	// if waitgroup completes, set cancel
	go func() {
		wg.Wait()
		cancel(nil)
	}()

	go func() {
		for n, v := range data {
			f := func() {
				defer wg.Done()
				if err := cb(n, v); err != nil {
					cancel(err)
				}
			}
			select {
			case r.ch <- f:
				// all good
			case <-ctx.Done():
				// empty the waitgroup so the waitgroup waiting routing also ends
				cnt := len(data) - n
				if cnt > 0 {
					wg.Add(-cnt)
				}
				return
			}
		}
	}()

	// wait for completion
	<-ctx.Done()
	return context.Cause(ctx)
}

func (r *Runner) thread() {
	for f := range r.ch {
		f()
	}
}

// Stop will disable this runner and stop all its threads. Any call to Run after Stop will panic
func (r *Runner) Stop() {
	if v := r.ch; v != nil {
		r.ch = nil
		close(v)
	}
}
