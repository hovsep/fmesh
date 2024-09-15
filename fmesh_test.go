package fmesh

import (
	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/cycle"
	"github.com/hovsep/fmesh/port"
	"github.com/hovsep/fmesh/signal"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestNew(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name string
		args args
		want *FMesh
	}{
		{
			name: "empty name is valid",
			args: args{
				name: "",
			},
			want: &FMesh{
				components: component.Components{},
			},
		},
		{
			name: "with name",
			args: args{
				name: "fm1",
			},
			want: &FMesh{
				name:       "fm1",
				components: component.Components{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := New(tt.args.name); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFMesh_WithDescription(t *testing.T) {
	type args struct {
		description string
	}
	tests := []struct {
		name string
		fm   *FMesh
		args args
		want *FMesh
	}{
		{
			name: "empty description",
			fm:   New("fm1"),
			args: args{
				description: "",
			},
			want: &FMesh{
				name:                  "fm1",
				description:           "",
				components:            component.Components{},
				errorHandlingStrategy: 0,
			},
		},
		{
			name: "with description",
			fm:   New("fm1"),
			args: args{
				description: "descr",
			},
			want: &FMesh{
				name:                  "fm1",
				description:           "descr",
				components:            component.Components{},
				errorHandlingStrategy: 0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.fm.WithDescription(tt.args.description); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WithDescription() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFMesh_WithErrorHandlingStrategy(t *testing.T) {
	type args struct {
		strategy ErrorHandlingStrategy
	}
	tests := []struct {
		name string
		fm   *FMesh
		args args
		want *FMesh
	}{
		{
			name: "default strategy",
			fm:   New("fm1"),
			args: args{
				strategy: 0,
			},
			want: &FMesh{
				name:                  "fm1",
				components:            component.Components{},
				errorHandlingStrategy: StopOnFirstError,
			},
		},
		{
			name: "custom strategy",
			fm:   New("fm1"),
			args: args{
				strategy: IgnoreAll,
			},
			want: &FMesh{
				name:                  "fm1",
				components:            component.Components{},
				errorHandlingStrategy: IgnoreAll,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.fm.WithErrorHandlingStrategy(tt.args.strategy); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WithErrorHandlingStrategy() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFMesh_WithComponents(t *testing.T) {
	type args struct {
		components []*component.Component
	}
	tests := []struct {
		name string
		fm   *FMesh
		args args
		want *FMesh
	}{
		{
			name: "no components",
			fm:   New("fm1"),
			args: args{
				components: nil,
			},
			want: &FMesh{
				name:                  "fm1",
				description:           "",
				components:            component.Components{},
				errorHandlingStrategy: 0,
			},
		},
		{
			name: "with single component",
			fm:   New("fm1"),
			args: args{
				components: []*component.Component{
					component.NewComponent("c1"),
				},
			},
			want: &FMesh{
				name: "fm1",
				components: component.Components{
					"c1": component.NewComponent("c1"),
				},
			},
		},
		{
			name: "with multiple components",
			fm:   New("fm1"),
			args: args{
				components: []*component.Component{
					component.NewComponent("c1"),
					component.NewComponent("c2"),
				},
			},
			want: &FMesh{
				name: "fm1",
				components: component.Components{
					"c1": component.NewComponent("c1"),
					"c2": component.NewComponent("c2"),
				},
			},
		},
		{
			name: "components with duplicating name are collapsed",
			fm:   New("fm1"),
			args: args{
				components: []*component.Component{
					component.NewComponent("c1").WithDescription("descr1"),
					component.NewComponent("c2").WithDescription("descr2"),
					component.NewComponent("c2").WithDescription("descr3"), //This will overwrite the previous one
					component.NewComponent("c4").WithDescription("descr4"),
				},
			},
			want: &FMesh{
				name: "fm1",
				components: component.Components{
					"c1": component.NewComponent("c1").WithDescription("descr1"),
					"c2": component.NewComponent("c2").WithDescription("descr3"),
					"c4": component.NewComponent("c4").WithDescription("descr4"),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.fm.WithComponents(tt.args.components...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WithComponents() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFMesh_Name(t *testing.T) {
	tests := []struct {
		name string
		fm   *FMesh
		want string
	}{
		{
			name: "empty name is valid",
			fm:   New(""),
			want: "",
		},
		{
			name: "with name",
			fm:   New("fm1"),
			want: "fm1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.fm.Name(); got != tt.want {
				t.Errorf("Name() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFMesh_Description(t *testing.T) {
	tests := []struct {
		name string
		fm   *FMesh
		want string
	}{
		{
			name: "empty description",
			fm:   New("fm1"),
			want: "",
		},
		{
			name: "with description",
			fm:   New("fm1").WithDescription("descr"),
			want: "descr",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.fm.Description(); got != tt.want {
				t.Errorf("Description() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFMesh_Run(t *testing.T) {
	type fields struct {
		name                  string
		description           string
		components            component.Components
		errorHandlingStrategy ErrorHandlingStrategy
	}
	tests := []struct {
		name    string
		fields  fields
		want    []*cycle.Result
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fm := &FMesh{
				name:                  tt.fields.name,
				description:           tt.fields.description,
				components:            tt.fields.components,
				errorHandlingStrategy: tt.fields.errorHandlingStrategy,
			}
			got, err := fm.Run()
			if (err != nil) != tt.wantErr {
				t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Run() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFMesh_runCycle(t *testing.T) {
	tests := []struct {
		name   string
		fm     *FMesh
		initFM func(fm *FMesh)
		want   *cycle.Result
	}{
		{
			name: "empty mesh",
			fm:   New("empty mesh"),
			want: cycle.NewResult(),
		},
		{
			name: "mesh has components, but no one is activated",
			fm: New("test").WithComponents(
				component.NewComponent("c1").
					WithDescription("I do not have any input signal set, hence I will never be activated").
					WithInputs("i1").
					WithOutputs("o1").
					WithActivationFunc(func(inputs port.Ports, outputs port.Ports) error {
						outputs.ByName("o1").PutSignal(signal.New("this signal will never be sent"))
						return nil
					}),

				component.NewComponent("c2").
					WithDescription("I do not have activation func set").
					WithInputs("i1").
					WithOutputs("o1"),

				component.NewComponent("c3").
					WithDescription("I'm waiting for specific input").
					WithInputs("i1", "i2").
					WithOutputs("o1").
					WithActivationFunc(func(inputs port.Ports, outputs port.Ports) error {
						if !inputs.ByNames("i1", "i2").AllHaveSignal() {
							return component.ErrWaitingForInputKeepInputs
						}
						return nil
					}),
			),
			initFM: func(fm *FMesh) {
				//Only i1 is set, while component is waiting for both i1 and i2 to be set
				fm.Components().ByName("c3").Inputs().ByName("i1").PutSignal(signal.New(123))
			},
			want: cycle.NewResult().
				WithActivationResults(
					component.NewActivationResult("c1").SetActivated(false).WithActivationCode(component.ActivationCodeNoInput),
					component.NewActivationResult("c2").SetActivated(false).WithActivationCode(component.ActivationCodeNoFunction),
					component.NewActivationResult("c3").SetActivated(false).WithActivationCode(component.ActivationCodeWaitingForInput)),
		},
		{
			name: "all components activated in one cycle (concurrently)",
			fm: New("test").WithComponents(
				component.NewComponent("c1").WithDescription("").WithInputs("i1").WithActivationFunc(func(inputs port.Ports, outputs port.Ports) error {
					return nil
				}),
				component.NewComponent("c2").WithDescription("").WithInputs("i1").WithActivationFunc(func(inputs port.Ports, outputs port.Ports) error {
					return nil
				}),
				component.NewComponent("c3").WithDescription("").WithInputs("i1").WithActivationFunc(func(inputs port.Ports, outputs port.Ports) error {
					return nil
				}),
			),
			initFM: func(fm *FMesh) {
				fm.Components().ByName("c1").Inputs().ByName("i1").PutSignal(signal.New(1))
				fm.Components().ByName("c2").Inputs().ByName("i1").PutSignal(signal.New(2))
				fm.Components().ByName("c3").Inputs().ByName("i1").PutSignal(signal.New(3))
			},
			want: cycle.NewResult().WithActivationResults(
				component.NewActivationResult("c1").SetActivated(true).WithActivationCode(component.ActivationCodeOK),
				component.NewActivationResult("c2").SetActivated(true).WithActivationCode(component.ActivationCodeOK),
				component.NewActivationResult("c3").SetActivated(true).WithActivationCode(component.ActivationCodeOK),
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.initFM != nil {
				tt.initFM(tt.fm)
			}
			if got := tt.fm.runCycle(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("runCycle() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFMesh_drainComponents(t *testing.T) {
	tests := []struct {
		name                 string
		fm                   *FMesh
		initFM               func(fm *FMesh)
		assertionsAfterDrain func(t *testing.T, fm *FMesh)
	}{
		{
			name: "no components",
			fm:   New("empty_fm"),
			assertionsAfterDrain: func(t *testing.T, fm *FMesh) {
				assert.Empty(t, fm.Components())
			},
		},
		{
			name: "no signals to be drained",
			fm: New("fm").WithComponents(
				component.NewComponent("c1").WithInputs("i1").WithOutputs("o1"),
				component.NewComponent("c2").WithInputs("i1").WithOutputs("o1"),
			),
			initFM: func(fm *FMesh) {
				//Create a pipe
				c1, c2 := fm.Components().ByName("c1"), fm.Components().ByName("c2")
				c1.Outputs().ByName("o1").PipeTo(c2.Inputs().ByName("i1"))
			},
			assertionsAfterDrain: func(t *testing.T, fm *FMesh) {
				//All ports in all components are empty
				assert.False(t, fm.Components().ByName("c1").Inputs().AnyHasSignal())
				assert.False(t, fm.Components().ByName("c1").Outputs().AnyHasSignal())
				assert.False(t, fm.Components().ByName("c2").Inputs().AnyHasSignal())
				assert.False(t, fm.Components().ByName("c2").Outputs().AnyHasSignal())
			},
		},
		{
			name: "there are signals on output, but no pipes",
			fm: New("fm").WithComponents(
				component.NewComponent("c1").WithInputs("i1").WithOutputs("o1"),
				component.NewComponent("c2").WithInputs("i1").WithOutputs("o1"),
			),
			initFM: func(fm *FMesh) {
				//Both components have signals on their outputs
				fm.Components().ByName("c1").Outputs().ByName("o1").PutSignal(signal.New(1))
				fm.Components().ByName("c2").Outputs().ByName("o1").PutSignal(signal.New(1))
			},
			assertionsAfterDrain: func(t *testing.T, fm *FMesh) {
				//Output signals are still there
				assert.True(t, fm.Components().ByName("c1").Outputs().ByName("o1").HasSignal())
				assert.True(t, fm.Components().ByName("c2").Outputs().ByName("o1").HasSignal())

				//Inputs are clear
				assert.False(t, fm.Components().ByName("c1").Inputs().ByName("i1").HasSignal())
				assert.False(t, fm.Components().ByName("c2").Inputs().ByName("i1").HasSignal())
			},
		},
		{
			name: "happy path",
			fm: New("fm").WithComponents(
				component.NewComponent("c1").WithInputs("i1").WithOutputs("o1"),
				component.NewComponent("c2").WithInputs("i1").WithOutputs("o1"),
			),
			initFM: func(fm *FMesh) {
				//Create a pipe
				c1, c2 := fm.Components().ByName("c1"), fm.Components().ByName("c2")
				c1.Outputs().ByName("o1").PipeTo(c2.Inputs().ByName("i1"))

				//c1 has a signal which must go to c2.i1 after drain
				c1.Outputs().ByName("o1").PutSignal(signal.New(123))
			},
			assertionsAfterDrain: func(t *testing.T, fm *FMesh) {
				c1, c2 := fm.Components().ByName("c1"), fm.Components().ByName("c2")

				assert.True(t, c2.Inputs().ByName("i1").HasSignal())                    //Signal is transferred to destination port
				assert.False(t, c1.Outputs().ByName("o1").HasSignal())                  //Source port is cleaned up
				assert.Equal(t, c2.Inputs().ByName("i1").Signal().Payload().(int), 123) //The signal is correct
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.initFM != nil {
				tt.initFM(tt.fm)
			}
			tt.fm.drainComponents()
			tt.assertionsAfterDrain(t, tt.fm)
		})
	}
}
