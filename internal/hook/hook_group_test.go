package hook

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Registration order is the mesh's ordering guarantee: plugins register hooks
// during Init, so the order two plugins observe the same event depends on it.
func TestGroup_TriggersInRegistrationOrder(t *testing.T) {
	var log []int
	hg := NewGroup[int]()
	for _, factor := range []int{1, 2, 3} {
		hg.Add(func(_ context.Context, i int) error {
			log = append(log, i*factor)
			return nil
		})
	}

	require.NoError(t, hg.Trigger(context.Background(), 10))
	assert.Equal(t, []int{10, 20, 30}, log)
}

// Trigger is fail-fast: the first error stops the remaining hooks. Callers rely
// on this to abort a run rather than continue past a failed hook.
func TestGroup_TriggerStopsOnFirstError(t *testing.T) {
	boom := errors.New("boom")
	ran := 0

	hg := NewGroup[int]()
	hg.Add(func(_ context.Context, _ int) error { ran++; return nil })
	hg.Add(func(_ context.Context, _ int) error { ran++; return boom })
	hg.Add(func(_ context.Context, _ int) error { ran++; return nil })

	err := hg.Trigger(context.Background(), 1)
	require.ErrorIs(t, err, boom)
	assert.Equal(t, 2, ran, "the hook after the failing one must not run")
}

func TestGroup_TriggerWithNoHooks(t *testing.T) {
	require.NoError(t, NewGroup[int]().Trigger(context.Background(), 42))
}

func TestGroup_All(t *testing.T) {
	hg := NewGroup[int]()
	hg.Add(func(_ context.Context, _ int) error { return nil })
	hg.Add(func(_ context.Context, _ int) error { return nil })

	assert.Len(t, hg.All(), 2)
}
