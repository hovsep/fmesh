package ports

import (
	"fmt"
	"testing"

	"github.com/hovsep/fmesh"
	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/port"
	"github.com/hovsep/fmesh/signal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_PortCreationAndManipulation(t *testing.T) {
	t.Run("mixed port creation with all features", func(t *testing.T) {
		// Create a component using both simple and advanced port creation APIs
		processor := component.New("data-processor").
			WithDescription("Demonstrates all port creation and manipulation features").
			// Simple API: create ports by name only
			AddInputs("raw_data", "filter").
			// Advanced API: create ports with descriptions and labels
			AttachInputPorts(
				port.NewInput("config").
					WithDescription("Configuration parameters").
					AddLabel("required", "true").
					AddLabel("type", "json"),
				port.NewInput("metadata").
					WithDescription("Request metadata").
					AddLabel("required", "false"),
			).
			// Simple API for outputs
			AddOutputs("processed", "metrics").
			// Advanced API for outputs with labels
			AttachOutputPorts(
				port.NewOutput("errors").
					WithDescription("Error details if processing fails").
					AddLabel("severity", "high").
					AddLabel("format", "structured"),
			).
			WithActivationFunc(func(this *component.Component) error {
				// Wait for required inputs
				if !this.InputByName("raw_data").HasSignals() ||
					!this.InputByName("config").HasSignals() {
					return nil
				}

				// Read inputs
				data := this.InputByName("raw_data").Signals().FirstPayloadOrDefault("").(string)
				config := this.InputByName("config").Signals().FirstPayloadOrDefault("").(string)
				filter := this.InputByName("filter").Signals().FirstPayloadOrDefault("").(string)
				metadata := this.InputByName("metadata").Signals().FirstPayloadOrDefault("none").(string)

				// Process data
				result := fmt.Sprintf("[%s:%s] %s (meta: %s)", config, filter, data, metadata)

				// Write outputs
				this.OutputByName("processed").PutSignals(signal.New(result))
				this.OutputByName("metrics").PutSignals(
					signal.New(fmt.Sprintf("processed %d chars", len(result))),
				)

				return nil
			})

		// Put signals on all inputs
		processor.InputByName("raw_data").PutSignals(signal.New("test data"))
		processor.InputByName("config").PutSignals(signal.New("prod"))
		processor.InputByName("filter").PutSignals(signal.New("all"))
		processor.InputByName("metadata").PutSignals(signal.New("user123"))

		// Create and run mesh
		fm := fmesh.New("test-mesh").AddComponents(processor)
		_, err := fm.Run()
		require.NoError(t, err)

		// Verify input ports (both simple and advanced)
		inputs, err := processor.Inputs().All()
		require.NoError(t, err)
		assert.Len(t, inputs, 4, "should have 4 input ports")

		// Simple port (created with AddInputs) has no description
		rawData := inputs["raw_data"]
		assert.NotNil(t, rawData)
		assert.Empty(t, rawData.Description())

		// Advanced port (created with AttachInputPorts) has description and labels
		config := inputs["config"]
		assert.NotNil(t, config)
		assert.Equal(t, "Configuration parameters", config.Description())
		assert.True(t, config.Labels().Has("required"))
		assert.True(t, config.Labels().Has("type"))

		// Verify output ports (both simple and advanced)
		outputs, err := processor.Outputs().All()
		require.NoError(t, err)
		assert.Len(t, outputs, 3, "should have 3 output ports")

		// Simple port has no description
		processed := outputs["processed"]
		assert.NotNil(t, processed)
		assert.Empty(t, processed.Description())

		// Advanced port has description and labels
		errors := outputs["errors"]
		assert.NotNil(t, errors)
		assert.Equal(t, "Error details if processing fails", errors.Description())
		assert.True(t, errors.Labels().Has("severity"))
		assert.True(t, errors.Labels().Has("format"))
		labelMap, err := errors.Labels().All()
		require.NoError(t, err)
		assert.Equal(t, "high", labelMap["severity"])
		assert.Equal(t, "structured", labelMap["format"])

		// Verify data flowed correctly through all port types
		processedData, err := processor.OutputByName("processed").Signals().FirstPayload()
		require.NoError(t, err)
		assert.Equal(t, "[prod:all] test data (meta: user123)", processedData.(string))

		metrics, err := processor.OutputByName("metrics").Signals().FirstPayload()
		require.NoError(t, err)
		assert.Contains(t, metrics.(string), "processed")

		// Error port should be empty (no errors occurred)
		assert.False(t, processor.OutputByName("errors").HasSignals())
	})

	t.Run("port label manipulation", func(t *testing.T) {
		// Create a component and manipulate port labels
		c := component.New("label-demo").
			AttachInputPorts(
				port.NewInput("input").
					AddLabel("env", "dev").
					AddLabel("version", "1.0").
					AddLabel("owner", "team-a"),
			).
			AddOutputs("output").
			WithActivationFunc(func(this *component.Component) error {
				if !this.InputByName("input").HasSignals() {
					return nil
				}

				// Manipulate port labels during processing
				inputPort := this.InputByName("input")

				// Add a single label
				inputPort.AddLabel("processed", "true")

				// Remove specific labels
				inputPort.WithoutLabels("version")

				// Update labels
				inputPort.AddLabels(map[string]string{
					"env":   "prod", // update
					"build": "123",  // add new
				})

				// Forward signal to output
				sig, _ := inputPort.Signals().FirstPayload()
				this.OutputByName("output").PutSignals(signal.New(sig))

				return nil
			})

		// Set up and run
		c.InputByName("input").PutSignals(signal.New("data"))
		fm := fmesh.New("label-mesh").AddComponents(c)
		_, err := fm.Run()
		require.NoError(t, err)

		// Verify label manipulation results
		inputPort := c.InputByName("input")
		labels, err := inputPort.Labels().All()
		require.NoError(t, err)

		assert.Equal(t, "prod", labels["env"], "env should be updated")
		assert.Equal(t, "123", labels["build"], "build should be added")
		assert.Equal(t, "true", labels["processed"], "processed should be added")
		assert.Equal(t, "team-a", labels["owner"], "owner should remain")
		assert.NotContains(t, labels, "version", "version should be removed")
	})

	t.Run("incremental port addition", func(t *testing.T) {
		// Demonstrate adding ports one by one
		c := component.New("incremental").
			AddInputs("a").                                                      // Add first input
			AddInputs("b").                                                      // Add second input
			AttachInputPorts(port.NewInput("c").WithDescription("Third input")). // Add with details
			AddOutputs("result").
			WithActivationFunc(func(this *component.Component) error {
				if !this.Inputs().AllHaveSignals() {
					return nil
				}

				a := this.InputByName("a").Signals().FirstPayloadOrDefault(0).(int)
				b := this.InputByName("b").Signals().FirstPayloadOrDefault(0).(int)
				c := this.InputByName("c").Signals().FirstPayloadOrDefault(0).(int)

				this.OutputByName("result").PutSignals(signal.New(a + b + c))
				return nil
			})

		// Verify all ports exist and work
		c.InputByName("a").PutSignals(signal.New(1))
		c.InputByName("b").PutSignals(signal.New(2))
		c.InputByName("c").PutSignals(signal.New(3))

		fm := fmesh.New("incremental-mesh").AddComponents(c)
		_, err := fm.Run()
		require.NoError(t, err)

		result, err := c.OutputByName("result").Signals().FirstPayload()
		require.NoError(t, err)
		assert.Equal(t, 6, result.(int))

		// Verify port c has description
		assert.Equal(t, "Third input", c.InputByName("c").Description())
	})

	t.Run("port collection operations", func(t *testing.T) {
		// Demonstrate port collection methods
		c := component.New("collection-demo").
			AddInputs("i1", "i2", "i3").
			AttachInputPorts(
				port.NewInput("i4").AddLabel("priority", "high"),
				port.NewInput("i5").AddLabel("priority", "low"),
			).
			AddOutputs("summary").
			WithActivationFunc(func(this *component.Component) error {
				inputs := this.Inputs()

				// Count ports with signals
				portsWithSignals := inputs.CountMatch(func(p *port.Port) bool {
					return p.HasSignals()
				})

				// Find high priority ports
				highPriorityPorts := inputs.Filter(func(p *port.Port) bool {
					labels, err := p.Labels().All()
					if err != nil {
						return false
					}
					return labels["priority"] == "high"
				})

				// Apply operation to all ports (add processing label)
				inputs.ForEach(func(p *port.Port) {
					p.AddLabel("checked", "true")
				})

				highPriorityCount, _ := highPriorityPorts.All()
				summary := fmt.Sprintf("Total: %d, WithSignals: %d, HighPriority: %d",
					inputs.Len(), portsWithSignals, len(highPriorityCount))

				this.OutputByName("summary").PutSignals(signal.New(summary))
				return nil
			})

		// Put signals on some ports
		c.InputByName("i1").PutSignals(signal.New(1))
		c.InputByName("i2").PutSignals(signal.New(2))
		c.InputByName("i4").PutSignals(signal.New(4))

		fm := fmesh.New("collection-mesh").AddComponents(c)
		_, err := fm.Run()
		require.NoError(t, err)

		summary, err := c.OutputByName("summary").Signals().FirstPayload()
		require.NoError(t, err)
		assert.Equal(t, "Total: 5, WithSignals: 3, HighPriority: 1", summary.(string))

		// Verify all ports were labeled
		inputs, err := c.Inputs().All()
		require.NoError(t, err)
		for _, p := range inputs {
			assert.True(t, p.Labels().Has("checked"))
		}
	})
}
