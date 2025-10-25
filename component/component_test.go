package component

import (
	"testing"

	"github.com/hovsep/fmesh/common"
	"github.com/hovsep/fmesh/labels"
	"github.com/hovsep/fmesh/port"
	"github.com/stretchr/testify/assert"
)

func TestNewComponent(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name string
		args args
		want *Component
	}{
		{
			name: "empty name is valid",
			args: args{
				name: "",
			},
			want: New(""),
		},
		{
			name: "with name",
			args: args{
				name: "multiplier",
			},
			want: New("multiplier"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, New(tt.args.name))
		})
	}
}

func TestComponent_WithDescription(t *testing.T) {
	type args struct {
		description string
	}
	tests := []struct {
		name      string
		component *Component
		args      args
		want      *Component
	}{
		{
			name:      "happy path",
			component: New("c1"),
			args: args{
				description: "descr",
			},
			want: &Component{
				name:        "c1",
				description: "descr",
				labels:      labels.NewCollection(nil),
				Chainable:   common.NewChainable(),
				inputs: port.NewCollection().WithDefaultLabels(labels.Map{
					port.DirectionLabel: port.DirectionIn,
				}),
				outputs: port.NewCollection().WithDefaultLabels(labels.Map{
					port.DirectionLabel: port.DirectionOut,
				}),
				f:     nil,
				state: NewState(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.component.WithDescription(tt.args.description))
		})
	}
}

func TestComponent_WithLabels(t *testing.T) {
	type args struct {
		labels labels.Map
	}
	tests := []struct {
		name       string
		component  *Component
		args       args
		assertions func(t *testing.T, component *Component)
	}{
		{
			name:      "happy path",
			component: New("c1"),
			args: args{
				labels: labels.Map{
					"l1": "v1",
					"l2": "v2",
				},
			},
			assertions: func(t *testing.T, component *Component) {
				assert.Equal(t, 2, component.Labels().Len())
				assert.True(t, component.labels.HasAll("l1", "l2"))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			componentAfter := tt.component.WithLabels(tt.args.labels)
			if tt.assertions != nil {
				tt.assertions(t, componentAfter)
			}
		})
	}
}
