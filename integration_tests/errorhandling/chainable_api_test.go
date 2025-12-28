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
				assert.False(t, sig.HasChainableErr())
				require.NoError(t, err)

				_ = sig.PayloadOrDefault(555)
				assert.False(t, sig.HasChainableErr())

				_ = sig.PayloadOrNil()
				assert.False(t, sig.HasChainableErr())
			},
		},
		{
			name: "error propagated from group to signal",
			test: func(t *testing.T) {
				emptyGroup := signal.NewGroup()

				sig := emptyGroup.First()
				assert.True(t, sig.HasChainableErr())
				require.Error(t, sig.ChainableErr())

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
				fm := fmesh.New("test").AddComponents(
					component.New("c1").AddInputs("num1", "num2").
						AddOutputs("sum").
						WithActivationFunc(func(this *component.Component) error {
							num1 := this.InputByName("num1").Signals().FirstPayloadOrDefault(0).(int)
							num2 := this.InputByName("num2").Signals().FirstPayloadOrDefault(0).(int)
							this.OutputByName("sum").PutSignals(signal.New(num1 + num2))
							return nil
						}),
				)

				fm.Components().ByName("c1").InputByName("num1").PutSignals(signal.New(10))
				fm.Components().ByName("c1").InputByName("num2").PutSignals(signal.New(5))

				_, err := fm.Run()
				assert.False(t, fm.HasChainableErr())
				assert.NoError(t, err)
			},
		},
		{
			name: "error propagated from component",
			test: func(t *testing.T) {
				fm := fmesh.New("test").AddComponents(
					component.New("c1").
						AddInputs("num1", "num2").
						AddOutputs("sum").
						WithActivationFunc(func(this *component.Component) error {
							num1 := this.InputByName("num1").Signals().FirstPayloadOrDefault(0).(int)
							num2 := this.InputByName("num2").Signals().FirstPayloadOrDefault(0).(int)
							this.OutputByName("sum").PutSignals(signal.New(num1 + num2))
							return nil
						}).
						WithChainableErr(errors.New("some error in component")),
				)

				fm.Components().ByName("c1").InputByName("num1").PutSignals(signal.New(10))
				fm.Components().ByName("c1").InputByName("num2").PutSignals(signal.New(5))

				_, err := fm.Run()
				assert.True(t, fm.HasChainableErr())
				require.Error(t, err)
				require.ErrorContains(t, err, "some error in component")
			},
		},
		{
			name: "port lookup error does not poison component or mesh",
			test: func(t *testing.T) {
				fm := fmesh.New("test").AddComponents(
					component.New("c1").
						AddInputs("num1", "num2").
						AddOutputs("sum").
						WithActivationFunc(func(this *component.Component) error {
							num1 := this.InputByName("num1").Signals().FirstPayloadOrDefault(0).(int)
							num2 := this.InputByName("num2").Signals().FirstPayloadOrDefault(0).(int)
							this.OutputByName("sum").PutSignals(signal.New(num1 + num2))
							return nil
						}),
				)

				// Looking up a non-existent port returns a port with error,
				// but does NOT poison the component or mesh (query error, not config error)
				badPort := fm.Components().ByName("c1").InputByName("num777")
				assert.True(t, badPort.HasChainableErr())
				require.ErrorContains(t, badPort.ChainableErr(), "port not found")
				require.ErrorContains(t, badPort.ChainableErr(), "port name: num777")

				// Component and mesh remain unpoisoned
				assert.False(t, fm.Components().ByName("c1").HasChainableErr())
				assert.False(t, fm.HasChainableErr())

				// PutSignals on the bad port does nothing useful but also doesn't panic
				badPort.PutSignals(signal.New(10))

				// Valid port operations work as expected
				fm.Components().ByName("c1").InputByName("num1").PutSignals(signal.New(10))
				fm.Components().ByName("c1").InputByName("num2").PutSignals(signal.New(5))

				// Mesh runs successfully
				_, err := fm.Run()
				assert.False(t, fm.HasChainableErr())
				require.NoError(t, err)
			},
		},
		{
			name: "error propagated from signal",
			test: func(t *testing.T) {
				fm := fmesh.New("test").AddComponents(
					component.New("c1").AddInputs("num1", "num2").
						AddOutputs("sum").WithActivationFunc(func(this *component.Component) error {
						num1 := this.InputByName("num1").Signals().FirstPayloadOrDefault(0).(int)
						num2 := this.InputByName("num2").Signals().FirstPayloadOrDefault(0).(int)
						this.OutputByName("sum").PutSignals(signal.New(num1 + num2))
						return nil
					}),
				)

				fm.Components().ByName("c1").InputByName("num1").PutSignals(signal.New(10).WithChainableErr(errors.New("some error in input signal")))
				fm.Components().ByName("c1").InputByName("num2").PutSignals(signal.New(5))

				_, err := fm.Run()
				assert.True(t, fm.HasChainableErr())
				require.Error(t, err)
				require.ErrorContains(t, err, "some error in input signal")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, tt.test)
	}
}
