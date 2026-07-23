package piping

import (
	"math/rand"
	"testing"
	"time"

	"github.com/hovsep/fmesh/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hovsep/fmesh"
	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/cycle"
	"github.com/hovsep/fmesh/port"
	"github.com/hovsep/fmesh/signal"
)

func Test_Fan(t *testing.T) {
	tests := []struct {
		name       string
		setupFM    func() *fmesh.FMesh
		setInputs  func(fm *fmesh.FMesh)
		assertions func(t *testing.T, fm *fmesh.FMesh, cycles []*cycle.Cycle, runErr error)
	}{
		{
			name: "fan-out (3 pipes from 1 source port)",
			setupFM: func() *fmesh.FMesh {
				producer := testutil.MustComponent("producer",
					component.WithInputs("start"),
					component.WithOutputs("o1"),
					component.WithActivationFunc(func(this *component.Component) error {
						return this.OutputByName("o1").PutSignals(signal.New(time.Now()))
					}))

				consumer1 := testutil.MustComponent("consumer1",
					component.WithInputs("i1"),
					component.WithOutputs("o1"),
					component.WithActivationFunc(func(this *component.Component) error {
						return port.ForwardSignals(this.InputByName("i1"), this.OutputByName("o1"))
					}))

				consumer2 := testutil.MustComponent("consumer2",
					component.WithInputs("i1"),
					component.WithOutputs("o1"),
					component.WithActivationFunc(func(this *component.Component) error {
						return port.ForwardSignals(this.InputByName("i1"), this.OutputByName("o1"))
					}))

				consumer3 := testutil.MustComponent("consumer3",
					component.WithInputs("i1"),
					component.WithOutputs("o1"),
					component.WithActivationFunc(func(this *component.Component) error {
						return port.ForwardSignals(this.InputByName("i1"), this.OutputByName("o1"))
					}))

				if err := producer.OutputByName("o1").PipeTo(
					consumer1.InputByName("i1"),
					consumer2.InputByName("i1"),
					consumer3.InputByName("i1")); err != nil {
					panic(err)
				}

				fm := testutil.MustFMesh("fan-out")
				if err := fm.AddComponents(producer, consumer1, consumer2, consumer3); err != nil {
					panic(err)
				}
				return fm
			},
			setInputs: func(fm *fmesh.FMesh) {
				// Fire the mesh
				if err := fm.Components().ByName("producer").InputByName("start").PutSignals(signal.New(struct{}{})); err != nil {
					panic(err)
				}
			},
			assertions: func(t *testing.T, fm *fmesh.FMesh, cycles []*cycle.Cycle, runErr error) {
				require.NoError(t, runErr)
				// AllMatch consumers received a signal
				c1, c2, c3 := fm.Components().ByName("consumer1"), fm.Components().ByName("consumer2"), fm.Components().ByName("consumer3")
				assert.True(t, c1.OutputByName("o1").HasSignals())
				assert.True(t, c2.OutputByName("o1").HasSignals())
				assert.True(t, c3.OutputByName("o1").HasSignals())

				// AllMatch 3 signals are the same (literally the same address in memory)
				sig1, err := c1.OutputByName("o1").Signals().FirstPayload()
				require.NoError(t, err)
				sig2, err := c2.OutputByName("o1").Signals().FirstPayload()
				require.NoError(t, err)
				sig3, err := c3.OutputByName("o1").Signals().FirstPayload()
				require.NoError(t, err)
				assert.Equal(t, sig1, sig2)
				assert.Equal(t, sig2, sig3)
			},
		},
		{
			name: "fan-in (3 pipes coming into 1 destination port)",
			setupFM: func() *fmesh.FMesh {
				producer1 := testutil.MustComponent("producer1",
					component.WithInputs("start"),
					component.WithOutputs("o1"),
					component.WithActivationFunc(func(this *component.Component) error {
						return this.OutputByName("o1").PutSignals(signal.New(rand.Int()))
					}))

				producer2 := testutil.MustComponent("producer2",
					component.WithInputs("start"),
					component.WithOutputs("o1"),
					component.WithActivationFunc(func(this *component.Component) error {
						return this.OutputByName("o1").PutSignals(signal.New(rand.Int()))
					}))

				producer3 := testutil.MustComponent("producer3",
					component.WithInputs("start"),
					component.WithOutputs("o1"),
					component.WithActivationFunc(func(this *component.Component) error {
						return this.OutputByName("o1").PutSignals(signal.New(rand.Int()))
					}))

				consumer := testutil.MustComponent("consumer",
					component.WithInputs("i1"),
					component.WithOutputs("o1"),
					component.WithActivationFunc(func(this *component.Component) error {
						return port.ForwardSignals(this.InputByName("i1"), this.OutputByName("o1"))
					}))

				if err := producer1.OutputByName("o1").PipeTo(consumer.InputByName("i1")); err != nil {
					panic(err)
				}
				if err := producer2.OutputByName("o1").PipeTo(consumer.InputByName("i1")); err != nil {
					panic(err)
				}
				if err := producer3.OutputByName("o1").PipeTo(consumer.InputByName("i1")); err != nil {
					panic(err)
				}

				fm := testutil.MustFMesh("multiplexer")
				if err := fm.AddComponents(producer1, producer2, producer3, consumer); err != nil {
					panic(err)
				}
				return fm
			},
			setInputs: func(fm *fmesh.FMesh) {
				if err := fm.Components().ByName("producer1").InputByName("start").PutSignals(signal.New(struct{}{})); err != nil {
					panic(err)
				}
				if err := fm.Components().ByName("producer2").InputByName("start").PutSignals(signal.New(struct{}{})); err != nil {
					panic(err)
				}
				if err := fm.Components().ByName("producer3").InputByName("start").PutSignals(signal.New(struct{}{})); err != nil {
					panic(err)
				}
			},
			assertions: func(t *testing.T, fm *fmesh.FMesh, cycles []*cycle.Cycle, runErr error) {
				require.NoError(t, runErr)
				// Consumer received a signal
				assert.True(t, fm.Components().ByName("consumer").OutputByName("o1").HasSignals())

				// The signal is combined and consist of 3 payloads
				resultSignals := fm.Components().ByName("consumer").OutputByName("o1").Signals()
				assert.Equal(t, 3, resultSignals.Len())

				// And they are all different
				signals := resultSignals.All()
				sig0, err := resultSignals.FirstPayload()
				require.NoError(t, err)
				sig1, err := signals[1].Payload()
				require.NoError(t, err)
				sig2, err := signals[2].Payload()
				require.NoError(t, err)

				assert.NotEqual(t, sig0, sig1)
				assert.NotEqual(t, sig1, sig2)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fm := tt.setupFM()
			tt.setInputs(fm)
			runResult, err := fm.Run()
			cycles := runResult.Cycles.All()
			tt.assertions(t, fm, cycles, err)
		})
	}
}
