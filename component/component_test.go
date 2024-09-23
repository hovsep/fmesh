package component

import (
	"errors"
	"github.com/hovsep/fmesh/port"
	"github.com/hovsep/fmesh/signal"
	"github.com/stretchr/testify/assert"
	"testing"
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
			want: &Component{
				name:        "",
				description: "",
				inputs:      port.Collection{},
				outputs:     port.Collection{},
				f:           nil,
			},
		},
		{
			name: "with name",
			args: args{
				name: "multiplier",
			},
			want: &Component{
				name:        "multiplier",
				description: "",
				inputs:      port.Collection{},
				outputs:     port.Collection{},
				f:           nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, New(tt.args.name), "New(%v)", tt.args.name)
		})
	}
}

func TestComponent_Name(t *testing.T) {
	tests := []struct {
		name      string
		component *Component
		want      string
	}{
		{
			name:      "empty name",
			component: New(""),
			want:      "",
		},
		{
			name:      "with name",
			component: New("c1"),
			want:      "c1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.component.Name())
		})
	}
}

func TestComponent_Description(t *testing.T) {
	tests := []struct {
		name      string
		component *Component
		want      string
	}{
		{
			name:      "no description",
			component: New("c1"),
			want:      "",
		},
		{
			name:      "with description",
			component: New("c1").WithDescription("descr"),
			want:      "descr",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.component.Description())
		})
	}
}

func TestComponent_FlushOutputs(t *testing.T) {
	sink := port.New("sink")

	componentWithNoOutputs := New("c1")
	componentWithCleanOutputs := New("c1").WithOutputs("o1", "o2")

	componentWithAllOutputsSet := New("c1").WithOutputs("o1", "o2")
	componentWithAllOutputsSet.Outputs().ByNames("o1").PutSignals(signal.New(777))
	componentWithAllOutputsSet.Outputs().ByNames("o2").PutSignals(signal.New(888))
	componentWithAllOutputsSet.Outputs().ByNames("o1", "o2").PipeTo(sink)

	tests := []struct {
		name             string
		component        *Component
		activationResult *ActivationResult
		destPort         *port.Port //Where the component flushes ALL it's inputs
		assertions       func(t *testing.T, componentAfterFlush *Component, destPort *port.Port)
	}{
		{
			name:      "no outputs",
			component: componentWithNoOutputs,
			activationResult: componentWithNoOutputs.newActivationResultOK().
				WithStateBefore(NewStateSnapshot()).
				WithStateAfter(NewStateSnapshot()),
			destPort: nil,
			assertions: func(t *testing.T, componentAfterFlush *Component, destPort *port.Port) {
				assert.NotNil(t, componentAfterFlush.Outputs())
				assert.Empty(t, componentAfterFlush.Outputs())
			},
		},
		{
			name:      "output has no signal set",
			component: componentWithCleanOutputs,
			activationResult: componentWithCleanOutputs.newActivationResultOK().
				WithStateBefore(NewStateSnapshot().WithOutputPortsMetadata(port.MetadataMap{
					"o1": &port.Metadata{SignalBufferLen: 0},
					"o2": &port.Metadata{SignalBufferLen: 0},
				})).
				WithStateAfter(NewStateSnapshot().WithOutputPortsMetadata(port.MetadataMap{
					"o1": &port.Metadata{SignalBufferLen: 0},
					"o2": &port.Metadata{SignalBufferLen: 0},
				})),
			destPort: nil,
			assertions: func(t *testing.T, componentAfterFlush *Component, destPort *port.Port) {
				assert.False(t, componentAfterFlush.Outputs().AnyHasSignals())
			},
		},
		{
			name:      "happy path",
			component: componentWithAllOutputsSet,
			activationResult: componentWithAllOutputsSet.newActivationResultOK().
				WithStateBefore(NewStateSnapshot()).
				WithStateAfter(NewStateSnapshot().
					WithOutputPortsMetadata(port.MetadataMap{
						"o1": &port.Metadata{
							SignalBufferLen: 1,
						},
						"o2": &port.Metadata{
							SignalBufferLen: 1,
						},
					})),
			destPort: sink,
			assertions: func(t *testing.T, componentAfterFlush *Component, destPort *port.Port) {
				assert.Contains(t, destPort.Signals().AllPayloads(), 777)
				assert.Contains(t, destPort.Signals().AllPayloads(), 888)
				assert.Len(t, destPort.Signals().AllPayloads(), 2)
				// Signals are disposed when port is flushed
				assert.False(t, componentAfterFlush.Outputs().AnyHasSignals())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.component.FlushOutputs(tt.activationResult)
			tt.assertions(t, tt.component, tt.destPort)
		})
	}
}

func TestComponent_Inputs(t *testing.T) {
	tests := []struct {
		name      string
		component *Component
		want      port.Collection
	}{
		{
			name:      "no inputs",
			component: New("c1"),
			want:      port.Collection{},
		},
		{
			name:      "with inputs",
			component: New("c1").WithInputs("i1", "i2"),
			want: port.Collection{
				"i1": port.New("i1"),
				"i2": port.New("i2"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.component.Inputs())
		})
	}
}

func TestComponent_Outputs(t *testing.T) {
	tests := []struct {
		name      string
		component *Component
		want      port.Collection
	}{
		{
			name:      "no outputs",
			component: New("c1"),
			want:      port.Collection{},
		},
		{
			name:      "with outputs",
			component: New("c1").WithOutputs("o1", "o2"),
			want: port.Collection{
				"o1": port.New("o1"),
				"o2": port.New("o2"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.component.Outputs())
		})
	}
}

func TestComponent_WithActivationFunc(t *testing.T) {
	type args struct {
		f ActivationFunc
	}
	tests := []struct {
		name      string
		component *Component
		args      args
	}{
		{
			name:      "happy path",
			component: New("c1"),
			args: args{
				f: func(inputs port.Collection, outputs port.Collection) error {
					outputs.ByName("out1").PutSignals(signal.New(23))
					return nil
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			componentAfter := tt.component.WithActivationFunc(tt.args.f)

			//Compare activation functions by they result and error
			testInputs1 := port.NewCollection().Add(port.NewGroup("in1", "in2")...)
			testInputs2 := port.NewCollection().Add(port.NewGroup("in1", "in2")...)
			testOutputs1 := port.NewCollection().Add(port.NewGroup("out1", "out2")...)
			testOutputs2 := port.NewCollection().Add(port.NewGroup("out1", "out2")...)
			err1 := componentAfter.f(testInputs1, testOutputs1)
			err2 := tt.args.f(testInputs2, testOutputs2)
			assert.Equal(t, err1, err2)

			//Compare signals without keys (because they are random)
			assert.ElementsMatch(t, testOutputs1.ByName("out1").Signals(), testOutputs2.ByName("out1").Signals())
			assert.ElementsMatch(t, testOutputs1.ByName("out2").Signals(), testOutputs2.ByName("out2").Signals())

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
				inputs:      port.Collection{},
				outputs:     port.Collection{},
				f:           nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.component.WithDescription(tt.args.description))
		})
	}
}

func TestComponent_WithInputs(t *testing.T) {
	type args struct {
		portNames []string
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
				portNames: []string{"p1", "p2"},
			},
			want: &Component{
				name:        "c1",
				description: "",
				inputs: port.Collection{
					"p1": port.New("p1"),
					"p2": port.New("p2"),
				},
				outputs: port.Collection{},
				f:       nil,
			},
		},
		{
			name:      "no arg",
			component: New("c1"),
			args: args{
				portNames: nil,
			},
			want: &Component{
				name:        "c1",
				description: "",
				inputs:      port.Collection{},
				outputs:     port.Collection{},
				f:           nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.component.WithInputs(tt.args.portNames...))
		})
	}
}

func TestComponent_WithOutputs(t *testing.T) {
	type args struct {
		portNames []string
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
				portNames: []string{"p1", "p2"},
			},
			want: &Component{
				name:        "c1",
				description: "",
				inputs:      port.Collection{},
				outputs: port.Collection{
					"p1": port.New("p1"),
					"p2": port.New("p2"),
				},
				f: nil,
			},
		},
		{
			name:      "no arg",
			component: New("c1"),
			args: args{
				portNames: nil,
			},
			want: &Component{
				name:        "c1",
				description: "",
				inputs:      port.Collection{},
				outputs:     port.Collection{},
				f:           nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.component.WithOutputs(tt.args.portNames...))
		})
	}
}

func TestComponent_MaybeActivate(t *testing.T) {
	tests := []struct {
		name                 string
		getComponent         func() *Component
		wantActivationResult *ActivationResult
	}{
		{
			name: "empty component is not activated",
			getComponent: func() *Component {
				return New("c1")
			},
			wantActivationResult: NewActivationResult("c1").SetActivated(false).WithActivationCode(ActivationCodeNoFunction),
		},
		{
			name: "component with inputs set, but no activation func",
			getComponent: func() *Component {
				c := New("c1").WithInputs("i1")
				c.Inputs().ByName("i1").PutSignals(signal.New(123))
				return c
			},
			wantActivationResult: NewActivationResult("c1").
				SetActivated(false).
				WithActivationCode(ActivationCodeNoFunction),
		},
		{
			name: "no input",
			getComponent: func() *Component {
				c := New("c1").
					WithInputs("i1", "i2").
					WithActivationFunc(func(inputs port.Collection, outputs port.Collection) error {

						if !inputs.ByNames("i1", "i2").AllHaveSignals() {
							return NewErrWaitForInputs(false)
						}

						return nil
					})
				return c
			},
			wantActivationResult: NewActivationResult("c1").
				SetActivated(false).
				WithActivationCode(ActivationCodeNoInput),
		},
		{
			name: "component is waiting for input, reset inputs",
			getComponent: func() *Component {
				c := New("c1").
					WithInputs("i1", "i2").
					WithActivationFunc(func(inputs port.Collection, outputs port.Collection) error {

						if !inputs.ByNames("i1", "i2").AllHaveSignals() {
							return NewErrWaitForInputs(false)
						}

						return nil
					})
				//Only one input set
				c.Inputs().ByName("i1").PutSignals(signal.New(123))
				return c
			},
			wantActivationResult: NewActivationResult("c1").
				SetActivated(false).
				WithActivationCode(ActivationCodeWaitingForInput),
		},
		{
			name: "component is waiting for input, keep inputs",
			getComponent: func() *Component {
				c := New("c1").
					WithInputs("i1", "i2").
					WithActivationFunc(func(inputs port.Collection, outputs port.Collection) error {

						if !inputs.ByNames("i1", "i2").AllHaveSignals() {
							return NewErrWaitForInputs(true)
						}

						return nil
					})
				//Only one input set
				c.Inputs().ByName("i1").PutSignals(signal.New(123))
				return c
			},
			wantActivationResult: NewActivationResult("c1").
				SetActivated(false).
				WithActivationCode(ActivationCodeWaitingForInput),
		},
		{
			name: "activated with error",
			getComponent: func() *Component {
				c := New("c1").
					WithInputs("i1").
					WithActivationFunc(func(inputs port.Collection, outputs port.Collection) error {
						return errors.New("test error")
					})
				//Only one input set
				c.Inputs().ByName("i1").PutSignals(signal.New(123))
				return c
			},
			wantActivationResult: NewActivationResult("c1").
				SetActivated(true).
				WithActivationCode(ActivationCodeReturnedError).
				WithError(errors.New("component returned an error: test error")),
		},
		{
			name: "activated without error",
			getComponent: func() *Component {
				c := New("c1").
					WithInputs("i1").
					WithOutputs("o1").
					WithActivationFunc(func(inputs port.Collection, outputs port.Collection) error {
						port.ForwardSignals(inputs.ByName("i1"), outputs.ByName("o1"))
						return nil
					})
				//Only one input set
				c.Inputs().ByName("i1").PutSignals(signal.New(123))
				return c
			},
			wantActivationResult: NewActivationResult("c1").
				SetActivated(true).
				WithActivationCode(ActivationCodeOK),
		},
		{
			name: "component panicked with error",
			getComponent: func() *Component {
				c := New("c1").
					WithInputs("i1").
					WithOutputs("o1").
					WithActivationFunc(func(inputs port.Collection, outputs port.Collection) error {
						port.ForwardSignals(inputs.ByName("i1"), outputs.ByName("o1"))
						panic(errors.New("oh shrimps"))
						return nil
					})
				//Only one input set
				c.Inputs().ByName("i1").PutSignals(signal.New(123))
				return c
			},
			wantActivationResult: NewActivationResult("c1").
				SetActivated(true).
				WithActivationCode(ActivationCodePanicked).
				WithError(errors.New("panicked with: oh shrimps")),
		},
		{
			name: "component panicked with string",
			getComponent: func() *Component {
				c := New("c1").
					WithInputs("i1").
					WithOutputs("o1").
					WithActivationFunc(func(inputs port.Collection, outputs port.Collection) error {
						port.ForwardSignals(inputs.ByName("i1"), outputs.ByName("o1"))
						panic("oh shrimps")
						return nil
					})
				//Only one input set
				c.Inputs().ByName("i1").PutSignals(signal.New(123))
				return c
			},
			wantActivationResult: NewActivationResult("c1").
				SetActivated(true).
				WithActivationCode(ActivationCodePanicked).
				WithError(errors.New("panicked with: oh shrimps")),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotActivationResult := tt.getComponent().MaybeActivate()
			assert.Equal(t, gotActivationResult.Activated(), tt.wantActivationResult.Activated())
			assert.Equal(t, gotActivationResult.ComponentName(), tt.wantActivationResult.ComponentName())
			assert.Equal(t, gotActivationResult.Code(), tt.wantActivationResult.Code())
			if tt.wantActivationResult.HasError() {
				assert.EqualError(t, gotActivationResult.Error(), tt.wantActivationResult.Error().Error())
			} else {
				assert.False(t, gotActivationResult.HasError())
			}

		})
	}
}
