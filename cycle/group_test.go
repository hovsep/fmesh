package cycle

import (
	"github.com/hovsep/fmesh/component"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewGroup(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		group := NewGroup()
		assert.NotNil(t, group)
	})
}

func TestGroup_With(t *testing.T) {
	type args struct {
		cycles []*Cycle
	}
	tests := []struct {
		name       string
		group      *Group
		args       args
		assertions func(t *testing.T, group *Group)
	}{
		{
			name:  "no addition to empty group",
			group: NewGroup(),
			args: args{
				cycles: nil,
			},
			assertions: func(t *testing.T, group *Group) {
				assert.Zero(t, group.Len())
			},
		},
		{
			name:  "adding nothing to existing group",
			group: NewGroup().With(New().WithActivationResults(component.NewActivationResult("c1").SetActivated(false))),
			args: args{
				cycles: nil,
			},
			assertions: func(t *testing.T, group *Group) {
				assert.Equal(t, 1, group.Len())
			},
		},
		{
			name:  "adding to empty group",
			group: NewGroup(),
			args: args{
				cycles: []*Cycle{New().WithActivationResults(component.NewActivationResult("c1").SetActivated(false))},
			},
			assertions: func(t *testing.T, group *Group) {
				assert.Equal(t, 1, group.Len())
			},
		},
		{
			name:  "adding to existing group",
			group: NewGroup().With(New().WithActivationResults(component.NewActivationResult("c1").SetActivated(true))),
			args: args{
				cycles: []*Cycle{New().WithActivationResults(component.NewActivationResult("c1").SetActivated(false))},
			},
			assertions: func(t *testing.T, group *Group) {
				assert.Equal(t, 2, group.Len())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			groupAfter := tt.group.With(tt.args.cycles...)
			if tt.assertions != nil {
				tt.assertions(t, groupAfter)
			}
		})
	}
}
