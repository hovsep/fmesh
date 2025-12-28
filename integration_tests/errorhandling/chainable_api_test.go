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
			name: "First on empty group returns nil",
			test: func(t *testing.T) {
				emptyGroup := signal.NewGroup()

				sig := emptyGroup.First()
				assert.Nil(t, sig)

				// FirstPayload returns the appropriate error
				_, err := emptyGroup.FirstPayload()
				require.Error(t, err)
				require.ErrorIs(t, err, signal.ErrNoSignalsInGroup)
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

				// Mesh has error propagated from component, so it's unusable
				assert.True(t, fm.HasChainableErr())

				// ComponentByName returns nil because mesh has error
				assert.Nil(t, fm.ComponentByName("c1"))

				// Mesh run fails with the propagated error
				_, err := fm.Run()
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

				// Looking up a non-existent port returns nil
				badPort := fm.Components().ByName("c1").InputByName("num777")
				assert.Nil(t, badPort)

				// Component and mesh remain unpoisoned
				c1 := fm.Components().ByName("c1")
				require.NotNil(t, c1)
				assert.False(t, c1.HasChainableErr())
				assert.False(t, fm.HasChainableErr())

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
