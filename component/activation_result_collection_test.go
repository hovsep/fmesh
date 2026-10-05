package component

import (
	"errors"
	"fmt"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestActivationResultCollection_Add(t *testing.T) {
	type args struct {
		activationResults []*ActivationResult
	}
	tests := []struct {
		name       string
		collection *ActivationResultCollection
		args       args
		assertions func(t *testing.T, collection *ActivationResultCollection)
	}{
		{
			name:       "adding nothing to empty collection",
			collection: NewActivationResultCollection(),
			args: args{
				activationResults: nil,
			},
			assertions: func(t *testing.T, collection *ActivationResultCollection) {
				assert.Zero(t, collection.Len())
				assert.False(t, collection.HasActivationErrors())
				assert.False(t, collection.HasActivationPanics())
				assert.False(t, collection.HasActivatedComponents())
			},
		},
		{
			name:       "adding to empty collection",
			collection: NewActivationResultCollection(),
			args: args{
				activationResults: []*ActivationResult{
					mustNew("c1").newResult(ActivationCodeOK),
					mustNew("c2").newResult(ActivationCodeReturnedError, errors.New("oops")),
				},
			},
			assertions: func(t *testing.T, collection *ActivationResultCollection) {
				assert.Equal(t, 2, collection.Len())
				assert.True(t, collection.HasActivatedComponents())
				assert.True(t, collection.HasActivationErrors())
				assert.False(t, collection.HasActivationPanics())
			},
		},
		{
			name: "adding to non-empty collection",
			collection: NewActivationResultCollection().Add(
				mustNew("c1").newResult(ActivationCodeOK),
				mustNew("c2").newResult(ActivationCodeOK),
			),
			args: args{
				activationResults: []*ActivationResult{
					mustNew("c4").newResult(ActivationCodeNoInput),
					mustNew("c5").newResult(ActivationCodePanicked, errors.New("panic")),
				},
			},
			assertions: func(t *testing.T, collection *ActivationResultCollection) {
				assert.Equal(t, 4, collection.Len())
				assert.True(t, collection.HasActivationPanics())
				assert.False(t, collection.HasActivationErrors())
				assert.True(t, collection.HasActivatedComponents())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.collection.Add(tt.args.activationResults...)
			if tt.assertions != nil {
				tt.assertions(t, tt.collection)
			}
		})
	}
}

func TestActivationResultCollection_ByName(t *testing.T) {
	r1 := NewActivationResult("c1").SetActivated(true)
	r2 := NewActivationResult("c2").SetActivated(false)
	collection := NewActivationResultCollection().Add(r1, r2)

	t.Run("existing result", func(t *testing.T) {
		result := collection.ByName("c1")
		assert.NotNil(t, result)
		assert.Equal(t, "c1", result.ComponentName())
		assert.True(t, result.Activated())
	})

	t.Run("non-existing result", func(t *testing.T) {
		result := collection.ByName("c3")
		assert.Nil(t, result)
	})
}

func TestActivationResultCollection_All(t *testing.T) {
	r1 := NewActivationResult("c1").SetActivated(true)
	r2 := NewActivationResult("c2").SetActivated(false)

	t.Run("returns all results", func(t *testing.T) {
		collection := NewActivationResultCollection().Add(r1, r2)
		all := collection.All()
		assert.Len(t, all, 2)
		assert.Contains(t, all, "c1")
		assert.Contains(t, all, "c2")
	})

	t.Run("empty collection", func(t *testing.T) {
		collection := NewActivationResultCollection()
		all := collection.All()
		assert.Empty(t, all)
	})
}

func TestActivationResultCollection_AllOrdered(t *testing.T) {
	c := NewActivationResultCollection()
	c.Add(NewActivationResult("charlie"), NewActivationResult("alpha"), NewActivationResult("bravo"))

	ordered := c.AllOrdered()
	require.Len(t, ordered, 3)
	assert.Equal(t, "alpha", ordered[0].ComponentName())
	assert.Equal(t, "bravo", ordered[1].ComponentName())
	assert.Equal(t, "charlie", ordered[2].ComponentName())

	assert.Empty(t, NewActivationResultCollection().AllOrdered())
}

func TestActivationResultCollection_IsEmpty(t *testing.T) {
	t.Run("empty collection", func(t *testing.T) {
		collection := NewActivationResultCollection()
		assert.True(t, collection.IsEmpty())
	})

	t.Run("non-empty collection", func(t *testing.T) {
		collection := NewActivationResultCollection().Add(NewActivationResult("c1"))
		assert.False(t, collection.IsEmpty())
	})
}

func TestActivationResultCollection_Every(t *testing.T) {
	r1 := NewActivationResult("c1").SetActivated(true)
	r2 := NewActivationResult("c2").SetActivated(true)
	r3 := NewActivationResult("c3").SetActivated(false)

	t.Run("all match", func(t *testing.T) {
		collection := NewActivationResultCollection().Add(r1, r2)
		result := collection.Every(func(r *ActivationResult) bool {
			return r.Activated()
		})
		assert.True(t, result)
	})

	t.Run("not all match", func(t *testing.T) {
		collection := NewActivationResultCollection().Add(r1, r3)
		result := collection.Every(func(r *ActivationResult) bool {
			return r.Activated()
		})
		assert.False(t, result)
	})

	t.Run("empty collection returns true", func(t *testing.T) {
		collection := NewActivationResultCollection()
		result := collection.Every(func(r *ActivationResult) bool {
			return false
		})
		assert.True(t, result)
	})
}

func TestActivationResultCollection_Any(t *testing.T) {
	r1 := NewActivationResult("c1").SetActivated(true)
	r2 := NewActivationResult("c2").SetActivated(false)

	t.Run("at least one matches", func(t *testing.T) {
		collection := NewActivationResultCollection().Add(r1, r2)
		result := collection.Any(func(r *ActivationResult) bool {
			return r.Activated()
		})
		assert.True(t, result)
	})

	t.Run("none match", func(t *testing.T) {
		collection := NewActivationResultCollection().Add(r2)
		result := collection.Any(func(r *ActivationResult) bool {
			return r.Activated()
		})
		assert.False(t, result)
	})

	t.Run("empty collection returns false", func(t *testing.T) {
		collection := NewActivationResultCollection()
		result := collection.Any(func(r *ActivationResult) bool {
			return true
		})
		assert.False(t, result)
	})
}

func TestActivationResultCollection_Count(t *testing.T) {
	r1 := NewActivationResult("c1").SetActivated(true)
	r2 := NewActivationResult("c2").SetActivated(false)
	r3 := NewActivationResult("c3").SetActivated(true)

	t.Run("counts matching results", func(t *testing.T) {
		collection := NewActivationResultCollection().Add(r1, r2, r3)
		count := collection.Count(func(r *ActivationResult) bool {
			return r.Activated()
		})
		assert.Equal(t, 2, count)
	})

	t.Run("no matches", func(t *testing.T) {
		collection := NewActivationResultCollection().Add(r2)
		count := collection.Count(func(r *ActivationResult) bool {
			return r.Activated()
		})
		assert.Equal(t, 0, count)
	})
}

func TestActivationResultCollection_ForEach(t *testing.T) {
	r1 := NewActivationResult("c1")
	r2 := NewActivationResult("c2")
	r3 := NewActivationResult("c3")

	t.Run("applies action to all results", func(t *testing.T) {
		collection := NewActivationResultCollection().Add(r1, r2, r3)
		count := 0
		require.NoError(t, collection.ForEach(func(r *ActivationResult) error {
			count++
			return nil
		}))
		assert.Equal(t, 3, count)
	})

	t.Run("empty collection", func(t *testing.T) {
		collection := NewActivationResultCollection()
		count := 0
		require.NoError(t, collection.ForEach(func(r *ActivationResult) error {
			count++
			return nil
		}))
		assert.Equal(t, 0, count)
	})

	t.Run("visits results in component-name order", func(t *testing.T) {
		// Ranging over the map visited results in a different order on every call.
		collection, names := newNamedResultCollection()
		var visited []string
		require.NoError(t, collection.ForEach(func(r *ActivationResult) error {
			visited = append(visited, r.ComponentName())
			return nil
		}))
		assert.Equal(t, names, visited)
	})

	t.Run("action may change the collection", func(t *testing.T) {
		// The action ran under the read lock, so Remove/Add/Clear from it deadlocked.
		collection, _ := newNamedResultCollection()
		done := make(chan error, 1)
		go func() {
			done <- collection.ForEach(func(r *ActivationResult) error {
				collection.Remove(r.ComponentName())
				collection.Add(NewActivationResult("added-" + r.ComponentName()))
				collection.Clear()
				return nil
			})
		}()
		select {
		case err := <-done:
			require.NoError(t, err)
		case <-time.After(5 * time.Second):
			t.Fatal("ForEach deadlocked when the action changed the collection")
		}
	})
}

// newNamedResultCollection adds enough results that map order is visibly not name order.
func newNamedResultCollection() (collection *ActivationResultCollection, names []string) {
	collection = NewActivationResultCollection()
	names = make([]string, 0, 32)
	for i := range 32 {
		names = append(names, fmt.Sprintf("c%02d", i))
	}
	for _, name := range slices.Backward(names) {
		collection.Add(NewActivationResult(name))
	}
	return collection, names
}

func TestActivationResultCollection_Clear(t *testing.T) {
	r1 := NewActivationResult("c1")
	r2 := NewActivationResult("c2")

	t.Run("clears all results", func(t *testing.T) {
		collection := NewActivationResultCollection().Add(r1, r2)
		assert.Equal(t, 2, collection.Len())
		collection.Clear()
		assert.Equal(t, 0, collection.Len())
		assert.True(t, collection.IsEmpty())
	})
}

func TestActivationResultCollection_Remove(t *testing.T) {
	r1 := NewActivationResult("c1").SetActivated(true)
	r2 := NewActivationResult("c2").SetActivated(false)
	r3 := NewActivationResult("c3").SetActivated(true)

	t.Run("removes by component name", func(t *testing.T) {
		collection := NewActivationResultCollection().Add(r1, r2, r3)
		result := collection.Remove("c2")
		assert.Equal(t, 2, result.Len())
	})

	t.Run("removes multiple", func(t *testing.T) {
		collection := NewActivationResultCollection().Add(r1, r2, r3)
		result := collection.Remove("c1", "c2")
		assert.Equal(t, 1, result.Len())
	})

	t.Run("removes all", func(t *testing.T) {
		collection := NewActivationResultCollection().Add(r1, r2)
		result := collection.Remove("c1", "c2")
		assert.Equal(t, 0, result.Len())
	})
}

func TestActivationResult_ActivationErrorWithComponentName(t *testing.T) {
	err := errors.New("activation failed")
	r := NewActivationResult("my-component").AddActivationError(err)

	t.Run("returns error with component name", func(t *testing.T) {
		wrappedErr := r.ActivationErrorWithComponentName()
		require.Error(t, wrappedErr)
		assert.Contains(t, wrappedErr.Error(), "my-component")
		assert.Contains(t, wrappedErr.Error(), "activation failed")
	})

	t.Run("keeps the original error in the chain", func(t *testing.T) {
		assert.ErrorIs(t, r.ActivationErrorWithComponentName(), err)
	})

	t.Run("nil without activation errors", func(t *testing.T) {
		// Wrapping a nil error rendered "%!w(<nil>)" and turned success into a non-nil error.
		r := NewActivationResult("comp")
		assert.NoError(t, r.ActivationErrorWithComponentName())
	})
}

func TestActivationResult_IsWaitingForInput(t *testing.T) {
	t.Run("is waiting", func(t *testing.T) {
		r := NewActivationResult("c").SetActivationCode(ActivationCodeWaitingForInputsClear)
		assert.True(t, IsWaitingForInput(r))
	})

	t.Run("is waiting and keeping inputs", func(t *testing.T) {
		r := NewActivationResult("c").SetActivationCode(ActivationCodeWaitingForInputsKeep)
		assert.True(t, IsWaitingForInput(r))
	})

	t.Run("not waiting", func(t *testing.T) {
		r := NewActivationResult("c").SetActivationCode(ActivationCodeOK)
		assert.False(t, IsWaitingForInput(r))
	})
}

func TestActivationResult_WantsToKeepInputs(t *testing.T) {
	t.Run("wants to keep", func(t *testing.T) {
		r := NewActivationResult("c").SetActivationCode(ActivationCodeWaitingForInputsKeep)
		assert.True(t, WantsToKeepInputs(r))
	})

	t.Run("does not want to keep", func(t *testing.T) {
		r := NewActivationResult("c").SetActivationCode(ActivationCodeWaitingForInputsClear)
		assert.False(t, WantsToKeepInputs(r))
	})

	t.Run("not waiting", func(t *testing.T) {
		r := NewActivationResult("c").SetActivationCode(ActivationCodeOK)
		assert.False(t, WantsToKeepInputs(r))
	})
}

func TestActivationResultCollection_FindAny(t *testing.T) {
	r1 := NewActivationResult("c1").SetActivated(true)
	r2 := NewActivationResult("c2").SetActivated(false)

	t.Run("one found", func(t *testing.T) {
		collection := NewActivationResultCollection().Add(r1, r2)
		result := collection.FindAny(func(r *ActivationResult) bool {
			return r.Activated()
		})
		assert.Equal(t, "c1", result.ComponentName())
	})

	t.Run("none match", func(t *testing.T) {
		collection := NewActivationResultCollection().Add(r2)
		result := collection.FindAny(func(r *ActivationResult) bool {
			return r.ComponentName() == "c3"
		})
		assert.Nil(t, result)
	})

	t.Run("empty collection returns nil", func(t *testing.T) {
		collection := NewActivationResultCollection()
		result := collection.FindAny(func(r *ActivationResult) bool {
			return true
		})
		assert.Nil(t, result)
	})

	t.Run("returns the first match in component-name order", func(t *testing.T) {
		// Ranging over the map returned a different match on identical calls.
		collection, names := newNamedResultCollection()
		for range 10 {
			result := collection.FindAny(func(*ActivationResult) bool { return true })
			require.NotNil(t, result)
			assert.Equal(t, names[0], result.ComponentName())
		}
	})
}

func TestActivationResultCollection_Filter(t *testing.T) {
	r1 := NewActivationResult("c1").SetActivated(true)
	r2 := NewActivationResult("c2").SetActivated(false)

	t.Run("one found", func(t *testing.T) {
		collection := NewActivationResultCollection().Add(r1, r2)
		result := collection.Filter(func(r *ActivationResult) bool {
			return r.Activated()
		})
		assert.False(t, result.IsEmpty())
		assert.Equal(t, 1, result.Len())
	})

	t.Run("none match", func(t *testing.T) {
		collection := NewActivationResultCollection().Add(r2)
		result := collection.Filter(func(r *ActivationResult) bool {
			return r.ComponentName() == "c3"
		})
		assert.True(t, result.IsEmpty())
	})

	t.Run("empty collection returns empty collection", func(t *testing.T) {
		collection := NewActivationResultCollection()
		result := collection.Filter(func(r *ActivationResult) bool {
			return true
		})
		assert.True(t, result.IsEmpty())
	})
}

// mustNew is a test helper that creates a component and panics on error.
func mustNew(name string, opts ...Option) *Component {
	c, err := New(name, opts...)
	if err != nil {
		panic(err)
	}
	return c
}

func TestActivationResult_WithActivationError_Accumulates(t *testing.T) {
	err1 := errors.New("first error")
	err2 := errors.New("second error")
	err3 := errors.New("third error")

	t.Run("single error", func(t *testing.T) {
		r := NewActivationResult("c").AddActivationError(err1)
		assert.Len(t, r.ActivationErrors(), 1)
		require.Error(t, r.ActivationError())
		assert.ErrorIs(t, r.ActivationError(), err1)
	})

	t.Run("multiple errors accumulate", func(t *testing.T) {
		r := NewActivationResult("c").
			AddActivationError(err1).
			AddActivationError(err2).
			AddActivationError(err3)
		assert.Len(t, r.ActivationErrors(), 3)
		require.ErrorIs(t, r.ActivationError(), err1)
		require.ErrorIs(t, r.ActivationError(), err2)
		assert.ErrorIs(t, r.ActivationError(), err3)
	})

	t.Run("no errors returns nil", func(t *testing.T) {
		r := NewActivationResult("c")
		assert.Empty(t, r.ActivationErrors())
		require.NoError(t, r.ActivationError())
	})
}

// TestActivationResultCollection_ConcurrentAddAndRead pins the collection's
// thread-safety contract: activation goroutines Add results while the run loop
// reads. Run under -race (CI does); assertions are deterministic.
func TestActivationResultCollection_ConcurrentAddAndRead(t *testing.T) {
	const writers, perWriter, readers = 8, 25, 4

	collection := NewActivationResultCollection()

	var wg sync.WaitGroup
	for w := range writers {
		wg.Go(func() {
			for i := range perWriter {
				collection.Add(NewActivationResult(fmt.Sprintf("c%d_%d", w, i)).SetActivated(true))
			}
		})
	}
	for range readers {
		wg.Go(func() {
			for i := range writers * perWriter {
				_ = collection.Len()
				_ = collection.ByName(fmt.Sprintf("c0_%d", i%perWriter))
			}
		})
	}
	wg.Wait()

	assert.Equal(t, writers*perWriter, collection.Len())
	for w := range writers {
		for i := range perWriter {
			name := fmt.Sprintf("c%d_%d", w, i)
			require.NotNil(t, collection.ByName(name), "result %q must survive the concurrent adds", name)
		}
	}
}
