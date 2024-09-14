package component

import (
	"errors"
	"github.com/hovsep/fmesh/cycle"
	"github.com/hovsep/fmesh/port"
	"github.com/hovsep/fmesh/signal"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestComponent_Description(t *testing.T) {
	tests := []struct {
		name      string
		component *Component
		want      string
	}{
		{
			name:      "no description",
			component: NewComponent("c1"),
			want:      "",
		},
		{
			name:      "with description",
			component: NewComponent("c1").WithDescription("descr"),
			want:      "descr",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.component.Description(); got != tt.want {
				t.Errorf("Description() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestComponent_FlushOutputs(t *testing.T) {
	sink := port.NewPort("sink")

	componentWithAllOutputsSet := NewComponent("c1").WithOutputs("o1", "o2")
	componentWithAllOutputsSet.Outputs().ByName("o1").PutSignal(signal.New(777))
	componentWithAllOutputsSet.Outputs().ByName("o1").PutSignal(signal.New(888))
	componentWithAllOutputsSet.Outputs().ByName("o1").PipeTo(sink)
	componentWithAllOutputsSet.Outputs().ByName("o2").PipeTo(sink)

	tests := []struct {
		name       string
		component  *Component
		destPort   *port.Port //Where the component flushes ALL it's inputs
		assertions func(t *testing.T, componentAfterFlush *Component, destPort *port.Port)
	}{
		{
			name:      "no outputs",
			component: NewComponent("c1"),
			destPort:  nil,
			assertions: func(t *testing.T, componentAfterFlush *Component, destPort *port.Port) {
				assert.Nil(t, componentAfterFlush.Outputs())
			},
		},
		{
			name:      "output has no signal set",
			component: NewComponent("c1").WithOutputs("o1", "o2"),
			destPort:  nil,
			assertions: func(t *testing.T, componentAfterFlush *Component, destPort *port.Port) {
				assert.False(t, componentAfterFlush.Outputs().AnyHasSignal())
			},
		},
		{
			name:      "happy path",
			component: componentWithAllOutputsSet,
			destPort:  sink,
			assertions: func(t *testing.T, componentAfterFlush *Component, destPort *port.Port) {
				assert.Contains(t, destPort.Signal().Payloads(), 777)
				assert.Contains(t, destPort.Signal().Payloads(), 888)
				assert.Len(t, destPort.Signal().Payloads(), 2)
				assert.False(t, componentAfterFlush.Outputs().AnyHasSignal())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.component.FlushOutputs()
			tt.assertions(t, tt.component, tt.destPort)
		})
	}
}

func TestComponent_Inputs(t *testing.T) {
	tests := []struct {
		name      string
		component *Component
		want      port.Ports
	}{
		{
			name:      "no inputs",
			component: NewComponent("c1"),
			want:      nil,
		},
		{
			name:      "with inputs",
			component: NewComponent("c1").WithInputs("i1", "i2"),
			want: port.Ports{
				"i1": port.NewPort("i1"),
				"i2": port.NewPort("i2"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.component.Inputs(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Inputs() = %v, want %v", got, tt.want)
			}
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
			component: NewComponent(""),
			want:      "",
		},
		{
			name:      "with name",
			component: NewComponent("c1"),
			want:      "c1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.component.Name(); got != tt.want {
				t.Errorf("Name() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestComponent_Outputs(t *testing.T) {
	tests := []struct {
		name      string
		component *Component
		want      port.Ports
	}{
		{
			name:      "no outputs",
			component: NewComponent("c1"),
			want:      nil,
		},
		{
			name:      "with outputs",
			component: NewComponent("c1").WithOutputs("o1", "o2"),
			want: port.Ports{
				"o1": port.NewPort("o1"),
				"o2": port.NewPort("o2"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.component.Outputs(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Outputs() = %v, want %v", got, tt.want)
			}
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
		want      *Component
	}{
		{
			name:      "happy path",
			component: NewComponent("c1"),
			args: args{
				f: func(inputs port.Ports, outputs port.Ports) error {
					outputs.ByName("out1").PutSignal(signal.New(23))
					return nil
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			componentAfter := tt.component.WithActivationFunc(tt.args.f)

			//Compare activation functions by they result and error
			testInputs1, testInputs2 := port.NewPorts("in1", "in2"), port.NewPorts("in1", "in2")
			testOutputs1, testOutputs2 := port.NewPorts("out1", "out2"), port.NewPorts("out1", "out2")
			err1 := componentAfter.f(testInputs1, testOutputs1)
			err2 := tt.args.f(testInputs2, testOutputs2)
			assert.Equal(t, err1, err2)
			assert.Equal(t, testOutputs1, testOutputs2)
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
			component: NewComponent("c1"),
			args: args{
				description: "descr",
			},
			want: &Component{
				name:        "c1",
				description: "descr",
				inputs:      nil,
				outputs:     nil,
				f:           nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.component.WithDescription(tt.args.description); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WithDescription() = %v, want %v", got, tt.want)
			}
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
			component: NewComponent("c1"),
			args: args{
				portNames: []string{"p1", "p2"},
			},
			want: &Component{
				name:        "c1",
				description: "",
				inputs: port.Ports{
					"p1": port.NewPort("p1"),
					"p2": port.NewPort("p2"),
				},
				outputs: nil,
				f:       nil,
			},
		},
		{
			name:      "no arg",
			component: NewComponent("c1"),
			args: args{
				portNames: nil,
			},
			want: &Component{
				name:        "c1",
				description: "",
				inputs:      port.Ports{},
				outputs:     nil,
				f:           nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.component.WithInputs(tt.args.portNames...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WithInputs() = %v, want %v", got, tt.want)
			}
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
			component: NewComponent("c1"),
			args: args{
				portNames: []string{"p1", "p2"},
			},
			want: &Component{
				name:        "c1",
				description: "",
				inputs:      nil,
				outputs: port.Ports{
					"p1": port.NewPort("p1"),
					"p2": port.NewPort("p2"),
				},
				f: nil,
			},
		},
		{
			name:      "no arg",
			component: NewComponent("c1"),
			args: args{
				portNames: nil,
			},
			want: &Component{
				name:        "c1",
				description: "",
				inputs:      nil,
				outputs:     port.Ports{},
				f:           nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.component.WithOutputs(tt.args.portNames...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WithOutputs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestComponents_ByName(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name       string
		components Components
		args       args
		want       *Component
	}{
		{
			name:       "component found",
			components: NewComponents("c1", "c2"),
			args: args{
				name: "c2",
			},
			want: &Component{
				name:        "c2",
				description: "",
				inputs:      nil,
				outputs:     nil,
				f:           nil,
			},
		},
		{
			name:       "component not found",
			components: NewComponents("c1", "c2"),
			args: args{
				name: "c3",
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.components.ByName(tt.args.name); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ByName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNew(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name string
		args args
		want *Component
	}{
		{
			name: "happy path",
			args: args{
				name: "c1",
			},
			want: &Component{
				name:        "c1",
				description: "",
				inputs:      nil,
				outputs:     nil,
				f:           nil,
			},
		},
		{
			name: "empty name is valid",
			args: args{
				name: "",
			},
			want: &Component{
				name:        "",
				description: "",
				inputs:      nil,
				outputs:     nil,
				f:           nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewComponent(tt.args.name); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewComponent() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestComponent_Activate(t *testing.T) {
	tests := []struct {
		name         string
		getComponent func() *Component
		wantARes     cycle.ActivationResult
	}{
		{
			name: "empty component is not activated",
			getComponent: func() *Component {
				return NewComponent("c1")
			},
			wantARes: cycle.ActivationResult{
				Activated:       false,
				ComponentName:   "c1",
				ActivationError: nil,
			},
		},
		{
			name: "component with inputs set, but no activation func",
			getComponent: func() *Component {
				c := NewComponent("c1").WithInputs("i1")
				c.Inputs().ByName("i1").PutSignal(signal.New(123))
				return c
			},
			wantARes: cycle.ActivationResult{
				Activated:       false,
				ComponentName:   "c1",
				ActivationError: nil,
			},
		},
		{
			name: "component is waiting for input",
			getComponent: func() *Component {
				c := NewComponent("c1").
					WithInputs("i1", "i2").
					WithActivationFunc(func(inputs port.Ports, outputs port.Ports) error {

						if !inputs.ByNames("i1", "i2").AllHaveSignal() {
							return ErrWaitingForInputResetInputs
						}

						return nil
					})
				//Only one input set
				c.Inputs().ByName("i1").PutSignal(signal.New(123))
				return c
			},
			wantARes: cycle.ActivationResult{
				Activated:       false,
				ComponentName:   "c1",
				ActivationError: nil,
			},
		},
		{
			name: "activated with error",
			getComponent: func() *Component {
				c := NewComponent("c1").
					WithInputs("i1").
					WithActivationFunc(func(inputs port.Ports, outputs port.Ports) error {
						return errors.New("test error")
					})
				//Only one input set
				c.Inputs().ByName("i1").PutSignal(signal.New(123))
				return c
			},
			wantARes: cycle.ActivationResult{
				Activated:       true,
				ComponentName:   "c1",
				ActivationError: errors.New("failed to activate component: test error"),
			},
		},
		{
			name: "activated without error",
			getComponent: func() *Component {
				c := NewComponent("c1").
					WithInputs("i1").
					WithOutputs("o1").
					WithActivationFunc(func(inputs port.Ports, outputs port.Ports) error {
						port.ForwardSignal(inputs.ByName("i1"), outputs.ByName("o1"))
						return nil
					})
				//Only one input set
				c.Inputs().ByName("i1").PutSignal(signal.New(123))
				return c
			},
			wantARes: cycle.ActivationResult{
				Activated:       true,
				ComponentName:   "c1",
				ActivationError: nil,
			},
		},
		{
			name: "component panicked with error",
			getComponent: func() *Component {
				c := NewComponent("c1").
					WithInputs("i1").
					WithOutputs("o1").
					WithActivationFunc(func(inputs port.Ports, outputs port.Ports) error {
						port.ForwardSignal(inputs.ByName("i1"), outputs.ByName("o1"))
						panic(errors.New("oh shrimps"))
						return nil
					})
				//Only one input set
				c.Inputs().ByName("i1").PutSignal(signal.New(123))
				return c
			},
			wantARes: cycle.ActivationResult{
				Activated:       true,
				ComponentName:   "c1",
				ActivationError: errors.New("panicked with: oh shrimps"),
			},
		},
		{
			name: "component panicked with string",
			getComponent: func() *Component {
				c := NewComponent("c1").
					WithInputs("i1").
					WithOutputs("o1").
					WithActivationFunc(func(inputs port.Ports, outputs port.Ports) error {
						port.ForwardSignal(inputs.ByName("i1"), outputs.ByName("o1"))
						panic("oh shrimps")
						return nil
					})
				//Only one input set
				c.Inputs().ByName("i1").PutSignal(signal.New(123))
				return c
			},
			wantARes: cycle.ActivationResult{
				Activated:       true,
				ComponentName:   "c1",
				ActivationError: errors.New("panicked with: oh shrimps"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.getComponent().MaybeActivate()
			assert.Equal(t, got.Activated, tt.wantARes.Activated)
			assert.Equal(t, got.ComponentName, tt.wantARes.ComponentName)
			if tt.wantARes.ActivationError == nil {
				assert.Nil(t, got.ActivationError)
			} else {
				assert.ErrorContains(t, got.ActivationError, tt.wantARes.ActivationError.Error())
			}

		})
	}
}

func TestComponents_Add(t *testing.T) {
	type args struct {
		component *Component
	}
	tests := []struct {
		name       string
		components Components
		args       args
		want       Components
	}{
		{
			name:       "adding to empty collection",
			components: NewComponents(),
			args: args{
				component: NewComponent("c1").WithDescription("descr"),
			},
			want: Components{
				"c1": {name: "c1", description: "descr"},
			},
		},
		{
			name:       "adding to existing collection",
			components: NewComponents("c1", "c2"),
			args: args{
				component: NewComponent("c3").WithDescription("descr"),
			},
			want: Components{
				"c1": {name: "c1"},
				"c2": {name: "c2"},
				"c3": {name: "c3", description: "descr"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, tt.components.Add(tt.args.component), "Add(%v)", tt.args.component)
		})
	}
}

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
			want: &Component{},
		},
		{
			name: "with name",
			args: args{
				name: "multiplier",
			},
			want: &Component{
				name: "multiplier",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, NewComponent(tt.args.name), "NewComponent(%v)", tt.args.name)
		})
	}
}

func TestNewComponents(t *testing.T) {
	type args struct {
		names []string
	}
	tests := []struct {
		name string
		args args
		want Components
	}{
		{
			name: "without specifying names",
			args: args{
				names: nil,
			},
			want: Components{},
		},
		{
			name: "with names",
			args: args{
				names: []string{"c1", "c2"},
			},
			want: Components{
				"c1": {name: "c1"},
				"c2": {name: "c2"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, NewComponents(tt.args.names...), "NewComponents(%v)", tt.args.names)
		})
	}
}
