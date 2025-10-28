package component

import (
	"errors"
	"testing"

	"github.com/hovsep/fmesh/labels"
	"github.com/hovsep/fmesh/port"
	"github.com/hovsep/fmesh/signal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComponent_WithInputs(t *testing.T) {
	type args struct {
		portNames []string
	}
	tests := []struct {
		name       string
		component  *Component
		args       args
		assertions func(t *testing.T, component *Component)
	}{
		{
			name: "happy path",
			component: New("c1").WithActivationFunc(func(this *Component) error {
				return nil
			}),
			args: args{
				portNames: []string{"p1", "p2"},
			},
			assertions: func(t *testing.T, component *Component) {
				assert.Equal(t, "c1", component.Name())
				assert.Equal(t, 2, component.Inputs().Len())
				assert.Zero(t, component.Outputs().Len())
				assert.Empty(t, component.Description())
				assert.Zero(t, component.labels.Len())
				assert.True(t, component.hasActivationFunction())
			},
		},
		{
			name:      "no arg",
			component: New("c1"),
			args: args{
				portNames: nil,
			},
			assertions: func(t *testing.T, component *Component) {
				assert.Equal(t, "c1", component.Name())
				assert.Zero(t, component.Inputs().Len())
				assert.Zero(t, component.Outputs().Len())
				assert.Empty(t, component.Description())
				assert.Zero(t, component.labels.Len())
				assert.False(t, component.hasActivationFunction())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			componentAfter := tt.component.WithInputs(tt.args.portNames...)
			if tt.assertions != nil {
				tt.assertions(t, componentAfter)
			}
		})
	}
}

func TestComponent_WithOutputs(t *testing.T) {
	type args struct {
		portNames []string
	}
	tests := []struct {
		name       string
		component  *Component
		args       args
		assertions func(t *testing.T, component *Component)
	}{
		{
			name: "happy path",
			component: New("c1").WithActivationFunc(func(this *Component) error {
				return nil
			}),
			args: args{
				portNames: []string{"p1", "p2"},
			},
			assertions: func(t *testing.T, component *Component) {
				assert.Equal(t, "c1", component.Name())
				assert.Equal(t, 2, component.Outputs().Len())
				assert.Zero(t, component.Inputs().Len())
				assert.Empty(t, component.Description())
				assert.Zero(t, component.labels.Len())
				assert.True(t, component.hasActivationFunction())
			},
		},
		{
			name:      "no arg",
			component: New("c1"),
			args: args{
				portNames: nil,
			},
			assertions: func(t *testing.T, component *Component) {
				assert.Equal(t, "c1", component.Name())
				assert.Zero(t, component.Inputs().Len())
				assert.Zero(t, component.Outputs().Len())
				assert.Empty(t, component.Description())
				assert.Zero(t, component.labels.Len())
				assert.False(t, component.hasActivationFunction())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			componentAfter := tt.component.WithOutputs(tt.args.portNames...)
			if tt.assertions != nil {
				tt.assertions(t, componentAfter)
			}
		})
	}
}

func TestComponent_WithInputsIndexed(t *testing.T) {
	type args struct {
		prefix     string
		startIndex int
		endIndex   int
	}
	tests := []struct {
		name       string
		component  *Component
		args       args
		assertions func(t *testing.T, component *Component)
	}{
		{
			name:      "component has no ports before",
			component: New("c").WithOutputs("o1", "o2"),
			args: args{
				prefix:     "p",
				startIndex: 1,
				endIndex:   3,
			},
			assertions: func(t *testing.T, component *Component) {
				assert.Equal(t, 2, component.Outputs().Len())
				assert.Equal(t, 3, component.Inputs().Len())
			},
		},
		{
			name:      "component has ports before",
			component: New("c").WithInputs("i1", "i2").WithOutputs("o1", "o2"),
			args: args{
				prefix:     "p",
				startIndex: 1,
				endIndex:   3,
			},
			assertions: func(t *testing.T, component *Component) {
				assert.Equal(t, 2, component.Outputs().Len())
				assert.Equal(t, 5, component.Inputs().Len())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			componentAfter := tt.component.WithInputsIndexed(tt.args.prefix, tt.args.startIndex, tt.args.endIndex)
			if tt.assertions != nil {
				tt.assertions(t, componentAfter)
			}
		})
	}
}

func TestComponent_WithOutputsIndexed(t *testing.T) {
	type args struct {
		prefix     string
		startIndex int
		endIndex   int
	}
	tests := []struct {
		name       string
		component  *Component
		args       args
		assertions func(t *testing.T, component *Component)
	}{
		{
			name:      "component has no ports before",
			component: New("c").WithInputs("i1", "i2"),
			args: args{
				prefix:     "p",
				startIndex: 1,
				endIndex:   3,
			},
			assertions: func(t *testing.T, component *Component) {
				assert.Equal(t, 2, component.Inputs().Len())
				assert.Equal(t, 3, component.Outputs().Len())
			},
		},
		{
			name:      "component has ports before",
			component: New("c").WithInputs("i1", "i2").WithOutputs("o1", "o2"),
			args: args{
				prefix:     "p",
				startIndex: 1,
				endIndex:   3,
			},
			assertions: func(t *testing.T, component *Component) {
				assert.Equal(t, 2, component.Inputs().Len())
				assert.Equal(t, 5, component.Outputs().Len())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			componentAfter := tt.component.WithOutputsIndexed(tt.args.prefix, tt.args.startIndex, tt.args.endIndex)
			if tt.assertions != nil {
				tt.assertions(t, componentAfter)
			}
		})
	}
}

func TestComponent_Inputs(t *testing.T) {
	tests := []struct {
		name       string
		component  *Component
		assertions func(t *testing.T, collection *port.Collection)
	}{
		{
			name:      "no inputs",
			component: New("c1"),
			assertions: func(t *testing.T, collection *port.Collection) {
				assert.Zero(t, collection.Len())
			},
		},
		{
			name:      "with inputs",
			component: New("c1").WithInputs("i1", "i2"),
			assertions: func(t *testing.T, collection *port.Collection) {
				assert.Equal(t, 2, collection.Len())
				assert.True(t, collection.AllMatch(func(p *port.Port) bool {
					return p.Labels().ValueIs(port.DirectionLabel, port.DirectionIn)
				}))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.component.Inputs()
			if tt.assertions != nil {
				tt.assertions(t, got)
			}
		})
	}
}

func TestComponent_Outputs(t *testing.T) {
	tests := []struct {
		name       string
		component  *Component
		assertions func(t *testing.T, collection *port.Collection)
	}{
		{
			name:      "no outputs",
			component: New("c1"),
			assertions: func(t *testing.T, collection *port.Collection) {
				assert.Zero(t, collection.Len())
			},
		},
		{
			name:      "with outputs",
			component: New("c1").WithOutputs("o1", "o2"),
			assertions: func(t *testing.T, collection *port.Collection) {
				assert.Equal(t, 2, collection.Len())
				assert.True(t, collection.AllMatch(func(p *port.Port) bool {
					return p.Labels().ValueIs(port.DirectionLabel, port.DirectionOut)
				}))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.component.Outputs()
			if tt.assertions != nil {
				tt.assertions(t, got)
			}
		})
	}
}

func TestComponent_ShortcutMethods(t *testing.T) {
	t.Run("InputByName", func(t *testing.T) {
		c := New("c").WithInputs("a", "b", "c")
		assert.Equal(t, "b", c.InputByName("b").Name())
	})

	t.Run("OutputByName", func(t *testing.T) {
		c := New("c").WithOutputs("a", "b", "c")
		assert.Equal(t, "c", c.OutputByName("c").Name())
	})
}

func TestComponent_ClearInputs(t *testing.T) {
	tests := []struct {
		name         string
		getComponent func() *Component
		assertions   func(t *testing.T, componentAfter *Component)
	}{
		{
			name: "no side effects",
			getComponent: func() *Component {
				return New("c").WithInputs("i1").WithOutputs("o1")
			},
			assertions: func(t *testing.T, componentAfter *Component) {
				assert.Equal(t, 1, componentAfter.Inputs().Len())
				assert.Equal(t, 1, componentAfter.Outputs().Len())
				assert.False(t, componentAfter.Inputs().AnyHasSignals())
				assert.False(t, componentAfter.Outputs().AnyHasSignals())
			},
		},
		{
			name: "only inputs are cleared",
			getComponent: func() *Component {
				c := New("c").WithInputs("i1").WithOutputs("o1")
				c.Inputs().ByName("i1").PutSignals(signal.New(10))
				c.Outputs().ByName("o1").PutSignals(signal.New(20))
				return c
			},
			assertions: func(t *testing.T, componentAfter *Component) {
				assert.Equal(t, 1, componentAfter.Inputs().Len())
				assert.Equal(t, 1, componentAfter.Outputs().Len())
				assert.False(t, componentAfter.Inputs().AnyHasSignals())
				assert.True(t, componentAfter.Outputs().ByName("o1").HasSignals())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			componentAfter := tt.getComponent().ClearInputs()
			if tt.assertions != nil {
				tt.assertions(t, componentAfter)
			}
		})
	}
}

func TestComponent_FlushOutputs(t *testing.T) {
	tests := []struct {
		name         string
		getComponent func() *Component
		assertions   func(t *testing.T, componentAfterFlush *Component)
	}{
		{
			name: "no outputs",
			getComponent: func() *Component {
				return New("c1")
			},
			assertions: func(t *testing.T, componentAfterFlush *Component) {
				assert.NotNil(t, componentAfterFlush.Outputs())
				assert.Zero(t, componentAfterFlush.Outputs().Len())
			},
		},
		{
			name: "output has no signal set",
			getComponent: func() *Component {
				return New("c1").WithOutputs("o1", "o2")
			},
			assertions: func(t *testing.T, componentAfterFlush *Component) {
				assert.False(t, componentAfterFlush.Outputs().AnyHasSignals())
			},
		},
		{
			name: "happy path",
			getComponent: func() *Component {
				sink := port.New("sink").WithLabels(labels.Map{
					port.DirectionLabel: port.DirectionIn,
				})
				c := New("c1").WithOutputs("o1", "o2")
				c.Outputs().ByNames("o1").PutSignals(signal.New(777))
				c.Outputs().ByNames("o2").PutSignals(signal.New(888))
				c.Outputs().ByNames("o1", "o2").PipeTo(sink)
				return c
			},
			assertions: func(t *testing.T, componentAfterFlush *Component) {
				destPort := componentAfterFlush.OutputByName("o1").Pipes().FirstOrNil()
				allPayloads, err := destPort.Signals().AllPayloads()
				require.NoError(t, err)
				assert.Contains(t, allPayloads, 777)
				assert.Contains(t, allPayloads, 888)
				assert.Len(t, allPayloads, 2)
				// Buffer is cleared when port is flushed
				assert.False(t, componentAfterFlush.Outputs().AnyHasSignals())
			},
		},
		{
			name: "with chain error",
			getComponent: func() *Component {
				sink := port.New("sink")
				c := New("c").WithOutputs("o1").WithChainableErr(errors.New("some error"))
				// Lines below are ignored as error immediately propagates up to component level
				c.Outputs().ByName("o1").PipeTo(sink)
				c.Outputs().ByName("o1").PutSignals(signal.New("signal from component with chain error"))
				return c
			},
			assertions: func(t *testing.T, componentAfterFlush *Component) {
				assert.False(t, componentAfterFlush.OutputByName("o1").HasPipes())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			componentAfter := tt.getComponent().FlushOutputs()
			tt.assertions(t, componentAfter)
		})
	}
}
