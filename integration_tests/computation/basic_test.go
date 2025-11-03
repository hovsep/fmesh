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
				c1 := component.New("c1").
					WithDescription("adds 2 to the input").
					AddInputs("num").
					AddOutputs("res").
					WithActivationFunc(func(this *component.Component) error {
						num := this.InputByName("num").Signals().FirstPayloadOrNil()
						this.OutputByName("res").PutSignals(signal.New(num.(int) + 2))
						return nil
					})

				c2 := component.New("c2").
					WithDescription("multiplies by 3").
					AddInputs("num").
					AddOutputs("res").
					WithActivationFunc(func(this *component.Component) error {
						num := this.InputByName("num").Signals().FirstPayloadOrDefault(0)
						this.OutputByName("res").PutSignals(signal.New(num.(int) * 3))
						return nil
					})

				c1.OutputByName("res").PipeTo(c2.InputByName("num"))
				return fmesh.NewWithConfig("fm", &fmesh.Config{
					ErrorHandlingStrategy: fmesh.StopOnFirstErrorOrPanic,
					CyclesLimit:           10,
				}).AddComponents(c1, c2)
			},
			setInputs: func(fm *fmesh.FMesh) {
				fm.Components().ByName("c1").InputByName("num").PutSignals(signal.New(32))
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
				processor := component.New("processor").
					WithDescription("processes data using mixed ports").
					// Simple ports (created by name only)
					AddInputs("raw_data", "metadata").
					// Advanced port (pre-configured with description and labels)
					AttachInputPorts(
						port.NewInput("config").
							WithDescription("Configuration parameters").
							AddLabel("required", "true").
							AddLabel("type", "config"),
					).
					// Simple output port
					AddOutputs("logs").
					// Advanced output ports (pre-configured)
					AttachOutputPorts(
						port.NewOutput("result").
							WithDescription("Processed result").
							AddLabel("format", "json"),
						port.NewOutput("error").
							WithDescription("Error details if any").
							AddLabel("status", "error"),
					).
					WithActivationFunc(func(this *component.Component) error {
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
						this.OutputByName("result").PutSignals(signal.New(result))
						this.OutputByName("logs").PutSignals(signal.New("Processed with metadata: " + metadata))
						// error port stays empty (no error)

						return nil
					})

				// Verifier component with simple ports
				verifier := component.New("verifier").
					AddInputs("value", "log").
					AddOutputs("verified").
					WithActivationFunc(func(this *component.Component) error {
						// Wait for all inputs
						if !this.Inputs().AllHaveSignals() {
							return nil
						}

						value := this.InputByName("value").Signals().FirstPayloadOrDefault(0).(int)
						log := this.InputByName("log").Signals().FirstPayloadOrDefault("").(string)

						// Verify we received data from both simple and advanced ports
						verified := value > 0 && log != ""
						this.OutputByName("verified").PutSignals(signal.New(verified))
						return nil
					})

				// Connect ports
				processor.OutputByName("result").PipeTo(verifier.InputByName("value"))
				processor.OutputByName("logs").PipeTo(verifier.InputByName("log"))

				return fmesh.NewWithConfig("mixed_ports_fm", &fmesh.Config{
					ErrorHandlingStrategy: fmesh.StopOnFirstErrorOrPanic,
					CyclesLimit:           10,
				}).AddComponents(processor, verifier)
			},
			setInputs: func(fm *fmesh.FMesh) {
				proc := fm.Components().ByName("processor")
				// Send data to simple ports
				proc.InputByName("raw_data").PutSignals(signal.New(10))
				proc.InputByName("metadata").PutSignals(signal.New("test"))
				// Send data to advanced port
				proc.InputByName("config").PutSignals(signal.New(5))
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
		fm := fmesh.NewWithConfig("hello world", &fmesh.Config{
			ErrorHandlingStrategy: fmesh.StopOnFirstErrorOrPanic,
			CyclesLimit:           10,
		}).
			AddComponents(
				component.New("concat").
					AddInputs("i1", "i2").
					AddOutputs("res").
					WithActivationFunc(func(this *component.Component) error {
						word1 := this.InputByName("i1").Signals().FirstPayloadOrDefault("").(string)
						word2 := this.InputByName("i2").Signals().FirstPayloadOrDefault("").(string)
						this.OutputByName("res").PutSignals(signal.New(word1 + word2))
						return nil
					}),
				component.New("case").
					AddInputs("i1").
					AddOutputs("res").
					WithActivationFunc(func(this *component.Component) error {
						inputString := this.InputByName("i1").Signals().FirstPayloadOrDefault("").(string)
						this.OutputByName("res").PutSignals(signal.New(strings.ToTitle(inputString)))
						return nil
					}))

		fm.Components().ByName("concat").OutputByName("res").PipeTo(
			fm.Components().ByName("case").InputByName("i1"),
		)

		// Init inputs
		fm.Components().ByName("concat").InputByName("i1").PutSignals(signal.New("hello "))
		fm.Components().ByName("concat").InputByName("i2").PutSignals(signal.New("world !"))

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
