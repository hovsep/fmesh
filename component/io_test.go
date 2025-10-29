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

func TestComponent_AddInputs(t *testing.T) {
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
			componentAfter := tt.component.AddInputs(tt.args.portNames...)
			if tt.assertions != nil {
				tt.assertions(t, componentAfter)
			}
		})
	}
}

func TestComponent_AddOutputs(t *testing.T) {
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
			componentAfter := tt.component.AddOutputs(tt.args.portNames...)
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
			component: New("c").AddOutputs("o1", "o2"),
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
			component: New("c").AddInputs("i1", "i2").AddOutputs("o1", "o2"),
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
			component: New("c").AddInputs("i1", "i2"),
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
			component: New("c").AddInputs("i1", "i2").AddOutputs("o1", "o2"),
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
			component: New("c1").AddInputs("i1", "i2"),
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
			component: New("c1").AddOutputs("o1", "o2"),
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
		c := New("c").AddInputs("a", "b", "c")
		assert.Equal(t, "b", c.InputByName("b").Name())
	})

	t.Run("OutputByName", func(t *testing.T) {
		c := New("c").AddOutputs("a", "b", "c")
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
				return New("c").AddInputs("i1").AddOutputs("o1")
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
				c := New("c").AddInputs("i1").AddOutputs("o1")
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
				return New("c1").AddOutputs("o1", "o2")
			},
			assertions: func(t *testing.T, componentAfterFlush *Component) {
				assert.False(t, componentAfterFlush.Outputs().AnyHasSignals())
			},
		},
		{
			name: "happy path",
			getComponent: func() *Component {
				sink := port.New("sink").SetLabels(labels.Map{
					port.DirectionLabel: port.DirectionIn,
				})
				c := New("c1").AddOutputs("o1", "o2")
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
				c := New("c").AddOutputs("o1").WithChainableErr(errors.New("some error"))
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

func TestComponent_AttachInputPorts(t *testing.T) {
	tests := []struct {
		name       string
		component  *Component
		ports      []*port.Port
		assertions func(t *testing.T, component *Component)
	}{
		{
			name:      "add single input port with description",
			component: New("c1"),
			ports: []*port.Port{
				port.New("in1").WithDescription("input port 1"),
			},
			assertions: func(t *testing.T, component *Component) {
				assert.Equal(t, 1, component.Inputs().Len())
				assert.Equal(t, "input port 1", component.InputByName("in1").Description())
			},
		},
		{
			name:      "add multiple input ports with descriptions and labels",
			component: New("c1"),
			ports: []*port.Port{
				port.New("in1").
					WithDescription("first input").
					AddLabel("priority", "high"),
				port.New("in2").
					WithDescription("second input").
					AddLabel("priority", "low"),
			},
			assertions: func(t *testing.T, component *Component) {
				assert.Equal(t, 2, component.Inputs().Len())
				assert.Equal(t, "first input", component.InputByName("in1").Description())
				assert.Equal(t, "second input", component.InputByName("in2").Description())
				assert.True(t, component.InputByName("in1").Labels().ValueIs("priority", "high"))
				assert.True(t, component.InputByName("in2").Labels().ValueIs("priority", "low"))
			},
		},
		{
			name:      "add ports to existing inputs",
			component: New("c1").AddInputs("in1"),
			ports: []*port.Port{
				port.New("in2").WithDescription("second input"),
			},
			assertions: func(t *testing.T, component *Component) {
				assert.Equal(t, 2, component.Inputs().Len())
				assert.Equal(t, "second input", component.InputByName("in2").Description())
			},
		},
		{
			name:      "chainable",
			component: New("c1"),
			ports: []*port.Port{
				port.New("in1").WithDescription("input 1"),
			},
			assertions: func(t *testing.T, component *Component) {
				result := component.AttachInputPorts(
					port.New("in2").WithDescription("input 2"),
				)
				assert.Equal(t, 2, result.Inputs().Len())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			componentAfter := tt.component.AttachInputPorts(tt.ports...)
			if tt.assertions != nil {
				tt.assertions(t, componentAfter)
			}
		})
	}
}

func TestComponent_AttachOutputPorts(t *testing.T) {
	tests := []struct {
		name       string
		component  *Component
		ports      []*port.Port
		assertions func(t *testing.T, component *Component)
	}{
		{
			name:      "add single output port with description",
			component: New("c1"),
			ports: []*port.Port{
				port.New("out1").WithDescription("output port 1"),
			},
			assertions: func(t *testing.T, component *Component) {
				assert.Equal(t, 1, component.Outputs().Len())
				assert.Equal(t, "output port 1", component.OutputByName("out1").Description())
			},
		},
		{
			name:      "add multiple output ports with descriptions and labels",
			component: New("c1"),
			ports: []*port.Port{
				port.New("out1").
					WithDescription("first output").
					AddLabel("type", "result"),
				port.New("out2").
					WithDescription("second output").
					AddLabel("type", "error"),
			},
			assertions: func(t *testing.T, component *Component) {
				assert.Equal(t, 2, component.Outputs().Len())
				assert.Equal(t, "first output", component.OutputByName("out1").Description())
				assert.Equal(t, "second output", component.OutputByName("out2").Description())
				assert.True(t, component.OutputByName("out1").Labels().ValueIs("type", "result"))
				assert.True(t, component.OutputByName("out2").Labels().ValueIs("type", "error"))
			},
		},
		{
			name:      "add ports to existing outputs",
			component: New("c1").AddOutputs("out1"),
			ports: []*port.Port{
				port.New("out2").WithDescription("second output"),
			},
			assertions: func(t *testing.T, component *Component) {
				assert.Equal(t, 2, component.Outputs().Len())
				assert.Equal(t, "second output", component.OutputByName("out2").Description())
			},
		},
		{
			name:      "chainable",
			component: New("c1"),
			ports: []*port.Port{
				port.New("out1").WithDescription("output 1"),
			},
			assertions: func(t *testing.T, component *Component) {
				result := component.AttachOutputPorts(
					port.New("out2").WithDescription("output 2"),
				)
				assert.Equal(t, 2, result.Outputs().Len())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			componentAfter := tt.component.AttachOutputPorts(tt.ports...)
			if tt.assertions != nil {
				tt.assertions(t, componentAfter)
			}
		})
	}
}

func TestComponent_MultipleInputOutputCalls(t *testing.T) {
	t.Run("multiple AddInputs calls add ports incrementally", func(t *testing.T) {
		c := New("c1").
			AddInputs("in1").
			AddInputs("in2", "in3").
			AddInputs("in4")

		assert.Equal(t, 4, c.Inputs().Len())
		assert.NotNil(t, c.InputByName("in1"))
		assert.NotNil(t, c.InputByName("in2"))
		assert.NotNil(t, c.InputByName("in3"))
		assert.NotNil(t, c.InputByName("in4"))
	})

	t.Run("multiple AddOutputs calls add ports incrementally", func(t *testing.T) {
		c := New("c1").
			AddOutputs("out1").
			AddOutputs("out2", "out3").
			AddOutputs("out4")

		assert.Equal(t, 4, c.Outputs().Len())
		assert.NotNil(t, c.OutputByName("out1"))
		assert.NotNil(t, c.OutputByName("out2"))
		assert.NotNil(t, c.OutputByName("out3"))
		assert.NotNil(t, c.OutputByName("out4"))
	})

	t.Run("mixing AddInputs and AttachInputPorts", func(t *testing.T) {
		c := New("c1").
			AddInputs("in1", "in2").
			AttachInputPorts(
				port.New("in3").WithDescription("third input"),
				port.New("in4").WithDescription("fourth input").AddLabel("important", "true"),
			).
			AddInputs("in5")

		assert.Equal(t, 5, c.Inputs().Len())
		assert.Empty(t, c.InputByName("in1").Description())
		assert.Empty(t, c.InputByName("in2").Description())
		assert.Equal(t, "third input", c.InputByName("in3").Description())
		assert.Equal(t, "fourth input", c.InputByName("in4").Description())
		assert.True(t, c.InputByName("in4").Labels().ValueIs("important", "true"))
		assert.Empty(t, c.InputByName("in5").Description())
	})

	t.Run("mixing AddOutputs and AttachOutputPorts", func(t *testing.T) {
		c := New("c1").
			AddOutputs("out1", "out2").
			AttachOutputPorts(
				port.New("out3").WithDescription("third output"),
				port.New("out4").WithDescription("fourth output").AddLabel("type", "error"),
			).
			AddOutputs("out5")

		assert.Equal(t, 5, c.Outputs().Len())
		assert.Empty(t, c.OutputByName("out1").Description())
		assert.Empty(t, c.OutputByName("out2").Description())
		assert.Equal(t, "third output", c.OutputByName("out3").Description())
		assert.Equal(t, "fourth output", c.OutputByName("out4").Description())
		assert.True(t, c.OutputByName("out4").Labels().ValueIs("type", "error"))
		assert.Empty(t, c.OutputByName("out5").Description())
	})
}
