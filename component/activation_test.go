package component

import (
	"bytes"
	"errors"
	"github.com/hovsep/fmesh/port"
	"github.com/hovsep/fmesh/signal"
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
)

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

			// Compare activation functions by they result and error
			dummyComponent1 := New("c1").WithInputs("i1", "i2").WithOutputs("o1", "o2")
			dummyComponent2 := New("c2").WithInputs("i1", "i2").WithOutputs("o1", "o2")
			err1 := componentAfter.f(dummyComponent1)
			err2 := tt.args.f(dummyComponent2)
			assert.Equal(t, err1, err2)

			// Compare signals without keys (because they are random)
			assert.ElementsMatch(t, dummyComponent1.OutputByName("o1").AllSignalsOrNil(), dummyComponent2.OutputByName("o1").AllSignalsOrNil())
			assert.ElementsMatch(t, dummyComponent1.OutputByName("o2").AllSignalsOrNil(), dummyComponent2.OutputByName("o2").AllSignalsOrNil())

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
				// Only one input set
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
				// Only one input set
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
					})
				// Only one input set
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
					})
				// Only one input set
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
						this.Logger().Println("This must not be logged, as component must not activate")
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
				// Only one input set
				c.InputByName("i1").PutSignals(signal.New(123))
				return c
			},
			wantActivationResult: NewActivationResult("c1").
				SetActivated(true).
				WithActivationCode(ActivationCodeReturnedError).
				WithActivationError(errors.New("component returned an error: test error")),
			loggerAssertions: func(t *testing.T, output []byte) {
				assert.Len(t, output, 2+3+21+24) // lengths of component name, prefix, flags and logged message
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
