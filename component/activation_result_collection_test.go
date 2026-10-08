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
					NewActivationResult("c1", ActivationCodeOK),
					NewActivationResult("c2", ActivationCodeReturnedError, errors.New("oops")),
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
				NewActivationResult("c1", ActivationCodeOK),
				NewActivationResult("c2", ActivationCodeOK),
			),
			args: args{
				activationResults: []*ActivationResult{
					NewActivationResult("c4", ActivationCodeNoInput),
					NewActivationResult("c5", ActivationCodePanicked, errors.New("panic")),
				},
			},
			assertions: func(t *testing.T, collection *ActivationResultCollection) {
				assert.Equal(t, 4, collection.Len())
				assert.True(t, collection.HasActivationPanics())
				assert.False(t, collection.HasActivationErrors())
				assert.True(t, collection.HasActivatedComponents())
			},
		},
		{
			// The collection keeps a name-sorted slice; unsorted input must still read back sorted.
			name:       "unsorted batch with a repeated name, last one wins",
			collection: NewActivationResultCollection(),
			args: args{
				activationResults: []*ActivationResult{
					NewActivationResult("c3", ActivationCodeOK),
					NewActivationResult("c1", ActivationCodeOK),
					NewActivationResult("c3", ActivationCodePanicked, errors.New("panic")),
					NewActivationResult("c2", ActivationCodeOK),
				},
			},
			assertions: func(t *testing.T, collection *ActivationResultCollection) {
				names := make([]string, 0, collection.Len())
				for ar := range collection.All() {
					names = append(names, ar.ComponentName())
				}
				assert.Equal(t, []string{"c1", "c2", "c3"}, names)
				assert.Equal(t, ActivationCodePanicked, collection.ByName("c3").Code())
			},
		},
		{
			name: "a result for a component already present replaces it",
			collection: NewActivationResultCollection().Add(
				NewActivationResult("c1", ActivationCodeOK),
				NewActivationResult("c3", ActivationCodeOK),
			),
			args: args{
				activationResults: []*ActivationResult{
					NewActivationResult("c3", ActivationCodeReturnedError, errors.New("oops")),
					NewActivationResult("c2", ActivationCodeOK),
				},
			},
			assertions: func(t *testing.T, collection *ActivationResultCollection) {
				assert.Equal(t, 3, collection.Len())
				assert.Equal(t, ActivationCodeReturnedError, collection.ByName("c3").Code())
				assert.Equal(t, "c2", slices.Collect(collection.All())[1].ComponentName())
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
	r1 := NewActivationResult("c1", ActivationCodeOK)
	r2 := NewActivationResult("c2", ActivationCodeUndefined)
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

func TestActivationResultCollection_IsEmpty(t *testing.T) {
	t.Run("empty collection", func(t *testing.T) {
		collection := NewActivationResultCollection()
		assert.True(t, collection.IsEmpty())
	})

	t.Run("non-empty collection", func(t *testing.T) {
		collection := NewActivationResultCollection().Add(NewActivationResult("c1", ActivationCodeUndefined))
		assert.False(t, collection.IsEmpty())
	})
}

func TestActivationResultCollection_Every(t *testing.T) {
	r1 := NewActivationResult("c1", ActivationCodeOK)
	r2 := NewActivationResult("c2", ActivationCodeOK)
	r3 := NewActivationResult("c3", ActivationCodeUndefined)

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
	r1 := NewActivationResult("c1", ActivationCodeOK)
	r2 := NewActivationResult("c2", ActivationCodeUndefined)

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
	r1 := NewActivationResult("c1", ActivationCodeOK)
	r2 := NewActivationResult("c2", ActivationCodeUndefined)
	r3 := NewActivationResult("c3", ActivationCodeOK)

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

func TestActivationResultCollection_All(t *testing.T) {
	t.Run("empty collection", func(t *testing.T) {
		assert.Empty(t, slices.Collect(NewActivationResultCollection().All()))
	})

	t.Run("visits results in component-name order", func(t *testing.T) {
		// Ranging over the map visited results in a different order on every call.
		collection, names := newNamedResultCollection()
		var visited []string
		for r := range collection.All() {
			visited = append(visited, r.ComponentName())
		}
		assert.Equal(t, names, visited)
	})

	t.Run("loop body may change the collection", func(t *testing.T) {
		// A body that ran under the read lock deadlocked on Add.
		collection, names := newNamedResultCollection()
		done := make(chan []string, 1)
		go func() {
			var visited []string
			for r := range collection.All() {
				visited = append(visited, r.ComponentName())
				collection.Add(NewActivationResult("added-"+r.ComponentName(), ActivationCodeUndefined))
			}
			done <- visited
		}()
		select {
		case visited := <-done:
			assert.Equal(t, names, visited, "the walk keeps the list it started with")
			assert.Equal(t, 2*len(names), collection.Len())
		case <-time.After(5 * time.Second):
			t.Fatal("All deadlocked when the loop body changed the collection")
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
		collection.Add(NewActivationResult(name, ActivationCodeUndefined))
	}
	return collection, names
}

func TestActivationResult_IsWaiting(t *testing.T) {
	t.Run("is waiting", func(t *testing.T) {
		r := NewActivationResult("c", ActivationCodeWaitingForInputsClear)
		assert.True(t, r.IsWaiting())
	})

	t.Run("is waiting and keeping inputs", func(t *testing.T) {
		r := NewActivationResult("c", ActivationCodeWaitingForInputsKeep)
		assert.True(t, r.IsWaiting())
	})

	t.Run("not waiting", func(t *testing.T) {
		r := NewActivationResult("c", ActivationCodeOK)
		assert.False(t, r.IsWaiting())
	})
}

func TestActivationResult_KeepsInputs(t *testing.T) {
	t.Run("wants to keep", func(t *testing.T) {
		r := NewActivationResult("c", ActivationCodeWaitingForInputsKeep)
		assert.True(t, r.KeepsInputs())
	})

	t.Run("does not want to keep", func(t *testing.T) {
		r := NewActivationResult("c", ActivationCodeWaitingForInputsClear)
		assert.False(t, r.KeepsInputs())
	})

	t.Run("not waiting", func(t *testing.T) {
		r := NewActivationResult("c", ActivationCodeOK)
		assert.False(t, r.KeepsInputs())
	})
}

func TestActivationResultCollection_Find(t *testing.T) {
	r1 := NewActivationResult("c1", ActivationCodeOK)
	r2 := NewActivationResult("c2", ActivationCodeUndefined)

	t.Run("one found", func(t *testing.T) {
		collection := NewActivationResultCollection().Add(r1, r2)
		result := collection.Find(func(r *ActivationResult) bool {
			return r.Activated()
		})
		assert.Equal(t, "c1", result.ComponentName())
	})

	t.Run("none match", func(t *testing.T) {
		collection := NewActivationResultCollection().Add(r2)
		result := collection.Find(func(r *ActivationResult) bool {
			return r.ComponentName() == "c3"
		})
		assert.Nil(t, result)
	})

	t.Run("empty collection returns nil", func(t *testing.T) {
		collection := NewActivationResultCollection()
		result := collection.Find(func(r *ActivationResult) bool {
			return true
		})
		assert.Nil(t, result)
	})

	t.Run("returns the first match in component-name order", func(t *testing.T) {
		// Ranging over the map returned a different match on identical calls.
		collection, names := newNamedResultCollection()
		for range 10 {
			result := collection.Find(func(*ActivationResult) bool { return true })
			require.NotNil(t, result)
			assert.Equal(t, names[0], result.ComponentName())
		}
	})
}

func TestActivationResultCollection_Filter(t *testing.T) {
	r1 := NewActivationResult("c1", ActivationCodeOK)
	r2 := NewActivationResult("c2", ActivationCodeUndefined)

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

func TestActivationResult_Errors(t *testing.T) {
	err1 := errors.New("first error")
	err2 := errors.New("second error")
	err3 := errors.New("third error")

	t.Run("single error", func(t *testing.T) {
		r := NewActivationResult("c", ActivationCodeUndefined, err1)
		assert.Len(t, r.Errors(), 1)
		require.Error(t, r.Err())
		assert.ErrorIs(t, r.Err(), err1)
	})

	t.Run("multiple errors accumulate", func(t *testing.T) {
		r := NewActivationResult("c", ActivationCodeUndefined, err1, err2, err3)
		assert.Len(t, r.Errors(), 3)
		require.ErrorIs(t, r.Err(), err1)
		require.ErrorIs(t, r.Err(), err2)
		assert.ErrorIs(t, r.Err(), err3)
	})

	t.Run("no errors returns nil", func(t *testing.T) {
		r := NewActivationResult("c", ActivationCodeUndefined)
		assert.Empty(t, r.Errors())
		require.NoError(t, r.Err())
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
				collection.Add(NewActivationResult(fmt.Sprintf("c%d_%d", w, i), ActivationCodeOK))
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
