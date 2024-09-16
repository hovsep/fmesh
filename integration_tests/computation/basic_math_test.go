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

func Test_BasicMath(t *testing.T) {
	tests := []struct {
		name       string
		setupFM    func() *fmesh.FMesh
		setInputs  func(fm *fmesh.FMesh)
		assertions func(t *testing.T, fm *fmesh.FMesh, cycles cycle.Collection, err error)
	}{
		{
			name: "add and multiply",
			setupFM: func() *fmesh.FMesh {
				c1 := component.NewComponent("c1").
					WithDescription("adds 2 to the input").
					WithInputs("num").
					WithOutputs("res").
					WithActivationFunc(func(inputs port.Collection, outputs port.Collection) error {
						num := inputs.ByName("num").Signal().Payload().(int)
						outputs.ByName("res").PutSignal(signal.New(num + 2))
						return nil
					})

				c2 := component.NewComponent("c2").
					WithDescription("multiplies by 3").
					WithInputs("num").
					WithOutputs("res").
					WithActivationFunc(func(inputs port.Collection, outputs port.Collection) error {
						num := inputs.ByName("num").Signal().Payload().(int)
						outputs.ByName("res").PutSignal(signal.New(num * 3))
						return nil
					})

				c1.Outputs().ByName("res").PipeTo(c2.Inputs().ByName("num"))

				return fmesh.New("fm").WithComponents(c1, c2).WithErrorHandlingStrategy(fmesh.StopOnFirstError)
			},
			setInputs: func(fm *fmesh.FMesh) {
				fm.Components().ByName("c1").Inputs().ByName("num").PutSignal(signal.New(32))
			},
			assertions: func(t *testing.T, fm *fmesh.FMesh, cycles cycle.Collection, err error) {
				assert.NoError(t, err)
				assert.Len(t, cycles, 3)

				resultSignal := fm.Components().ByName("c2").Outputs().ByName("res").Signal()
				assert.Len(t, resultSignal.Payloads(), 1)
				assert.Equal(t, 102, resultSignal.Payload().(int))
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
