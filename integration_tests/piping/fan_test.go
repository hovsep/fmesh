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

func Test_Fan(t *testing.T) {
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
					component.New("producer").
						WithInputs("start").
						WithOutputs("o1").
						WithActivationFunc(func(inputs port.Collection, outputs port.Collection) error {
							outputs.ByName("o1").PutSignals(signal.New(time.Now()))
							return nil
						}),

					component.New("consumer1").
						WithInputs("i1").
						WithOutputs("o1").
						WithActivationFunc(func(inputs port.Collection, outputs port.Collection) error {
							//Bypass received signal to output
							port.ForwardSignals(inputs.ByName("i1"), outputs.ByName("o1"))
							return nil
						}),

					component.New("consumer2").
						WithInputs("i1").
						WithOutputs("o1").
						WithActivationFunc(func(inputs port.Collection, outputs port.Collection) error {
							//Bypass received signal to output
							port.ForwardSignals(inputs.ByName("i1"), outputs.ByName("o1"))
							return nil
						}),

					component.New("consumer3").
						WithInputs("i1").
						WithOutputs("o1").
						WithActivationFunc(func(inputs port.Collection, outputs port.Collection) error {
							//Bypass received signal to output
							port.ForwardSignals(inputs.ByName("i1"), outputs.ByName("o1"))
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
				fm.Components().ByName("producer").Inputs().ByName("start").PutSignals(signal.New(struct{}{}))
			},
			assertions: func(t *testing.T, fm *fmesh.FMesh, cycles cycle.Collection, err error) {
				//All consumers received a signal
				c1, c2, c3 := fm.Components().ByName("consumer1"), fm.Components().ByName("consumer2"), fm.Components().ByName("consumer3")
				assert.True(t, c1.Outputs().ByName("o1").HasSignals())
				assert.True(t, c2.Outputs().ByName("o1").HasSignals())
				assert.True(t, c3.Outputs().ByName("o1").HasSignals())

				//All 3 signals are the same (literally the same address in memory)
				sig1, err := c1.Outputs().ByName("o1").Buffer().FirstPayload()
				assert.NoError(t, err)
				sig2, err := c2.Outputs().ByName("o1").Buffer().FirstPayload()
				assert.NoError(t, err)
				sig3, err := c3.Outputs().ByName("o1").Buffer().FirstPayload()
				assert.NoError(t, err)
				assert.Equal(t, sig1, sig2)
				assert.Equal(t, sig2, sig3)
			},
		},
		{
			name: "fan-in (3 pipes coming into 1 destination port)",
			setupFM: func() *fmesh.FMesh {
				producer1 := component.New("producer1").
					WithInputs("start").
					WithOutputs("o1").
					WithActivationFunc(func(inputs port.Collection, outputs port.Collection) error {
						outputs.ByName("o1").PutSignals(signal.New(rand.Int()))
						return nil
					})

				producer2 := component.New("producer2").
					WithInputs("start").
					WithOutputs("o1").
					WithActivationFunc(func(inputs port.Collection, outputs port.Collection) error {
						outputs.ByName("o1").PutSignals(signal.New(rand.Int()))
						return nil
					})

				producer3 := component.New("producer3").
					WithInputs("start").
					WithOutputs("o1").
					WithActivationFunc(func(inputs port.Collection, outputs port.Collection) error {
						outputs.ByName("o1").PutSignals(signal.New(rand.Int()))
						return nil
					})
				consumer := component.New("consumer").
					WithInputs("i1").
					WithOutputs("o1").
					WithActivationFunc(func(inputs port.Collection, outputs port.Collection) error {
						//Bypass
						port.ForwardSignals(inputs.ByName("i1"), outputs.ByName("o1"))
						return nil
					})

				producer1.Outputs().ByName("o1").PipeTo(consumer.Inputs().ByName("i1"))
				producer2.Outputs().ByName("o1").PipeTo(consumer.Inputs().ByName("i1"))
				producer3.Outputs().ByName("o1").PipeTo(consumer.Inputs().ByName("i1"))

				return fmesh.New("multiplexer").WithComponents(producer1, producer2, producer3, consumer)
			},
			setInputs: func(fm *fmesh.FMesh) {
				fm.Components().ByName("producer1").Inputs().ByName("start").PutSignals(signal.New(struct{}{}))
				fm.Components().ByName("producer2").Inputs().ByName("start").PutSignals(signal.New(struct{}{}))
				fm.Components().ByName("producer3").Inputs().ByName("start").PutSignals(signal.New(struct{}{}))
			},
			assertions: func(t *testing.T, fm *fmesh.FMesh, cycles cycle.Collection, err error) {
				assert.NoError(t, err)
				//Consumer received a signal
				assert.True(t, fm.Components().ByName("consumer").Outputs().ByName("o1").HasSignals())

				//The signal is combined and consist of 3 payloads
				resultSignals := fm.Components().ByName("consumer").Outputs().ByName("o1").Buffer()
				assert.Len(t, resultSignals.SignalsOrNil(), 3)

				//And they are all different
				sig0, err := resultSignals.FirstPayload()
				assert.NoError(t, err)
				sig1, err := resultSignals.SignalsOrNil()[1].Payload()
				assert.NoError(t, err)
				sig2, err := resultSignals.SignalsOrNil()[2].Payload()
				assert.NoError(t, err)

				assert.NotEqual(t, sig0, sig1)
				assert.NotEqual(t, sig1, sig2)
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
