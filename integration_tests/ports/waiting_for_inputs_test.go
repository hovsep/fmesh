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

func Test_WaitingForInputs(t *testing.T) {
	tests := []struct {
		name       string
		setupFM    func() *fmesh.FMesh
		setInputs  func(fm *fmesh.FMesh)
		assertions func(t *testing.T, fm *fmesh.FMesh, cycles cycle.Collection, err error)
	}{
		{
			name: "waits for single input and keep signals",
			setupFM: func() *fmesh.FMesh {
				return fmesh.New("fm").WithComponents(
					component.New("waiter").
						WithInputs("i1", "i2").
						WithActivationFunc(func(inputs port.Collection, outputs port.Collection) error {
							if !inputs.ByNames("i1", "i2").AllHaveSignals() {
								return component.NewErrWaitForInputs(true)
							}
							return nil
						}),
				)
			},
			setInputs: func(fm *fmesh.FMesh) {
				//Only one input set
				fm.Components().ByName("waiter").Inputs().ByName("i1").PutSignals(signal.New(1))
			},
			assertions: func(t *testing.T, fm *fmesh.FMesh, cycles cycle.Collection, err error) {
				assert.NoError(t, err)

				// Signal is kept on input port
				assert.True(t, fm.Components().ByName("waiter").Inputs().ByName("i1").HasSignals())
			},
		},
		{
			//@TODO:make this test pass
			name: "waits for multiple input",
			setupFM: func() *fmesh.FMesh {
				return fmesh.New("fm").WithComponents(
					component.New("waiter").
						WithInputs("i1", "i2", "i3").
						WithActivationFunc(func(inputs port.Collection, outputs port.Collection) error {
							if !inputs.ByNames("i2", "i3").AllHaveSignals() {
								return component.NewErrWaitForInputs(false)
							}
							return nil
						}),
				)
			},
			setInputs: func(fm *fmesh.FMesh) {
				//Only one input set
				fm.Components().ByName("waiter").Inputs().ByName("i1").PutSignals(signal.New(1))
			},
			assertions: func(t *testing.T, fm *fmesh.FMesh, cycles cycle.Collection, err error) {
				assert.NoError(t, err)

				// Signal is not kept on input port
				assert.False(t, fm.Components().ByName("waiter").Inputs().ByName("i1").HasSignals())
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
