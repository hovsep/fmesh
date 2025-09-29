package errorhandling

import (
	"errors"
	"testing"

	"github.com/hovsep/fmesh"
	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/signal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Signal(t *testing.T) {
	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "no errors",
			test: func(t *testing.T) {
				sig := signal.New(123)
				_, err := sig.Payload()
				assert.False(t, sig.HasErr())
				require.NoError(t, err)

				_ = sig.PayloadOrDefault(555)
				assert.False(t, sig.HasErr())

				_ = sig.PayloadOrNil()
				assert.False(t, sig.HasErr())
			},
		},
		{
			name: "error propagated from group to signal",
			test: func(t *testing.T) {
				emptyGroup := signal.NewGroup()

				sig := emptyGroup.FirstSignal()
				assert.True(t, sig.HasErr())
				require.Error(t, sig.Err())

				_, err := sig.Payload()
				require.Error(t, err)
				require.EqualError(t, err, signal.ErrNoSignalsInGroup.Error())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, tt.test)
	}
}

func Test_FMesh(t *testing.T) {
	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "no errors",
			test: func(t *testing.T) {
				fm := fmesh.New("test").WithComponents(
					component.New("c1").WithInputs("num1", "num2").
						WithOutputs("sum").
						WithActivationFunc(func(this *component.Component) error {
							num1 := this.InputByName("num1").FirstSignalPayloadOrDefault(0).(int)
							num2 := this.InputByName("num2").FirstSignalPayloadOrDefault(0).(int)
							this.OutputByName("sum").PutSignals(signal.New(num1 + num2))
							return nil
						}),
				)

				fm.Components().ByName("c1").InputByName("num1").PutSignals(signal.New(10))
				fm.Components().ByName("c1").InputByName("num2").PutSignals(signal.New(5))

				_, err := fm.Run()
				assert.False(t, fm.HasErr())
				assert.NoError(t, err)
			},
		},
		{
			name: "error propagated from component",
			test: func(t *testing.T) {
				fm := fmesh.New("test").WithComponents(
					component.New("c1").
						WithInputs("num1", "num2").
						WithOutputs("sum").
						WithActivationFunc(func(this *component.Component) error {
							num1 := this.InputByName("num1").FirstSignalPayloadOrDefault(0).(int)
							num2 := this.InputByName("num2").FirstSignalPayloadOrDefault(0).(int)
							this.OutputByName("sum").PutSignals(signal.New(num1 + num2))
							return nil
						}).
						WithErr(errors.New("some error in component")),
				)

				fm.Components().ByName("c1").InputByName("num1").PutSignals(signal.New(10))
				fm.Components().ByName("c1").InputByName("num2").PutSignals(signal.New(5))

				_, err := fm.Run()
				assert.True(t, fm.HasErr())
				require.Error(t, err)
				require.EqualError(t, err, "some error in component")
			},
		},
		{
			name: "error propagated from port",
			test: func(t *testing.T) {
				fm := fmesh.New("test").WithComponents(
					component.New("c1").
						WithInputs("num1", "num2").
						WithOutputs("sum").
						WithActivationFunc(func(this *component.Component) error {
							num1 := this.InputByName("num1").FirstSignalPayloadOrDefault(0).(int)
							num2 := this.InputByName("num2").FirstSignalPayloadOrDefault(0).(int)
							this.OutputByName("sum").PutSignals(signal.New(num1 + num2))
							return nil
						}),
				)

				// Trying to search port by wrong name must lead to error which will bubble up at f-mesh level
				fm.Components().ByName("c1").InputByName("num777").PutSignals(signal.New(10))
				fm.Components().ByName("c1").InputByName("num2").PutSignals(signal.New(5))

				_, err := fm.Run()
				assert.True(t, fm.HasErr())
				require.Error(t, err)
				require.ErrorContains(t, err, "port not found, port name: num777")
			},
		},
		{
			name: "error propagated from signal",
			test: func(t *testing.T) {
				fm := fmesh.New("test").WithComponents(
					component.New("c1").WithInputs("num1", "num2").
						WithOutputs("sum").WithActivationFunc(func(this *component.Component) error {
						num1 := this.InputByName("num1").FirstSignalPayloadOrDefault(0).(int)
						num2 := this.InputByName("num2").FirstSignalPayloadOrDefault(0).(int)
						this.OutputByName("sum").PutSignals(signal.New(num1 + num2))
						return nil
					}),
				)

				fm.Components().ByName("c1").InputByName("num1").PutSignals(signal.New(10).WithErr(errors.New("some error in input signal")))
				fm.Components().ByName("c1").InputByName("num2").PutSignals(signal.New(5))

				_, err := fm.Run()
				assert.True(t, fm.HasErr())
				require.Error(t, err)
				require.ErrorContains(t, err, "some error in input signal")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, tt.test)
	}
}
