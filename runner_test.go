package pararun_test

import (
	"errors"
	"sync/atomic"
	"testing"

	"github.com/KarpelesLab/pararun"
)

func TestForeach(t *testing.T) {
	// run some tests with foreach
	runner1 := pararun.New(1)
	runner5 := pararun.New(5)
	defer runner1.Stop()
	defer runner5.Stop()

	var v uint32
	x := make([]uint32, 64)
	e := errors.New("test error")

	err := pararun.ForEach(runner1, x, func(n int, n2 uint32) error {
		if n > 5 {
			return e
		}
		atomic.AddUint32(&v, 1)
		return nil
	})

	if err != e {
		t.Errorf("failed foreach cancel: got error %v instead of test error", err)
	}
	if v != 6 {
		t.Errorf("expected v=6, got v=%d", v)
	}

	v = 0

	err = pararun.ForEach(runner5, x, func(n int, n2 uint32) error {
		atomic.AddUint32(&v, uint32(n))
		return nil
	})

	if v != 2016 {
		t.Errorf("expected v=2016, got v=%d", v)
	}
	if err != nil {
		t.Errorf("failed to run foreach: %s", err)
	}
}
