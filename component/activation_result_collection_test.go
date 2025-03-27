package component

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
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
				assert.Equal(t, collection.Len(), 0)
				assert.False(t, collection.HasErrors())
				assert.False(t, collection.HasPanics())
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
				assert.Equal(t, collection.Len(), 2)
				assert.True(t, collection.HasActivatedComponents())
				assert.True(t, collection.HasErrors())
				assert.False(t, collection.HasPanics())
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
				assert.Equal(t, collection.Len(), 4)
				assert.True(t, collection.HasPanics())
				assert.False(t, collection.HasErrors())
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
