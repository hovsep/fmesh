package computation

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/hovsep/fmesh"
	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/cycle"
	"github.com/hovsep/fmesh/port"
	"github.com/hovsep/fmesh/signal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mustComponent(name string, opts ...component.Option) *component.Component {
	c, err := component.New(name, opts...)
	if err != nil {
		panic(err)
	}
	return c
}

func mustInputPort(name string, opts ...port.Option) *port.Port {
	p, err := port.NewInput(name, opts...)
	if err != nil {
		panic(err)
	}
	return p
}

func mustOutputPort(name string, opts ...port.Option) *port.Port {
	p, err := port.NewOutput(name, opts...)
	if err != nil {
		panic(err)
	}
	return p
}

func mustFMesh(name string, opts ...fmesh.Option) *fmesh.FMesh {
	fm, err := fmesh.New(name, opts...)
	if err != nil {
		panic(err)
	}
	return fm
}

func mustPutSignals(p *port.Port, signals ...*signal.Signal) {
	if err := p.PutSignals(signals...); err != nil {
		panic(err)
	}
}

func mustPipeTo(src *port.Port, dsts ...*port.Port) {
	if err := src.PipeTo(dsts...); err != nil {
		panic(err)
	}
}

func Test_Math(t *testing.T) {
	tests := []struct {
		name       string
		setupFM    func() *fmesh.FMesh
		setInputs  func(fm *fmesh.FMesh)
		assertions func(t *testing.T, fm *fmesh.FMesh, cycles cycle.Cycles, err error)
	}{
		{
			name: "add and multiply",
			setupFM: func() *fmesh.FMesh {
				c1 := mustComponent("c1",
					component.WithInputs("num"),
					component.WithOutputs("res"),
					component.WithActivationFunc(func(this *component.Component) error {
						num := this.InputByName("num").Signals().FirstPayloadOrNil()
						return this.OutputByName("res").PutSignals(signal.New(num.(int) + 2))
					}),
				).WithDescription("adds 2 to the input")

				c2 := mustComponent("c2",
					component.WithInputs("num"),
					component.WithOutputs("res"),
					component.WithActivationFunc(func(this *component.Component) error {
						num := this.InputByName("num").Signals().FirstPayloadOrDefault(0)
						return this.OutputByName("res").PutSignals(signal.New(num.(int) * 3))
					}),
				).WithDescription("multiplies by 3")

				if err := c1.OutputByName("res").PipeTo(c2.InputByName("num")); err != nil {
					panic(err)
				}
				fm := mustFMesh("fm", fmesh.WithConfig(&fmesh.Config{
					ErrorHandlingStrategy: fmesh.StopOnFirstErrorOrPanic,
					CyclesLimit:           10,
				}))
				if err := fm.AddComponents(c1, c2); err != nil {
					panic(err)
				}
				return fm
			},
			setInputs: func(fm *fmesh.FMesh) {
				if err := fm.Components().ByName("c1").InputByName("num").PutSignals(signal.New(32)); err != nil {
					panic(err)
				}
			},
			assertions: func(t *testing.T, fm *fmesh.FMesh, cycles cycle.Cycles, err error) {
				require.NoError(t, err)
				assert.Len(t, cycles, 3)

				resultSignals := fm.Components().ByName("c2").OutputByName("res").Signals()
				sig, err := resultSignals.FirstPayload()
				require.NoError(t, err)
				assert.Equal(t, 1, resultSignals.Len())
				assert.Equal(t, 102, sig.(int))
			},
		},
		{
			name: "mixed port creation - simple and advanced",
			setupFM: func() *fmesh.FMesh {
				// Component with mixed port creation
				processor := mustComponent("processor",
					component.WithInputs("raw_data", "metadata"),
					component.WithOutputs("logs"),
					component.WithActivationFunc(func(this *component.Component) error {
						// Check if all required inputs have signals
						if !this.Inputs().AllHaveSignals() {
							return nil // Wait for all inputs
						}

						// Verify all ports are functional regardless of the creation method
						rawData := this.InputByName("raw_data").Signals().FirstPayloadOrDefault(0).(int)
						metadata := this.InputByName("metadata").Signals().FirstPayloadOrDefault("").(string)
						config := this.InputByName("config").Signals().FirstPayloadOrDefault(1).(int)

						// Process: (rawData * config) + len(metadata)
						result := (rawData * config) + len(metadata)

						// Write to all outputs (both simple and advanced)
						if err := this.OutputByName("result").PutSignals(signal.New(result)); err != nil {
							return err
						}
						return this.OutputByName("logs").PutSignals(signal.New("Processed with metadata: " + metadata))
						// error port stays empty (no error)
					}),
				).WithDescription("processes data using mixed ports")

				// Add advanced ports
				if err := processor.AttachInputPorts(
					mustInputPort("config",
						port.WithDescription("Configuration parameters"),
						port.WithLabel("required", "true"),
						port.WithLabel("type", "config"),
					),
				); err != nil {
					panic(err)
				}
				if err := processor.AttachOutputPorts(
					mustOutputPort("result",
						port.WithDescription("Processed result"),
						port.WithLabel("format", "json"),
					),
					mustOutputPort("error",
						port.WithDescription("Error details if any"),
						port.WithLabel("status", "error"),
					),
				); err != nil {
					panic(err)
				}

				// Verifier component with simple ports
				verifier := mustComponent("verifier",
					component.WithInputs("value", "log"),
					component.WithOutputs("verified"),
					component.WithActivationFunc(func(this *component.Component) error {
						// Wait for all inputs
						if !this.Inputs().AllHaveSignals() {
							return nil
						}

						value := this.InputByName("value").Signals().FirstPayloadOrDefault(0).(int)
						log := this.InputByName("log").Signals().FirstPayloadOrDefault("").(string)

						// Verify we received data from both simple and advanced ports
						verified := value > 0 && log != ""
						return this.OutputByName("verified").PutSignals(signal.New(verified))
					}),
				)

				// Connect ports
				if err := processor.OutputByName("result").PipeTo(verifier.InputByName("value")); err != nil {
					panic(err)
				}
				if err := processor.OutputByName("logs").PipeTo(verifier.InputByName("log")); err != nil {
					panic(err)
				}

				fm := mustFMesh("mixed_ports_fm", fmesh.WithConfig(&fmesh.Config{
					ErrorHandlingStrategy: fmesh.StopOnFirstErrorOrPanic,
					CyclesLimit:           10,
				}))
				if err := fm.AddComponents(processor, verifier); err != nil {
					panic(err)
				}
				return fm
			},
			setInputs: func(fm *fmesh.FMesh) {
				proc := fm.Components().ByName("processor")
				// Send data to simple ports
				mustPutSignals(proc.InputByName("raw_data"), signal.New(10))
				mustPutSignals(proc.InputByName("metadata"), signal.New("test"))
				// Send data to advanced port
				mustPutSignals(proc.InputByName("config"), signal.New(5))
			},
			assertions: func(t *testing.T, fm *fmesh.FMesh, cycles cycle.Cycles, err error) {
				require.NoError(t, err)
				require.NotEmpty(t, cycles, "should have at least one cycle")

				proc := fm.Components().ByName("processor")
				verif := fm.Components().ByName("verifier")

				// Verify mesh executed successfully
				assert.Len(t, cycles, 3, "should take 3 cycles: processor -> verifier -> done")

				// Verify port metadata (only advanced ports have descriptions/labels)
				assert.Empty(t, proc.InputByName("raw_data").Description(), "simple port should have no description")
				assert.Empty(t, proc.InputByName("metadata").Description(), "simple port should have no description")
				assert.Equal(t, "Configuration parameters", proc.InputByName("config").Description(), "advanced port should have description")
				assert.True(t, proc.InputByName("config").Labels().ValueIs("required", "true"), "advanced port should have labels")

				assert.Empty(t, proc.OutputByName("logs").Description(), "simple port should have no description")
				assert.Equal(t, "Processed result", proc.OutputByName("result").Description(), "advanced port should have description")
				assert.True(t, proc.OutputByName("result").Labels().ValueIs("format", "json"), "advanced port should have labels")

				// Verify data flowed correctly through the entire chain (processor -> verifier)
				// The verifier's output confirms that both simple and advanced ports worked
				verifiedSignals := verif.OutputByName("verified").Signals()
				verified, err := verifiedSignals.FirstPayload()
				require.NoError(t, err)
				assert.True(t, verified.(bool), "verifier should confirm data flowed through all port types (simple and advanced)")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fm := tt.setupFM()
			tt.setInputs(fm)
			runResult, err := fm.Run()
			cycles, cycleErr := runResult.Cycles.All()
			require.NoError(t, cycleErr)
			tt.assertions(t, fm, cycles, err)
		})
	}
}

func Test_Readme(t *testing.T) {
	t.Run("readme test", func(t *testing.T) {
		fm := mustFMesh("hello world", fmesh.WithConfig(&fmesh.Config{
			ErrorHandlingStrategy: fmesh.StopOnFirstErrorOrPanic,
			CyclesLimit:           10,
		}))

		concat := mustComponent("concat",
			component.WithInputs("i1", "i2"),
			component.WithOutputs("res"),
			component.WithActivationFunc(func(this *component.Component) error {
				word1 := this.InputByName("i1").Signals().FirstPayloadOrDefault("").(string)
				word2 := this.InputByName("i2").Signals().FirstPayloadOrDefault("").(string)
				return this.OutputByName("res").PutSignals(signal.New(word1 + word2))
			}),
		)

		caseC := mustComponent("case",
			component.WithInputs("i1"),
			component.WithOutputs("res"),
			component.WithActivationFunc(func(this *component.Component) error {
				inputString := this.InputByName("i1").Signals().FirstPayloadOrDefault("").(string)
				return this.OutputByName("res").PutSignals(signal.New(strings.ToTitle(inputString)))
			}),
		)

		if err := fm.AddComponents(concat, caseC); err != nil {
			panic(err)
		}

		if err := fm.Components().ByName("concat").OutputByName("res").PipeTo(
			fm.Components().ByName("case").InputByName("i1"),
		); err != nil {
			panic(err)
		}

		// Init inputs
		mustPutSignals(fm.Components().ByName("concat").InputByName("i1"), signal.New("hello "))
		mustPutSignals(fm.Components().ByName("concat").InputByName("i2"), signal.New("world !"))

		// Run the mesh
		_, err := fm.Run()

		// Check for errors
		if err != nil {
			fmt.Println("F-Mesh returned an error")
			os.Exit(1)
		}

		// Extract results
		results := fm.ComponentByName("case").OutputByName("res").Signals().FirstPayloadOrNil()
		fmt.Printf("Result is :%v", results)
		assert.Equal(t, "HELLO WORLD !", results)
	})
}
