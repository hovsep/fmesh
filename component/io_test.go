package component

import (
	"testing"

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
			component: func() *Component {
				c := mustNew("c1")
				c.SetActivationFunc(func(this *Component) error { return nil })
				return c
			}(),
			args: args{
				portNames: []string{"p1", "p2"},
			},
			assertions: func(t *testing.T, component *Component) {
				assert.Equal(t, "c1", component.Name())
				assert.Equal(t, 2, component.Inputs().Len())
				assert.Zero(t, component.Outputs().Len())
				assert.Empty(t, component.Description())
				assert.Zero(t, component.labels.Len())
			},
		},
		{
			name:      "no arg",
			component: mustNew("c1"),
			args: args{
				portNames: nil,
			},
			assertions: func(t *testing.T, component *Component) {
				assert.Equal(t, "c1", component.Name())
				assert.Zero(t, component.Inputs().Len())
				assert.Zero(t, component.Outputs().Len())
				assert.Empty(t, component.Description())
				assert.Zero(t, component.labels.Len())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.component.AddInputs(tt.args.portNames...)
			require.NoError(t, err)
			if tt.assertions != nil {
				tt.assertions(t, tt.component)
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
			component: func() *Component {
				c := mustNew("c1")
				c.SetActivationFunc(func(this *Component) error { return nil })
				return c
			}(),
			args: args{
				portNames: []string{"p1", "p2"},
			},
			assertions: func(t *testing.T, component *Component) {
				assert.Equal(t, "c1", component.Name())
				assert.Equal(t, 2, component.Outputs().Len())
				assert.Zero(t, component.Inputs().Len())
				assert.Empty(t, component.Description())
				assert.Zero(t, component.labels.Len())
			},
		},
		{
			name:      "no arg",
			component: mustNew("c1"),
			args: args{
				portNames: nil,
			},
			assertions: func(t *testing.T, component *Component) {
				assert.Equal(t, "c1", component.Name())
				assert.Zero(t, component.Inputs().Len())
				assert.Zero(t, component.Outputs().Len())
				assert.Empty(t, component.Description())
				assert.Zero(t, component.labels.Len())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.component.AddOutputs(tt.args.portNames...)
			require.NoError(t, err)
			if tt.assertions != nil {
				tt.assertions(t, tt.component)
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
			name: "component has no ports before",
			component: func() *Component {
				c := mustNew("c")
				require.NoError(t, c.AddOutputs("o1", "o2"))
				return c
			}(),
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
			name: "component has ports before",
			component: func() *Component {
				c := mustNew("c")
				require.NoError(t, c.AddInputs("i1", "i2"))
				require.NoError(t, c.AddOutputs("o1", "o2"))
				return c
			}(),
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
			err := tt.component.AddIndexedInputs(tt.args.prefix, tt.args.startIndex, tt.args.endIndex)
			require.NoError(t, err)
			if tt.assertions != nil {
				tt.assertions(t, tt.component)
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
			name: "component has no ports before",
			component: func() *Component {
				c := mustNew("c")
				require.NoError(t, c.AddInputs("i1", "i2"))
				return c
			}(),
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
			name: "component has ports before",
			component: func() *Component {
				c := mustNew("c")
				require.NoError(t, c.AddInputs("i1", "i2"))
				require.NoError(t, c.AddOutputs("o1", "o2"))
				return c
			}(),
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
			err := tt.component.AddIndexedOutputs(tt.args.prefix, tt.args.startIndex, tt.args.endIndex)
			require.NoError(t, err)
			if tt.assertions != nil {
				tt.assertions(t, tt.component)
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
			component: mustNew("c1"),
			assertions: func(t *testing.T, collection *port.Collection) {
				assert.Zero(t, collection.Len())
			},
		},
		{
			name: "with inputs",
			component: func() *Component {
				c := mustNew("c1")
				require.NoError(t, c.AddInputs("i1", "i2"))
				return c
			}(),
			assertions: func(t *testing.T, collection *port.Collection) {
				assert.Equal(t, 2, collection.Len())
				assert.True(t, collection.Every(func(p *port.Port) bool {
					return p.IsInput()
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
			component: mustNew("c1"),
			assertions: func(t *testing.T, collection *port.Collection) {
				assert.Zero(t, collection.Len())
			},
		},
		{
			name: "with outputs",
			component: func() *Component {
				c := mustNew("c1")
				require.NoError(t, c.AddOutputs("o1", "o2"))
				return c
			}(),
			assertions: func(t *testing.T, collection *port.Collection) {
				assert.Equal(t, 2, collection.Len())
				assert.True(t, collection.Every(func(p *port.Port) bool {
					return p.IsOutput()
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
		c := mustNew("c")
		require.NoError(t, c.AddInputs("a", "b", "c"))
		assert.Equal(t, "b", c.InputByName("b").Name())
	})

	t.Run("OutputByName", func(t *testing.T) {
		c := mustNew("c")
		require.NoError(t, c.AddOutputs("a", "b", "c"))
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
				c := mustNew("c")
				require.NoError(t, c.AddInputs("i1"))
				require.NoError(t, c.AddOutputs("o1"))
				return c
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
				c := mustNew("c")
				require.NoError(t, c.AddInputs("i1"))
				require.NoError(t, c.AddOutputs("o1"))
				require.NoError(t, c.Inputs().ByName("i1").PutSignals(signal.New(10)))
				require.NoError(t, c.Outputs().ByName("o1").PutSignals(signal.New(20)))
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
			c := tt.getComponent()
			require.NoError(t, c.ClearInputs())
			if tt.assertions != nil {
				tt.assertions(t, c)
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
				return mustNew("c1")
			},
			assertions: func(t *testing.T, componentAfterFlush *Component) {
				assert.NotNil(t, componentAfterFlush.Outputs())
				assert.Zero(t, componentAfterFlush.Outputs().Len())
			},
		},
		{
			name: "output has no signal set",
			getComponent: func() *Component {
				c := mustNew("c1")
				require.NoError(t, c.AddOutputs("o1", "o2"))
				return c
			},
			assertions: func(t *testing.T, componentAfterFlush *Component) {
				assert.False(t, componentAfterFlush.Outputs().AnyHasSignals())
			},
		},
		{
			name: "happy path",
			getComponent: func() *Component {
				sink, err := port.NewInput("sink")
				require.NoError(t, err)
				c := mustNew("c1")
				require.NoError(t, c.AddOutputs("o1", "o2"))
				require.NoError(t, c.Outputs().ByNames("o1").PutSignals(signal.New(777)))
				require.NoError(t, c.Outputs().ByNames("o2").PutSignals(signal.New(888)))
				require.NoError(t, c.Outputs().ByNames("o1", "o2").PipeTo(sink))
				return c
			},
			assertions: func(t *testing.T, componentAfterFlush *Component) {
				destPort := componentAfterFlush.OutputByName("o1").Pipes().First()
				allPayloads, err := destPort.Signals().AllPayloads()
				require.NoError(t, err)
				assert.Contains(t, allPayloads, 777)
				assert.Contains(t, allPayloads, 888)
				assert.Len(t, allPayloads, 2)
				// Buffer is cleared when port is flushed
				assert.False(t, componentAfterFlush.Outputs().AnyHasSignals())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := tt.getComponent()
			require.NoError(t, c.FlushOutputs())
			tt.assertions(t, c)
		})
	}
}

func TestComponent_AttachInputPorts(t *testing.T) {
	t.Run("add single input port with description", func(t *testing.T) {
		c := mustNew("c1")
		p, err := port.NewInput("in1", port.WithDescription("input port 1"))
		require.NoError(t, err)
		require.NoError(t, c.AttachInputPorts(p))
		assert.Equal(t, 1, c.Inputs().Len())
		assert.Equal(t, "input port 1", c.InputByName("in1").Description())
	})

	t.Run("add multiple input ports with descriptions and labels", func(t *testing.T) {
		c := mustNew("c1")
		p1, err := port.NewInput("in1", port.WithDescription("first input"))
		require.NoError(t, err)
		p1.AddLabel("priority", "high")
		p2, err := port.NewInput("in2", port.WithDescription("second input"))
		require.NoError(t, err)
		p2.AddLabel("priority", "low")
		require.NoError(t, c.AttachInputPorts(p1, p2))
		assert.Equal(t, 2, c.Inputs().Len())
		assert.Equal(t, "first input", c.InputByName("in1").Description())
		assert.Equal(t, "second input", c.InputByName("in2").Description())
		assert.True(t, c.InputByName("in1").Labels().ValueIs("priority", "high"))
		assert.True(t, c.InputByName("in2").Labels().ValueIs("priority", "low"))
	})

	t.Run("add ports to existing inputs", func(t *testing.T) {
		c := mustNew("c1")
		require.NoError(t, c.AddInputs("in1"))
		p, err := port.NewInput("in2", port.WithDescription("second input"))
		require.NoError(t, err)
		require.NoError(t, c.AttachInputPorts(p))
		assert.Equal(t, 2, c.Inputs().Len())
		assert.Equal(t, "second input", c.InputByName("in2").Description())
	})

	t.Run("sequential attach calls", func(t *testing.T) {
		c := mustNew("c1")
		p1, err := port.NewInput("in1", port.WithDescription("input 1"))
		require.NoError(t, err)
		require.NoError(t, c.AttachInputPorts(p1))
		p2, err := port.NewInput("in2", port.WithDescription("input 2"))
		require.NoError(t, err)
		require.NoError(t, c.AttachInputPorts(p2))
		assert.Equal(t, 2, c.Inputs().Len())
	})

	t.Run("GUARDRAIL: reject output port (wrong direction)", func(t *testing.T) {
		c := mustNew("c1")
		p, err := port.NewOutput("wrong_direction")
		require.NoError(t, err)
		attachErr := c.AttachInputPorts(p)
		require.Error(t, attachErr)
		require.ErrorIs(t, attachErr, port.ErrWrongPortDirection)
		assert.Contains(t, attachErr.Error(), "wrong_direction")
		assert.Contains(t, attachErr.Error(), "not an input port")
	})

	t.Run("GUARDRAIL: reject mixed input and output ports", func(t *testing.T) {
		c := mustNew("c1")
		p1, err := port.NewInput("correct")
		require.NoError(t, err)
		p2, err := port.NewOutput("wrong")
		require.NoError(t, err)
		attachErr := c.AttachInputPorts(p1, p2)
		require.Error(t, attachErr)
		require.ErrorIs(t, attachErr, port.ErrWrongPortDirection)
		assert.Contains(t, attachErr.Error(), "wrong")
	})
}

func TestComponent_AttachOutputPorts(t *testing.T) {
	t.Run("add single output port with description", func(t *testing.T) {
		c := mustNew("c1")
		p, err := port.NewOutput("out1", port.WithDescription("output port 1"))
		require.NoError(t, err)
		require.NoError(t, c.AttachOutputPorts(p))
		assert.Equal(t, 1, c.Outputs().Len())
		assert.Equal(t, "output port 1", c.OutputByName("out1").Description())
	})

	t.Run("add multiple output ports with descriptions and labels", func(t *testing.T) {
		c := mustNew("c1")
		p1, err := port.NewOutput("out1", port.WithDescription("first output"))
		require.NoError(t, err)
		p1.AddLabel("type", "result")
		p2, err := port.NewOutput("out2", port.WithDescription("second output"))
		require.NoError(t, err)
		p2.AddLabel("type", "error")
		require.NoError(t, c.AttachOutputPorts(p1, p2))
		assert.Equal(t, 2, c.Outputs().Len())
		assert.Equal(t, "first output", c.OutputByName("out1").Description())
		assert.Equal(t, "second output", c.OutputByName("out2").Description())
		assert.True(t, c.OutputByName("out1").Labels().ValueIs("type", "result"))
		assert.True(t, c.OutputByName("out2").Labels().ValueIs("type", "error"))
	})

	t.Run("add ports to existing outputs", func(t *testing.T) {
		c := mustNew("c1")
		require.NoError(t, c.AddOutputs("out1"))
		p, err := port.NewOutput("out2", port.WithDescription("second output"))
		require.NoError(t, err)
		require.NoError(t, c.AttachOutputPorts(p))
		assert.Equal(t, 2, c.Outputs().Len())
		assert.Equal(t, "second output", c.OutputByName("out2").Description())
	})

	t.Run("sequential attach calls", func(t *testing.T) {
		c := mustNew("c1")
		p1, err := port.NewOutput("out1", port.WithDescription("output 1"))
		require.NoError(t, err)
		require.NoError(t, c.AttachOutputPorts(p1))
		p2, err := port.NewOutput("out2", port.WithDescription("output 2"))
		require.NoError(t, err)
		require.NoError(t, c.AttachOutputPorts(p2))
		assert.Equal(t, 2, c.Outputs().Len())
	})

	t.Run("GUARDRAIL: reject input port (wrong direction)", func(t *testing.T) {
		c := mustNew("c1")
		p, err := port.NewInput("wrong_direction")
		require.NoError(t, err)
		attachErr := c.AttachOutputPorts(p)
		require.Error(t, attachErr)
		require.ErrorIs(t, attachErr, port.ErrWrongPortDirection)
		assert.Contains(t, attachErr.Error(), "wrong_direction")
		assert.Contains(t, attachErr.Error(), "not an output port")
	})

	t.Run("GUARDRAIL: reject mixed output and input ports", func(t *testing.T) {
		c := mustNew("c1")
		p1, err := port.NewOutput("correct")
		require.NoError(t, err)
		p2, err := port.NewInput("wrong")
		require.NoError(t, err)
		attachErr := c.AttachOutputPorts(p1, p2)
		require.Error(t, attachErr)
		require.ErrorIs(t, attachErr, port.ErrWrongPortDirection)
		assert.Contains(t, attachErr.Error(), "wrong")
	})
}

func TestComponent_MultipleInputOutputCalls(t *testing.T) {
	t.Run("multiple AddInputs calls add ports incrementally", func(t *testing.T) {
		c := mustNew("c1")
		require.NoError(t, c.AddInputs("in1"))
		require.NoError(t, c.AddInputs("in2", "in3"))
		require.NoError(t, c.AddInputs("in4"))

		assert.Equal(t, 4, c.Inputs().Len())
		assert.NotNil(t, c.InputByName("in1"))
		assert.NotNil(t, c.InputByName("in2"))
		assert.NotNil(t, c.InputByName("in3"))
		assert.NotNil(t, c.InputByName("in4"))
	})

	t.Run("multiple AddOutputs calls add ports incrementally", func(t *testing.T) {
		c := mustNew("c1")
		require.NoError(t, c.AddOutputs("out1"))
		require.NoError(t, c.AddOutputs("out2", "out3"))
		require.NoError(t, c.AddOutputs("out4"))

		assert.Equal(t, 4, c.Outputs().Len())
		assert.NotNil(t, c.OutputByName("out1"))
		assert.NotNil(t, c.OutputByName("out2"))
		assert.NotNil(t, c.OutputByName("out3"))
		assert.NotNil(t, c.OutputByName("out4"))
	})

	t.Run("mixing AddInputs and AttachInputPorts", func(t *testing.T) {
		c := mustNew("c1")
		require.NoError(t, c.AddInputs("in1", "in2"))
		p3, err := port.NewInput("in3", port.WithDescription("third input"))
		require.NoError(t, err)
		p4, err := port.NewInput("in4", port.WithDescription("fourth input"))
		require.NoError(t, err)
		p4.AddLabel("important", "true")
		require.NoError(t, c.AttachInputPorts(p3, p4))
		require.NoError(t, c.AddInputs("in5"))

		assert.Equal(t, 5, c.Inputs().Len())
		assert.Empty(t, c.InputByName("in1").Description())
		assert.Empty(t, c.InputByName("in2").Description())
		assert.Equal(t, "third input", c.InputByName("in3").Description())
		assert.Equal(t, "fourth input", c.InputByName("in4").Description())
		assert.True(t, c.InputByName("in4").Labels().ValueIs("important", "true"))
		assert.Empty(t, c.InputByName("in5").Description())
	})

	t.Run("mixing AddOutputs and AttachOutputPorts", func(t *testing.T) {
		c := mustNew("c1")
		require.NoError(t, c.AddOutputs("out1", "out2"))
		p3, err := port.NewOutput("out3", port.WithDescription("third output"))
		require.NoError(t, err)
		p4, err := port.NewOutput("out4", port.WithDescription("fourth output"))
		require.NoError(t, err)
		p4.AddLabel("type", "error")
		require.NoError(t, c.AttachOutputPorts(p3, p4))
		require.NoError(t, c.AddOutputs("out5"))

		assert.Equal(t, 5, c.Outputs().Len())
		assert.Empty(t, c.OutputByName("out1").Description())
		assert.Empty(t, c.OutputByName("out2").Description())
		assert.Equal(t, "third output", c.OutputByName("out3").Description())
		assert.Equal(t, "fourth output", c.OutputByName("out4").Description())
		assert.True(t, c.OutputByName("out4").Labels().ValueIs("type", "error"))
		assert.Empty(t, c.OutputByName("out5").Description())
	})
}
