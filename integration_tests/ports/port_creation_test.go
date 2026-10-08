package ports

import (
	"context"
	"fmt"
	"testing"

	"github.com/hovsep/fmesh/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/port"
	"github.com/hovsep/fmesh/signal"
)

func Test_PortCreationAndManipulation(t *testing.T) {
	t.Run("mixed port creation with all features", func(t *testing.T) {
		// Create a component using both simple and advanced port creation APIs
		processor := testutil.MustComponent("data-processor",
			component.WithInputs("raw_data", "filter"),
			component.WithOutputs("processed", "metrics"),
			component.WithDescription("Demonstrates all port creation and manipulation features"),
			component.WithActivationFunc(func(_ context.Context, this *component.Component) error {
				// Wait for required inputs
				if !this.InputByName("raw_data").HasSignals() ||
					!this.InputByName("config").HasSignals() {
					return nil
				}

				// Read inputs
				data := this.InputByName("raw_data").Signals().FirstPayloadOrDefault("")
				config := this.InputByName("config").Signals().FirstPayloadOrDefault("")
				filter := this.InputByName("filter").Signals().FirstPayloadOrDefault("")
				metadata := this.InputByName("metadata").Signals().FirstPayloadOrDefault("none")

				// Process data
				result := fmt.Sprintf("[%s:%s] %s (meta: %s)", config, filter, data, metadata)

				// Write outputs
				if err := this.OutputByName("processed").PutSignals(signal.New(result)); err != nil {
					return err
				}
				return this.OutputByName("metrics").PutSignals(
					signal.New(fmt.Sprintf("processed %d chars", len(result))),
				)
			}),
		)

		// Advanced API: attach ports with descriptions and metadata
		require.NoError(t, processor.AttachInputPorts(
			testutil.MustInputPort("config",
				port.WithDescription("Configuration parameters"),
				port.WithMeta("required", "true"),
				port.WithMeta("type", "json")),
			testutil.MustInputPort("metadata",
				port.WithDescription("Request metadata"),
				port.WithMeta("required", "false")),
		))
		require.NoError(t, processor.AttachOutputPorts(
			testutil.MustOutputPort("errors",
				port.WithDescription("Error details if processing fails"),
				port.WithMeta("severity", "high"),
				port.WithMeta("format", "structured")),
		))

		// Put signals on all inputs
		require.NoError(t, processor.InputByName("raw_data").PutSignals(signal.New("test data")))
		require.NoError(t, processor.InputByName("config").PutSignals(signal.New("prod")))
		require.NoError(t, processor.InputByName("filter").PutSignals(signal.New("all")))
		require.NoError(t, processor.InputByName("metadata").PutSignals(signal.New("user123")))

		fm := testutil.MustFMesh("test-mesh")
		require.NoError(t, fm.AddComponents(processor))
		_, err := fm.Run(context.Background())
		require.NoError(t, err)

		// Verify input ports (both simple and advanced)
		inputs := processor.Inputs()
		require.NoError(t, err)
		assert.Equal(t, 4, inputs.Len(), "should have 4 input ports")

		// Simple port (created with AddInputs) has no description
		rawData := inputs.ByName("raw_data")
		assert.NotNil(t, rawData)
		assert.Empty(t, rawData.Description())

		// Advanced port (created with AttachInputPorts) has description and metadata
		config := inputs.ByName("config")
		assert.NotNil(t, config)
		assert.Equal(t, "Configuration parameters", config.Description())
		assert.True(t, config.Meta().Has("required"))
		assert.True(t, config.Meta().Has("type"))

		// Verify output ports (both simple and advanced)
		outputs := processor.Outputs()
		require.NoError(t, err)
		assert.Equal(t, 3, outputs.Len(), "should have 3 output ports")

		// Simple port has no description
		processed := outputs.ByName("processed")
		assert.NotNil(t, processed)
		assert.Empty(t, processed.Description())

		// Advanced port has description and metadata
		errors := outputs.ByName("errors")
		assert.NotNil(t, errors)
		assert.Equal(t, "Error details if processing fails", errors.Description())
		assert.True(t, errors.Meta().Has("severity"))
		assert.True(t, errors.Meta().Has("format"))
		entries := errors.Meta().All()
		require.NoError(t, err)
		assert.Equal(t, "high", entries["severity"])
		assert.Equal(t, "structured", entries["format"])

		processedData, err := processor.OutputByName("processed").Signals().FirstPayload()
		require.NoError(t, err)
		assert.Equal(t, "[prod:all] test data (meta: user123)", processedData.(string))

		metrics, err := processor.OutputByName("metrics").Signals().FirstPayload()
		require.NoError(t, err)
		assert.Contains(t, metrics.(string), "processed")

		// Error port should be empty (no errors occurred)
		assert.False(t, processor.OutputByName("errors").HasSignals())
	})

	t.Run("port metadata manipulation", func(t *testing.T) {
		c := testutil.MustComponent("meta-demo",
			component.WithOutputs("output"),
			component.WithActivationFunc(func(_ context.Context, this *component.Component) error {
				if !this.InputByName("input").HasSignals() {
					return nil
				}

				// Manipulate port metadata during processing
				inputPort := this.InputByName("input")

				inputPort.Meta().Set("processed", "true")

				// Remove specific entries
				inputPort.Meta().Remove("version")

				// Update entries
				inputPort.Meta().SetMany(map[string]string{
					"env":   "prod", // update
					"build": "123",  // add new
				})

				// Forward signal to output
				sig, _ := inputPort.Signals().FirstPayload()
				return this.OutputByName("output").PutSignals(signal.New(sig))
			}),
		)
		require.NoError(t, c.AttachInputPorts(
			testutil.MustInputPort("input",
				port.WithMeta("env", "dev"),
				port.WithMeta("version", "1.0"),
				port.WithMeta("owner", "team-a")),
		))

		require.NoError(t, c.InputByName("input").PutSignals(signal.New("data")))
		fm := testutil.MustFMesh("meta-mesh")
		require.NoError(t, fm.AddComponents(c))
		_, err := fm.Run(context.Background())
		require.NoError(t, err)

		inputPort := c.InputByName("input")
		lbls := inputPort.Meta().All()
		require.NoError(t, err)

		assert.Equal(t, "prod", lbls["env"], "env should be updated")
		assert.Equal(t, "123", lbls["build"], "build should be added")
		assert.Equal(t, "true", lbls["processed"], "processed should be added")
		assert.Equal(t, "team-a", lbls["owner"], "owner should remain")
		assert.NotContains(t, lbls, "version", "version should be removed")
	})

	t.Run("incremental port addition", func(t *testing.T) {
		// Demonstrate adding ports one by one
		c := testutil.MustComponent("incremental",
			component.WithOutputs("result"),
			component.WithActivationFunc(func(_ context.Context, this *component.Component) error {
				if !this.Inputs().AllHaveSignals() {
					return nil
				}

				a := this.InputByName("a").Signals().FirstPayloadOrDefault(0)
				b := this.InputByName("b").Signals().FirstPayloadOrDefault(0)
				cv := this.InputByName("c").Signals().FirstPayloadOrDefault(0)

				return this.OutputByName("result").PutSignals(signal.New(a + b + cv))
			}),
		)
		require.NoError(t, c.AddInputs("a"))   // Add first input
		require.NoError(t, c.AddInputs("b"))   // Add second input
		require.NoError(t, c.AttachInputPorts( // Add with details
			testutil.MustInputPort("c", port.WithDescription("Third input")),
		))

		require.NoError(t, c.InputByName("a").PutSignals(signal.New(1)))
		require.NoError(t, c.InputByName("b").PutSignals(signal.New(2)))
		require.NoError(t, c.InputByName("c").PutSignals(signal.New(3)))

		fm := testutil.MustFMesh("incremental-mesh")
		require.NoError(t, fm.AddComponents(c))
		_, err := fm.Run(context.Background())
		require.NoError(t, err)

		result, err := c.OutputByName("result").Signals().FirstPayload()
		require.NoError(t, err)
		assert.Equal(t, 6, result.(int))

		assert.Equal(t, "Third input", c.InputByName("c").Description())
	})

	t.Run("port collection operations", func(t *testing.T) {
		// Demonstrate port collection methods
		c := testutil.MustComponent("collection-demo",
			component.WithInputs("i1", "i2", "i3"),
			component.WithOutputs("summary"),
			component.WithActivationFunc(func(_ context.Context, this *component.Component) error {
				inputs := this.Inputs()

				// Count ports with signals
				portsWithSignals := inputs.Filter(func(p *port.Port) bool {
					return p.HasSignals()
				}).Len()

				// Find high priority ports
				highPriorityPorts := inputs.Filter(func(p *port.Port) bool {
					lbls := p.Meta().All()
					return lbls["priority"] == "high"
				})

				// Apply operation to all ports (add processing entry)
				for p := range inputs.All() {
					p.Meta().Set("checked", "true")
				}

				summary := fmt.Sprintf("Total: %d, WithSignals: %d, HighPriority: %d",
					inputs.Len(), portsWithSignals, highPriorityPorts.Len())

				return this.OutputByName("summary").PutSignals(signal.New(summary))
			}),
		)
		require.NoError(t, c.AttachInputPorts(
			testutil.MustInputPort("i4", port.WithMeta("priority", "high")),
			testutil.MustInputPort("i5", port.WithMeta("priority", "low")),
		))

		// Put signals on some ports
		require.NoError(t, c.InputByName("i1").PutSignals(signal.New(1)))
		require.NoError(t, c.InputByName("i2").PutSignals(signal.New(2)))
		require.NoError(t, c.InputByName("i4").PutSignals(signal.New(4)))

		fm := testutil.MustFMesh("collection-mesh")
		require.NoError(t, fm.AddComponents(c))
		_, err := fm.Run(context.Background())
		require.NoError(t, err)

		summary, err := c.OutputByName("summary").Signals().FirstPayload()
		require.NoError(t, err)
		assert.Equal(t, "Total: 5, WithSignals: 3, HighPriority: 1", summary.(string))

		for p := range c.Inputs().All() {
			assert.True(t, p.Meta().Has("checked"))
		}
	})
}
