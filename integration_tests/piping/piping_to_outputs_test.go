package integration_tests

import (
	"github.com/hovsep/fmesh"
	"github.com/hovsep/fmesh/cycle"
	"testing"
)

func Test_PipingToOutputs(t *testing.T) {
	tests := []struct {
		name       string
		setupFM    func() *fmesh.FMesh
		setInputs  func(fm *fmesh.FMesh)
		assertions func(t *testing.T, fm *fmesh.FMesh, cycles cycle.Collection, err error)
	}{
		{
			name: "injector",
			setupFM: func() *fmesh.FMesh {
				return fmesh.New("injector")
			},
			setInputs: func(fm *fmesh.FMesh) {

			},
			assertions: func(t *testing.T, fm *fmesh.FMesh, cycles cycle.Collection, err error) {

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
