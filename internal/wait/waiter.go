package wait

// Original source https://gist.github.com/maratori/010bfbf05639aa3a5ba832cdd75320ec

import (
	"sync"
	"time"

	"github.com/portworx/pds-integration-test/internal/tests"
)

const (
	ShortRetryInterval = time.Second * 1
	RetryInterval      = time.Second * 10

	// VeryLongTimeout is generally appropriate for highly resource-intensive operations,
	// which are expected to take a very long time to reconcile.
	VeryLongTimeout = time.Minute * 20
	// LongTimeout is generally appropriate for highly resource-intensive operations,
	// which are expected to take a long time to reconcile.
	LongTimeout = time.Minute * 10
	// StandardTimeout is the normal expected waiting period for I/O-bound resources.
	StandardTimeout = time.Minute * 5
	// ShortTimeout is intended for lightweight resource reconciliations.
	ShortTimeout = time.Minute * 1
	// QuickCheckTimeout is appropriate operations which aren't constrained by waiting
	// for resource provisioning (e.g. API-only operations, data roundtrips).
	QuickCheckTimeout = time.Second * 30
)

func For(t tests.T, timeout time.Duration, tick time.Duration, fn func(t tests.T)) {
	timer := time.NewTimer(timeout)
	ticker := time.NewTicker(tick)
	defer timer.Stop()
	defer ticker.Stop()
	for {
		ft := &fakeT{name: t.Name()}
		didPanic := false
		func() {
			defer func() {
				if recover() != nil {
					didPanic = true
				}
			}()
			fn(ft)
		}()
		if !ft.Failed() && !didPanic {
			return
		}

		select {
		case <-timer.C:
			fn(t)
			return
		case <-ticker.C:
		}
	}
}

type fakeT struct {
	sync.Mutex
	failed bool
	name   string
}

func (t *fakeT) fail() {
	t.Lock()
	defer t.Unlock()
	t.failed = true
}

func (t *fakeT) panic() {
	t.fail()
	panic("panic")
}

func (t *fakeT) Name() string {
	t.Lock()
	defer t.Unlock()
	return t.name
}

func (t *fakeT) Failed() bool {
	t.Lock()
	defer t.Unlock()
	return t.failed
}

func (t *fakeT) Error(_ ...any)            { t.fail() }
func (t *fakeT) Errorf(_ string, _ ...any) { t.fail() }
func (t *fakeT) Fail()                     { t.fail() }
func (t *fakeT) FailNow()                  { t.panic() }
func (t *fakeT) Fatal(_ ...any)            { t.panic() }
func (t *fakeT) Fatalf(_ string, _ ...any) { t.panic() }
func (t *fakeT) Log(_ ...any)              {}
func (t *fakeT) Logf(_ string, _ ...any)   {}
func (t *fakeT) Helper()                   {}
