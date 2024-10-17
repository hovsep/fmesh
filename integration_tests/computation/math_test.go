package integration_tests

import (
	"github.com/hovsep/fmesh"
	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/cycle"
	"github.com/hovsep/fmesh/port"
	"github.com/hovsep/fmesh/signal"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_Math(t *testing.T) {
	tests := []struct {
		name       string
		setupFM    func() *fmesh.FMesh
		setInputs  func(fm *fmesh.FMesh)
		assertions func(t *testing.T, fm *fmesh.FMesh, cycles cycle.Collection, err error)
	}{
		{
			name: "add and multiply",
			setupFM: func() *fmesh.FMesh {
				c1 := component.New("c1").
					WithDescription("adds 2 to the input").
					WithInputs("num").
					WithOutputs("res").
					WithActivationFunc(func(inputs *port.Collection, outputs *port.Collection) error {
						num, err := inputs.ByName("num").Buffer().FirstPayload()
						if err != nil {
							return err
						}
						outputs.ByName("res").PutSignals(signal.New(num.(int) + 2))
						return nil
					})

				c2 := component.New("c2").
					WithDescription("multiplies by 3").
					WithInputs("num").
					WithOutputs("res").
					WithActivationFunc(func(inputs *port.Collection, outputs *port.Collection) error {
						num, err := inputs.ByName("num").Buffer().FirstPayload()
						if err != nil {
							return err
						}
						outputs.ByName("res").PutSignals(signal.New(num.(int) * 3))
						return nil
					})

				c1.Outputs().ByName("res").PipeTo(c2.Inputs().ByName("num"))

				return fmesh.New("fm").WithComponents(c1, c2).WithConfig(fmesh.Config{
					ErrorHandlingStrategy: fmesh.StopOnFirstErrorOrPanic,
					CyclesLimit:           10,
				})
			},
			setInputs: func(fm *fmesh.FMesh) {
				fm.Components().ByName("c1").Inputs().ByName("num").PutSignals(signal.New(32))
			},
			assertions: func(t *testing.T, fm *fmesh.FMesh, cycles cycle.Collection, err error) {
				assert.NoError(t, err)
				assert.Len(t, cycles, 3)

				resultSignals := fm.Components().ByName("c2").Outputs().ByName("res").Buffer()
				sig, err := resultSignals.FirstPayload()
				assert.NoError(t, err)
				assert.Len(t, resultSignals.SignalsOrNil(), 1)
				assert.Equal(t, 102, sig.(int))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fm := tt.setupFM()
			tt.setInputs(fm)
			cycles, err := fm.Run()
			tt.assertions(t, fm, cycles, err)
		})
	}
}
