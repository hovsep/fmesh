package component

import (
	"errors"
	"testing"

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
					New("c1").newActivationResultOK(),
					New("c2").newActivationResultReturnedError(errors.New("oops")),
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
				New("c1").newActivationResultOK(),
				New("c2").newActivationResultOK(),
			),
			args: args{
				activationResults: []*ActivationResult{
					New("c4").newActivationResultNoInput(),
					New("c5").newActivationResultPanicked(errors.New("panic")),
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
		all, err := collection.All()
		require.NoError(t, err)
		assert.Len(t, all, 2)
		assert.Contains(t, all, "c1")
		assert.Contains(t, all, "c2")
	})

	t.Run("empty collection", func(t *testing.T) {
		collection := NewActivationResultCollection()
		all, err := collection.All()
		require.NoError(t, err)
		assert.Empty(t, all)
	})
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

func TestActivationResultCollection_AllMatch(t *testing.T) {
	r1 := NewActivationResult("c1").SetActivated(true)
	r2 := NewActivationResult("c2").SetActivated(true)
	r3 := NewActivationResult("c3").SetActivated(false)

	t.Run("all match", func(t *testing.T) {
		collection := NewActivationResultCollection().Add(r1, r2)
		result := collection.AllMatch(func(r *ActivationResult) bool {
			return r.Activated()
		})
		assert.True(t, result)
	})

	t.Run("not all match", func(t *testing.T) {
		collection := NewActivationResultCollection().Add(r1, r3)
		result := collection.AllMatch(func(r *ActivationResult) bool {
			return r.Activated()
		})
		assert.False(t, result)
	})

	t.Run("empty collection returns true", func(t *testing.T) {
		collection := NewActivationResultCollection()
		result := collection.AllMatch(func(r *ActivationResult) bool {
			return false
		})
		assert.True(t, result)
	})
}

func TestActivationResultCollection_AnyMatch(t *testing.T) {
	r1 := NewActivationResult("c1").SetActivated(true)
	r2 := NewActivationResult("c2").SetActivated(false)

	t.Run("at least one matches", func(t *testing.T) {
		collection := NewActivationResultCollection().Add(r1, r2)
		result := collection.AnyMatch(func(r *ActivationResult) bool {
			return r.Activated()
		})
		assert.True(t, result)
	})

	t.Run("none match", func(t *testing.T) {
		collection := NewActivationResultCollection().Add(r2)
		result := collection.AnyMatch(func(r *ActivationResult) bool {
			return r.Activated()
		})
		assert.False(t, result)
	})

	t.Run("empty collection returns false", func(t *testing.T) {
		collection := NewActivationResultCollection()
		result := collection.AnyMatch(func(r *ActivationResult) bool {
			return true
		})
		assert.False(t, result)
	})
}

func TestActivationResultCollection_CountMatch(t *testing.T) {
	r1 := NewActivationResult("c1").SetActivated(true)
	r2 := NewActivationResult("c2").SetActivated(false)
	r3 := NewActivationResult("c3").SetActivated(true)

	t.Run("counts matching results", func(t *testing.T) {
		collection := NewActivationResultCollection().Add(r1, r2, r3)
		count := collection.CountMatch(func(r *ActivationResult) bool {
			return r.Activated()
		})
		assert.Equal(t, 2, count)
	})

	t.Run("no matches", func(t *testing.T) {
		collection := NewActivationResultCollection().Add(r2)
		count := collection.CountMatch(func(r *ActivationResult) bool {
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
		collection.ForEach(func(r *ActivationResult) error {
			count++
			return nil
		})
		assert.Equal(t, 3, count)
	})

	t.Run("empty collection", func(t *testing.T) {
		collection := NewActivationResultCollection()
		count := 0
		collection.ForEach(func(r *ActivationResult) error {
			count++
			return nil
		})
		assert.Equal(t, 0, count)
	})
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

func TestActivationResultCollection_Without(t *testing.T) {
	r1 := NewActivationResult("c1").SetActivated(true)
	r2 := NewActivationResult("c2").SetActivated(false)
	r3 := NewActivationResult("c3").SetActivated(true)

	t.Run("removes by component name", func(t *testing.T) {
		collection := NewActivationResultCollection().Add(r1, r2, r3)
		result := collection.Without("c2")
		assert.Equal(t, 2, result.Len())
	})

	t.Run("removes multiple", func(t *testing.T) {
		collection := NewActivationResultCollection().Add(r1, r2, r3)
		result := collection.Without("c1", "c2")
		assert.Equal(t, 1, result.Len())
	})

	t.Run("removes all", func(t *testing.T) {
		collection := NewActivationResultCollection().Add(r1, r2)
		result := collection.Without("c1", "c2")
		assert.Equal(t, 0, result.Len())
	})
}

func TestActivationResultCollection_ChainableErr(t *testing.T) {
	t.Run("with error", func(t *testing.T) {
		collection := NewActivationResultCollection().WithChainableErr(errors.New("test error"))
		assert.True(t, collection.HasChainableErr())
		assert.EqualError(t, collection.ChainableErr(), "test error")
	})

	t.Run("without error", func(t *testing.T) {
		collection := NewActivationResultCollection()
		assert.False(t, collection.HasChainableErr())
		assert.NoError(t, collection.ChainableErr())
	})
}

func TestActivationResult_ActivationErrorWithComponentName(t *testing.T) {
	err := errors.New("activation failed")
	r := NewActivationResult("my-component").WithActivationError(err)

	t.Run("returns error with component name", func(t *testing.T) {
		wrappedErr := r.ActivationErrorWithComponentName()
		require.Error(t, wrappedErr)
		assert.Contains(t, wrappedErr.Error(), "my-component")
		assert.Contains(t, wrappedErr.Error(), "activation failed")
	})

	t.Run("wraps nil activation error", func(t *testing.T) {
		r := NewActivationResult("comp")
		wrappedErr := r.ActivationErrorWithComponentName()
		// The method wraps even nil errors, so it always returns an error
		require.Error(t, wrappedErr)
		assert.Contains(t, wrappedErr.Error(), "comp")
	})
}

func TestActivationResult_IsWaitingForInput(t *testing.T) {
	t.Run("is waiting", func(t *testing.T) {
		r := NewActivationResult("c").WithActivationCode(ActivationCodeWaitingForInputsClear)
		assert.True(t, IsWaitingForInput(r))
	})

	t.Run("is waiting and keeping inputs", func(t *testing.T) {
		r := NewActivationResult("c").WithActivationCode(ActivationCodeWaitingForInputsKeep)
		assert.True(t, IsWaitingForInput(r))
	})

	t.Run("not waiting", func(t *testing.T) {
		r := NewActivationResult("c").WithActivationCode(ActivationCodeOK)
		assert.False(t, IsWaitingForInput(r))
	})
}

func TestActivationResult_WantsToKeepInputs(t *testing.T) {
	t.Run("wants to keep", func(t *testing.T) {
		r := NewActivationResult("c").WithActivationCode(ActivationCodeWaitingForInputsKeep)
		assert.True(t, WantsToKeepInputs(r))
	})

	t.Run("does not want to keep", func(t *testing.T) {
		r := NewActivationResult("c").WithActivationCode(ActivationCodeWaitingForInputsClear)
		assert.False(t, WantsToKeepInputs(r))
	})

	t.Run("not waiting", func(t *testing.T) {
		r := NewActivationResult("c").WithActivationCode(ActivationCodeOK)
		assert.False(t, WantsToKeepInputs(r))
	})
}

func TestActivationResult_ChainableErr(t *testing.T) {
	t.Run("with error", func(t *testing.T) {
		r := NewActivationResult("c").WithChainableErr(errors.New("chain error"))
		assert.True(t, r.HasChainableErr())
		assert.ErrorContains(t, r.ChainableErr(), "chain error")
	})

	t.Run("without error", func(t *testing.T) {
		r := NewActivationResult("c")
		assert.False(t, r.HasChainableErr())
		assert.NoError(t, r.ChainableErr())
	})
}
