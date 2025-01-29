package component

import (
	"bytes"
	"errors"
	"github.com/hovsep/fmesh/common"
	"github.com/hovsep/fmesh/port"
	"github.com/hovsep/fmesh/signal"
	"github.com/stretchr/testify/assert"
	"log"
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
				sink := port.New("sink").WithLabels(common.LabelsCollection{
					port.DirectionLabel: port.DirectionIn,
				})
				c := New("c1").WithOutputs("o1", "o2")
				c.Outputs().ByNames("o1").PutSignals(signal.New(777))
				c.Outputs().ByNames("o2").PutSignals(signal.New(888))
				c.Outputs().ByNames("o1", "o2").PipeTo(sink)
				return c
			},
			assertions: func(t *testing.T, componentAfterFlush *Component) {
				destPort := componentAfterFlush.OutputByName("o1").Pipes().PortsOrNil()[0]
				allPayloads, err := destPort.AllSignalsPayloads()
				assert.NoError(t, err)
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
				c := New("c").WithOutputs("o1").WithErr(errors.New("some error"))
				//Lines below are ignored as error immediately propagates up to component level
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

func TestComponent_Inputs(t *testing.T) {
	tests := []struct {
		name      string
		component *Component
		want      *port.Collection
	}{
		{
			name:      "no inputs",
			component: New("c1"),
			want: port.NewCollection().WithDefaultLabels(common.LabelsCollection{
				port.DirectionLabel: port.DirectionIn,
			}),
		},
		{
			name:      "with inputs",
			component: New("c1").WithInputs("i1", "i2"),
			want: port.NewCollection().WithDefaultLabels(common.LabelsCollection{
				port.DirectionLabel: port.DirectionIn,
			}).With(port.New("i1"), port.New("i2")),
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
		want      *port.Collection
	}{
		{
			name:      "no outputs",
			component: New("c1"),
			want: port.NewCollection().WithDefaultLabels(common.LabelsCollection{
				port.DirectionLabel: port.DirectionOut,
			}),
		},
		{
			name:      "with outputs",
			component: New("c1").WithOutputs("o1", "o2"),
			want: port.NewCollection().WithDefaultLabels(common.LabelsCollection{
				port.DirectionLabel: port.DirectionOut,
			}).With(port.New("o1"), port.New("o2")),
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
				f: func(this *Component) error {
					this.OutputByName("out1").PutSignals(signal.New(23))
					return nil
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			componentAfter := tt.component.WithActivationFunc(tt.args.f)

			//Compare activation functions by they result and error
			dummyComponent1 := New("c1").WithInputs("i1", "i2").WithOutputs("o1", "o2")
			dummyComponent2 := New("c2").WithInputs("i1", "i2").WithOutputs("o1", "o2")
			err1 := componentAfter.f(dummyComponent1)
			err2 := tt.args.f(dummyComponent2)
			assert.Equal(t, err1, err2)

			//Compare signals without keys (because they are random)
			assert.ElementsMatch(t, dummyComponent1.OutputByName("o1").AllSignalsOrNil(), dummyComponent2.OutputByName("o1").AllSignalsOrNil())
			assert.ElementsMatch(t, dummyComponent1.OutputByName("o2").AllSignalsOrNil(), dummyComponent2.OutputByName("o2").AllSignalsOrNil())

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
				NamedEntity:     common.NewNamedEntity("c1"),
				DescribedEntity: common.NewDescribedEntity("descr"),
				LabeledEntity:   common.NewLabeledEntity(nil),
				Chainable:       common.NewChainable(),
				inputs: port.NewCollection().WithDefaultLabels(common.LabelsCollection{
					port.DirectionLabel: port.DirectionIn,
				}),
				outputs: port.NewCollection().WithDefaultLabels(common.LabelsCollection{
					port.DirectionLabel: port.DirectionOut,
				}),
				f: nil,
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
				NamedEntity:     common.NewNamedEntity("c1"),
				DescribedEntity: common.NewDescribedEntity(""),
				LabeledEntity:   common.NewLabeledEntity(nil),
				Chainable:       common.NewChainable(),
				inputs: port.NewCollection().WithDefaultLabels(common.LabelsCollection{
					port.DirectionLabel: port.DirectionIn,
				}).With(port.New("p1"), port.New("p2")),
				outputs: port.NewCollection().WithDefaultLabels(common.LabelsCollection{
					port.DirectionLabel: port.DirectionOut,
				}),
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
				NamedEntity:     common.NewNamedEntity("c1"),
				DescribedEntity: common.NewDescribedEntity(""),
				LabeledEntity:   common.NewLabeledEntity(nil),
				Chainable:       common.NewChainable(),
				inputs: port.NewCollection().WithDefaultLabels(common.LabelsCollection{
					port.DirectionLabel: port.DirectionIn,
				}),
				outputs: port.NewCollection().WithDefaultLabels(common.LabelsCollection{
					port.DirectionLabel: port.DirectionOut,
				}),
				f: nil,
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
				NamedEntity:     common.NewNamedEntity("c1"),
				DescribedEntity: common.NewDescribedEntity(""),
				LabeledEntity:   common.NewLabeledEntity(nil),
				Chainable:       common.NewChainable(),
				inputs: port.NewCollection().WithDefaultLabels(common.LabelsCollection{
					port.DirectionLabel: port.DirectionIn,
				}),
				outputs: port.NewCollection().WithDefaultLabels(common.LabelsCollection{
					port.DirectionLabel: port.DirectionOut,
				}).With(port.New("p1"), port.New("p2")),
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
				NamedEntity:     common.NewNamedEntity("c1"),
				DescribedEntity: common.NewDescribedEntity(""),
				LabeledEntity:   common.NewLabeledEntity(nil),
				Chainable:       common.NewChainable(),
				inputs: port.NewCollection().WithDefaultLabels(common.LabelsCollection{
					port.DirectionLabel: port.DirectionIn,
				}),
				outputs: port.NewCollection().WithDefaultLabels(common.LabelsCollection{
					port.DirectionLabel: port.DirectionOut,
				}),
				f: nil,
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
		loggerAssertions     func(t *testing.T, output []byte)
	}{
		{
			name: "component with no activation function and no inputs",
			getComponent: func() *Component {
				return New("c1")
			},
			wantActivationResult: NewActivationResult("c1").SetActivated(false).WithActivationCode(ActivationCodeNoFunction),
		},
		{
			name: "component with inputs set, but no activation func",
			getComponent: func() *Component {
				c := New("c1").WithInputs("i1")
				c.InputByName("i1").PutSignals(signal.New(123))
				return c
			},
			wantActivationResult: NewActivationResult("c1").
				SetActivated(false).
				WithActivationCode(ActivationCodeNoFunction),
		},
		{
			name: "component with activation func, but no inputs",
			getComponent: func() *Component {
				c := New("c1").
					WithInputs("i1").
					WithOutputs("o1").
					WithActivationFunc(func(this *Component) error {
						return port.ForwardSignals(this.InputByName("i1"), this.OutputByName("o1"))
					})
				return c
			},
			wantActivationResult: NewActivationResult("c1").
				SetActivated(false).
				WithActivationCode(ActivationCodeNoInput),
		},
		{
			name: "activated with error",
			getComponent: func() *Component {
				c := New("c1").
					WithInputs("i1").
					WithActivationFunc(func(this *Component) error {
						return errors.New("test error")
					})
				//Only one input set
				c.InputByName("i1").PutSignals(signal.New(123))
				return c
			},
			wantActivationResult: NewActivationResult("c1").
				SetActivated(true).
				WithActivationCode(ActivationCodeReturnedError).
				WithActivationError(errors.New("component returned an error: test error")),
		},
		{
			name: "activated without error",
			getComponent: func() *Component {
				c := New("c1").
					WithInputs("i1").
					WithOutputs("o1").
					WithActivationFunc(func(this *Component) error {
						return port.ForwardSignals(this.InputByName("i1"), this.OutputByName("o1"))
					})
				//Only one input set
				c.InputByName("i1").PutSignals(signal.New(123))
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
					WithActivationFunc(func(this *Component) error {
						panic(errors.New("oh shrimps"))
						return nil
					})
				//Only one input set
				c.InputByName("i1").PutSignals(signal.New(123))
				return c
			},
			wantActivationResult: NewActivationResult("c1").
				SetActivated(true).
				WithActivationCode(ActivationCodePanicked).
				WithActivationError(errors.New("panicked with: oh shrimps")),
		},
		{
			name: "component panicked with string",
			getComponent: func() *Component {
				c := New("c1").
					WithInputs("i1").
					WithOutputs("o1").
					WithActivationFunc(func(this *Component) error {
						panic("oh shrimps")
						return nil
					})
				//Only one input set
				c.InputByName("i1").PutSignals(signal.New(123))
				return c
			},
			wantActivationResult: NewActivationResult("c1").
				SetActivated(true).
				WithActivationCode(ActivationCodePanicked).
				WithActivationError(errors.New("panicked with: oh shrimps")),
		},
		{
			name: "component is waiting for inputs",
			getComponent: func() *Component {
				c1 := New("c1").
					WithInputs("i1", "i2").
					WithOutputs("o1").
					WithActivationFunc(func(this *Component) error {
						if !this.Inputs().ByNames("i1", "i2").AllHaveSignals() {
							return NewErrWaitForInputs(false)
						}
						return nil
					})

				// Only one input set
				c1.InputByName("i1").PutSignals(signal.New(123))

				return c1
			},
			wantActivationResult: &ActivationResult{
				componentName:   "c1",
				activated:       true,
				code:            ActivationCodeWaitingForInputsClear,
				activationError: NewErrWaitForInputs(false),
			},
		},
		{
			name: "component is waiting for inputs and wants to keep them",
			getComponent: func() *Component {
				c1 := New("c1").
					WithInputs("i1", "i2").
					WithOutputs("o1").
					WithActivationFunc(func(this *Component) error {
						if !this.Inputs().ByNames("i1", "i2").AllHaveSignals() {
							return NewErrWaitForInputs(true)
						}
						return nil
					})

				// Only one input set
				c1.InputByName("i1").PutSignals(signal.New(123))

				return c1
			},
			wantActivationResult: &ActivationResult{
				componentName:   "c1",
				activated:       true,
				code:            ActivationCodeWaitingForInputsKeep,
				activationError: NewErrWaitForInputs(true),
			},
		},
		{
			name: "with chain error from input port",
			getComponent: func() *Component {
				c := New("c").WithInputs("i1").WithOutputs("o1")
				c.Inputs().With(port.New("p").WithErr(errors.New("some error")))
				return c
			},
			wantActivationResult: NewActivationResult("c").
				WithActivationCode(ActivationCodeUndefined).
				WithErr(errors.New("some error")),
		},
		{
			name: "with chain error from output port",
			getComponent: func() *Component {
				c := New("c").WithInputs("i1").WithOutputs("o1")
				c.Outputs().With(port.New("p").WithErr(errors.New("some error")))
				return c
			},
			wantActivationResult: NewActivationResult("c").
				WithActivationCode(ActivationCodeUndefined).
				WithErr(errors.New("some error")),
		},
		{
			name: "component not activated, logger must be empty",
			getComponent: func() *Component {
				c := New("c1").
					WithInputs("i1").
					WithOutputs("o1").
					WithActivationFunc(func(this *Component) error {
						log.Println("This must not be logged, as component must not activate")
						return nil
					})
				return c
			},

			wantActivationResult: NewActivationResult("c1").
				SetActivated(false).
				WithActivationCode(ActivationCodeNoInput),
			loggerAssertions: func(t *testing.T, output []byte) {
				assert.Len(t, output, 0)
			},
		},
		{
			name: "activated with error, with logging",
			getComponent: func() *Component {
				c := New("c1").
					WithInputs("i1").
					WithActivationFunc(func(this *Component) error {
						this.logger.Println("This line must be logged")
						return errors.New("test error")
					})
				//Only one input set
				c.InputByName("i1").PutSignals(signal.New(123))
				return c
			},
			wantActivationResult: NewActivationResult("c1").
				SetActivated(true).
				WithActivationCode(ActivationCodeReturnedError).
				WithActivationError(errors.New("component returned an error: test error")),
			loggerAssertions: func(t *testing.T, output []byte) {
				assert.Len(t, output, 2+3+21+24) //lengths of component name, prefix, flags and logged message
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := log.Default()

			var loggerOutput bytes.Buffer
			logger.SetOutput(&loggerOutput)

			gotActivationResult := tt.getComponent().WithLogger(logger).MaybeActivate()
			assert.Equal(t, tt.wantActivationResult.Activated(), gotActivationResult.Activated())
			assert.Equal(t, tt.wantActivationResult.ComponentName(), gotActivationResult.ComponentName())
			assert.Equal(t, tt.wantActivationResult.Code(), gotActivationResult.Code())
			if tt.wantActivationResult.IsError() {
				assert.EqualError(t, gotActivationResult.ActivationError(), tt.wantActivationResult.ActivationError().Error())
			} else {
				assert.False(t, gotActivationResult.IsError())
			}

			if tt.loggerAssertions != nil {
				tt.loggerAssertions(t, loggerOutput.Bytes())
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
				assert.Equal(t, component.Outputs().Len(), 2)
				assert.Equal(t, component.Inputs().Len(), 3)
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
				assert.Equal(t, component.Outputs().Len(), 2)
				assert.Equal(t, component.Inputs().Len(), 5)
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
				assert.Equal(t, component.Inputs().Len(), 2)
				assert.Equal(t, component.Outputs().Len(), 3)
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
				assert.Equal(t, component.Inputs().Len(), 2)
				assert.Equal(t, component.Outputs().Len(), 5)
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

func TestComponent_WithLabels(t *testing.T) {
	type args struct {
		labels common.LabelsCollection
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
				labels: common.LabelsCollection{
					"l1": "v1",
					"l2": "v2",
				},
			},
			assertions: func(t *testing.T, component *Component) {
				assert.Len(t, component.Labels(), 2)
				assert.True(t, component.HasAllLabels("l1", "l2"))
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

func TestComponent_ShortcutMethods(t *testing.T) {
	t.Run("InputByName", func(t *testing.T) {
		c := New("c").WithInputs("a", "b", "c")
		assert.Equal(t, port.New("b").WithLabels(common.LabelsCollection{
			port.DirectionLabel: port.DirectionIn,
		}), c.InputByName("b"))
	})

	t.Run("OutputByName", func(t *testing.T) {
		c := New("c").WithOutputs("a", "b", "c")
		assert.Equal(t, port.New("b").WithLabels(common.LabelsCollection{
			port.DirectionLabel: port.DirectionOut,
		}), c.OutputByName("b"))
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
