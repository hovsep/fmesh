package integration_tests

import (
	"github.com/hovsep/fmesh"
	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/cycle"
	"github.com/hovsep/fmesh/port"
	"github.com/hovsep/fmesh/signal"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
	"time"
)

func Test_Multiplexing(t *testing.T) {
	tests := []struct {
		name       string
		setupFM    func() *fmesh.FMesh
		setInputs  func(fm *fmesh.FMesh)
		assertions func(t *testing.T, fm *fmesh.FMesh, cycles cycle.Collection, err error)
	}{
		{
			name: "fan-out (3 pipes from 1 source port)",
			setupFM: func() *fmesh.FMesh {
				fm := fmesh.New("fan-out").WithComponents(
					component.NewComponent("producer").
						WithInputs("start").
						WithOutputs("o1").
						WithActivationFunc(func(inputs port.Collection, outputs port.Collection) error {
							outputs.ByName("o1").PutSignal(signal.New(time.Now()))
							return nil
						}),

					component.NewComponent("consumer1").
						WithInputs("i1").
						WithOutputs("o1").
						WithActivationFunc(func(inputs port.Collection, outputs port.Collection) error {
							//Bypass received signal to output
							port.ForwardSignal(inputs.ByName("i1"), outputs.ByName("o1"))
							return nil
						}),

					component.NewComponent("consumer2").
						WithInputs("i1").
						WithOutputs("o1").
						WithActivationFunc(func(inputs port.Collection, outputs port.Collection) error {
							//Bypass received signal to output
							port.ForwardSignal(inputs.ByName("i1"), outputs.ByName("o1"))
							return nil
						}),

					component.NewComponent("consumer3").
						WithInputs("i1").
						WithOutputs("o1").
						WithActivationFunc(func(inputs port.Collection, outputs port.Collection) error {
							//Bypass received signal to output
							port.ForwardSignal(inputs.ByName("i1"), outputs.ByName("o1"))
							return nil
						}),
				)

				fm.Components().ByName("producer").Outputs().ByName("o1").PipeTo(
					fm.Components().ByName("consumer1").Inputs().ByName("i1"),
					fm.Components().ByName("consumer2").Inputs().ByName("i1"),
					fm.Components().ByName("consumer3").Inputs().ByName("i1"))

				return fm
			},
			setInputs: func(fm *fmesh.FMesh) {
				//Fire the mesh
				fm.Components().ByName("producer").Inputs().ByName("start").PutSignal(signal.New(struct{}{}))
			},
			assertions: func(t *testing.T, fm *fmesh.FMesh, cycles cycle.Collection, err error) {
				//All consumers received a signal
				c1, c2, c3 := fm.Components().ByName("consumer1"), fm.Components().ByName("consumer2"), fm.Components().ByName("consumer3")
				assert.True(t, c1.Outputs().ByName("o1").HasSignal())
				assert.True(t, c2.Outputs().ByName("o1").HasSignal())
				assert.True(t, c3.Outputs().ByName("o1").HasSignal())

				//All 3 signals are the same (literally the same address in memory)
				sig1, sig2, sig3 := c1.Outputs().ByName("o1").Signal(), c2.Outputs().ByName("o1").Signal(), c3.Outputs().ByName("o1").Signal()
				assert.Equal(t, sig1.Payload(), sig2.Payload())
				assert.Equal(t, sig2.Payload(), sig3.Payload())
			},
		},
		{
			name: "multiplexing",
			setupFM: func() *fmesh.FMesh {
				producer1 := component.NewComponent("producer1").
					WithInputs("start").
					WithOutputs("o1").
					WithActivationFunc(func(inputs port.Collection, outputs port.Collection) error {
						outputs.ByName("o1").PutSignal(signal.New(rand.Int()))
						return nil
					})

				producer2 := component.NewComponent("producer2").
					WithInputs("start").
					WithOutputs("o1").
					WithActivationFunc(func(inputs port.Collection, outputs port.Collection) error {
						outputs.ByName("o1").PutSignal(signal.New(rand.Int()))
						return nil
					})

				producer3 := component.NewComponent("producer3").
					WithInputs("start").
					WithOutputs("o1").
					WithActivationFunc(func(inputs port.Collection, outputs port.Collection) error {
						outputs.ByName("o1").PutSignal(signal.New(rand.Int()))
						return nil
					})
				consumer := component.NewComponent("consumer").
					WithInputs("i1").
					WithOutputs("o1").
					WithActivationFunc(func(inputs port.Collection, outputs port.Collection) error {
						//Bypass
						port.ForwardSignal(inputs.ByName("i1"), outputs.ByName("o1"))
						return nil
					})

				producer1.Outputs().ByName("o1").PipeTo(consumer.Inputs().ByName("i1"))
				producer2.Outputs().ByName("o1").PipeTo(consumer.Inputs().ByName("i1"))
				producer3.Outputs().ByName("o1").PipeTo(consumer.Inputs().ByName("i1"))

				return fmesh.New("multiplexer").WithComponents(producer1, producer2, producer3, consumer)
			},
			setInputs: func(fm *fmesh.FMesh) {
				fm.Components().ByName("producer1").Inputs().ByName("start").PutSignal(signal.New(struct{}{}))
				fm.Components().ByName("producer2").Inputs().ByName("start").PutSignal(signal.New(struct{}{}))
				fm.Components().ByName("producer3").Inputs().ByName("start").PutSignal(signal.New(struct{}{}))
			},
			assertions: func(t *testing.T, fm *fmesh.FMesh, cycles cycle.Collection, err error) {
				//Consumer received a signal
				assert.True(t, fm.Components().ByName("consumer").Outputs().ByName("o1").HasSignal())

				//The signal is combined and consist of 3 payloads
				resultSignal := fm.Components().ByName("consumer").Outputs().ByName("o1").Signal()
				assert.Equal(t, resultSignal.Len(), 3)

				//And they are all different
				assert.NotEqual(t, resultSignal.Payloads()[0], resultSignal.Payloads()[1])
				assert.NotEqual(t, resultSignal.Payloads()[1], resultSignal.Payloads()[2])
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
