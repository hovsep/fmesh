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

func Test_PipingToOutputs(t *testing.T) {
	tests := []struct {
		name       string
		setupFM    func() *fmesh.FMesh
		setInputs  func(fm *fmesh.FMesh)
		assertions func(t *testing.T, fm *fmesh.FMesh, cycles cycle.Collection, err error)
	}{
		{
			// Simple case: both generator and injector are activated only once and in the same cycle
			name: "single signal injection",
			setupFM: func() *fmesh.FMesh {
				gen := component.New("generator").
					WithDescription("Just generates a signal").
					WithInputs("start").
					WithOutputs("res").WithActivationFunc(func(inputs port.Collection, outputs port.Collection) error {
					outputs.PutSignals(signal.New(111))
					return nil
				})

				inj := component.New("injector").
					WithDescription("Adds signals to gen.res output port").
					WithInputs("start").
					WithOutputs("res").WithActivationFunc(func(inputs port.Collection, outputs port.Collection) error {
					outputs.PutSignals(signal.New(222))
					return nil
				})

				// o2o pipe:
				inj.Outputs().ByName("res").PipeTo(gen.Outputs().ByName("res"))

				fm := fmesh.New("injector").WithComponents(gen, inj)

				return fm
			},
			setInputs: func(fm *fmesh.FMesh) {
				fm.Components().ByName("generator").Inputs().ByName("start").PutSignals(signal.New("start gen"))
				fm.Components().ByName("injector").Inputs().ByName("start").PutSignals(signal.New("start inj"))
			},
			assertions: func(t *testing.T, fm *fmesh.FMesh, cycles cycle.Collection, err error) {
				assert.NoError(t, err)
				assert.Len(t, fm.Components().ByName("generator").Outputs().ByName("res").Signals(), 2)
				assert.False(t, fm.Components().ByName("injector").Outputs().ByName("res").HasSignals())
			},
		},
		{
			// 2 components have symmetrically connected output (both are generators and injectors at the same time)
			name: "outputs exchange",
			setupFM: func() *fmesh.FMesh {
				c1 := component.New("c1").
					WithDescription("Generates a signal").
					WithInputs("start").
					WithOutputs("res").WithActivationFunc(func(inputs port.Collection, outputs port.Collection) error {
					outputs.PutSignals(signal.New("signal from c1"))
					return nil
				})

				c2 := component.New("c2").
					WithDescription("Generates a signal").
					WithInputs("start").
					WithOutputs("res").WithActivationFunc(func(inputs port.Collection, outputs port.Collection) error {
					outputs.PutSignals(signal.New("signal from c2"))
					return nil
				})

				// o2o pipe 1:
				c1.Outputs().ByName("res").PipeTo(c2.Outputs().ByName("res"))
				c2.Outputs().ByName("res").PipeTo(c1.Outputs().ByName("res"))

				fm := fmesh.New("exchange").WithComponents(c1, c2)

				return fm
			},
			setInputs: func(fm *fmesh.FMesh) {
				fm.Components().ByName("c1").Inputs().ByName("start").PutSignals(signal.New("start c1"))
				fm.Components().ByName("c2").Inputs().ByName("start").PutSignals(signal.New("start c2"))
			},
			assertions: func(t *testing.T, fm *fmesh.FMesh, cycles cycle.Collection, err error) {
				assert.NoError(t, err)
				assert.Len(t, fm.Components().ByName("c1").Outputs().ByName("res").Signals(), 1)
				assert.Len(t, fm.Components().ByName("c2").Outputs().ByName("res").Signals(), 1)
				assert.Contains(t, fm.Components().ByName("c1").Outputs().ByName("res").Signals().AllPayloads(), "signal from c2")
				assert.Contains(t, fm.Components().ByName("c2").Outputs().ByName("res").Signals().AllPayloads(), "signal from c1")
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
