// Package leaks checks that no way of ending a run leaves a goroutine behind.
//
// Go 1.27's goroutineleak profile finds goroutines blocked on something no
// runnable goroutine can ever unblock. The profile is process-wide, which is why
// this suite has a directory of its own: every other test binary's goroutines
// are invisible here, and nothing here runs in parallel.
package leaks

import (
	"bytes"
	"context"
	"errors"
	"regexp"
	"runtime/pprof"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hovsep/fmesh"
	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/internal/testutil"
	"github.com/hovsep/fmesh/plugin/profiler"
	"github.com/hovsep/fmesh/signal"
)

var leakHeader = regexp.MustCompile(`^goroutineleak profile: total (\d+)`)

// leakedGoroutines runs the leak detector and returns how many goroutines it
// found. Writing the profile is what triggers the detecting GC; debug=1 is the
// text form, whose first line carries the count.
func leakedGoroutines(t *testing.T) int {
	t.Helper()

	// A goroutine is only reported once it is parked; give any straggler the
	// chance to block before the detector looks.
	time.Sleep(10 * time.Millisecond)

	prof := pprof.Lookup("goroutineleak")
	require.NotNil(t, prof, "the goroutineleak profile is generally available since Go 1.27")

	var buf bytes.Buffer
	require.NoError(t, prof.WriteTo(&buf, 1))

	m := leakHeader.FindSubmatch(buf.Bytes())
	require.NotNil(t, m, "unexpected profile header:\n%s", buf.String())
	n, err := strconv.Atoi(string(m[1]))
	require.NoError(t, err)
	if n > 0 {
		t.Log(buf.String())
	}
	return n
}

// loopingMesh is a component feeding itself, so the mesh runs until something
// stops it; work runs on each activation and may end the run.
func loopingMesh(t *testing.T, work func(ctx context.Context) error, opts ...fmesh.Option) *fmesh.FMesh {
	t.Helper()

	c := testutil.MustComponent("looper",
		component.WithInputs("in"),
		component.WithOutputs("out"),
		component.WithActivationFunc(func(ctx context.Context, this *component.Component) error {
			if work != nil {
				if err := work(ctx); err != nil {
					return err
				}
			}
			return this.OutputByName("out").PutSignals(signal.New(1))
		}))
	require.NoError(t, c.LoopbackPipe("out", "in"))

	fm := testutil.MustFMesh("leak-mesh", opts...)
	require.NoError(t, fm.AddComponents(c))
	require.NoError(t, c.InputByName("in").PutSignals(signal.New(1)))
	return fm
}

func TestRun_LeavesNoGoroutinesBehind(t *testing.T) {
	errBoom := errors.New("boom")

	scenarios := []struct {
		name string
		run  func(t *testing.T) error
	}{
		{"natural stop", func(t *testing.T) error {
			fm := loopingMesh(t, nil, fmesh.WithCyclesLimit(5))
			_, err := fm.Run(context.Background())
			return err
		}},
		{"canceled context", func(t *testing.T) error {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			cycles := 0
			fm := loopingMesh(t, func(context.Context) error {
				cycles++
				if cycles >= 3 {
					cancel()
				}
				return nil
			}, fmesh.WithUnlimitedCycles())
			_, err := fm.Run(ctx)
			return err
		}},
		{"time limit", func(t *testing.T) error {
			fm := loopingMesh(t, func(ctx context.Context) error {
				<-ctx.Done() // block until the mesh's own deadline interrupts
				return ctx.Err()
			}, fmesh.WithTimeLimit(10*time.Millisecond), fmesh.WithUnlimitedCycles())
			_, err := fm.Run(context.Background())
			return err
		}},
		{"panicking component", func(t *testing.T) error {
			fm := loopingMesh(t, func(context.Context) error { panic("kaboom") })
			_, err := fm.Run(context.Background())
			return err
		}},
		{"error with StopOnFirstErrorOrPanic", func(t *testing.T) error {
			fm := loopingMesh(t, func(context.Context) error { return errBoom },
				fmesh.WithErrorHandlingStrategy(fmesh.StopOnFirstErrorOrPanic))
			_, err := fm.Run(context.Background())
			return err
		}},
		{"error with StopOnFirstPanic", func(t *testing.T) error {
			fm := loopingMesh(t, func(context.Context) error { return errBoom },
				fmesh.WithErrorHandlingStrategy(fmesh.StopOnFirstPanic), fmesh.WithCyclesLimit(5))
			_, err := fm.Run(context.Background())
			return err
		}},
		{"error with IgnoreAll", func(t *testing.T) error {
			fm := loopingMesh(t, func(context.Context) error { return errBoom },
				fmesh.WithErrorHandlingStrategy(fmesh.IgnoreAll), fmesh.WithCyclesLimit(5))
			_, err := fm.Run(context.Background())
			return err
		}},
		{"livelock", func(t *testing.T) error {
			waiter := func(name string) *component.Component {
				return testutil.MustComponent(name,
					component.WithInputs("in"),
					component.WithOutputs("out"),
					component.WithActivationFunc(func(context.Context, *component.Component) error {
						return component.ErrWaitKeepingInputs
					}))
			}
			a, b := waiter("alpha"), waiter("bravo")
			fm := testutil.MustFMesh("mutual-wait")
			require.NoError(t, fm.AddComponents(a, b))
			require.NoError(t, a.OutputByName("out").PipeTo(b.InputByName("in")))
			require.NoError(t, b.OutputByName("out").PipeTo(a.InputByName("in")))
			require.NoError(t, a.InputByName("in").PutSignals(signal.New(1)))
			_, err := fm.Run(context.Background())
			return err
		}},
		{"profiled mesh", func(t *testing.T) error {
			fm := loopingMesh(t, nil, fmesh.WithCyclesLimit(5),
				fmesh.WithPlugins(profiler.New(profiler.ModeAll)))
			_, err := fm.Run(context.Background())
			return err
		}},
	}

	for _, sc := range scenarios {
		t.Run(sc.name, func(t *testing.T) {
			err := sc.run(t)
			t.Logf("run ended with: %v", err)
			assert.Zero(t, leakedGoroutines(t), "a finished run must not leave goroutines behind")
		})
	}
}
