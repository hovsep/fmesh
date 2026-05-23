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

func Test_PortCreationAndManipulation(t *testing.T) {
	t.Run("mixed port creation with all features", func(t *testing.T) {
		// Create a component using both simple and advanced port creation APIs
		processor := mustComponent("data-processor",
			component.WithInputs("raw_data", "filter"),
			component.WithOutputs("processed", "metrics"),
		).
			WithDescription("Demonstrates all port creation and manipulation features")

		// Advanced API: attach ports with descriptions and labels
		require.NoError(t, processor.AttachInputPorts(
			mustInputPort("config", port.WithDescription("Configuration parameters")).
				AddLabel("required", "true").
				AddLabel("type", "json"),
			mustInputPort("metadata", port.WithDescription("Request metadata")).
				AddLabel("required", "false"),
		))
		require.NoError(t, processor.AttachOutputPorts(
			mustOutputPort("errors", port.WithDescription("Error details if processing fails")).
				AddLabel("severity", "high").
				AddLabel("format", "structured"),
		))
		processor.WithActivationFunc(func(this *component.Component) error {
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
			if err := this.OutputByName("processed").PutSignals(signal.New(result)); err != nil {
				return err
			}
			return this.OutputByName("metrics").PutSignals(
				signal.New(fmt.Sprintf("processed %d chars", len(result))),
			)
		})

		// Put signals on all inputs
		require.NoError(t, processor.InputByName("raw_data").PutSignals(signal.New("test data")))
		require.NoError(t, processor.InputByName("config").PutSignals(signal.New("prod")))
		require.NoError(t, processor.InputByName("filter").PutSignals(signal.New("all")))
		require.NoError(t, processor.InputByName("metadata").PutSignals(signal.New("user123")))

		// Create and run mesh
		fm := mustFMesh("test-mesh")
		require.NoError(t, fm.AddComponents(processor))
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
		c := mustComponent("label-demo",
			component.WithOutputs("output"),
		)
		require.NoError(t, c.AttachInputPorts(
			mustInputPort("input").
				AddLabel("env", "dev").
				AddLabel("version", "1.0").
				AddLabel("owner", "team-a"),
		))
		c.WithActivationFunc(func(this *component.Component) error {
			if !this.InputByName("input").HasSignals() {
				return nil
			}

			// Manipulate port labels during processing
			inputPort := this.InputByName("input")

			// Add a single label
			inputPort.AddLabel("processed", "true")

			// Remove specific labels
			inputPort.RemoveLabels("version")

			// Update labels
			inputPort.AddLabels(map[string]string{
				"env":   "prod", // update
				"build": "123",  // add new
			})

			// Forward signal to output
			sig, _ := inputPort.Signals().FirstPayload()
			return this.OutputByName("output").PutSignals(signal.New(sig))
		})

		// Set up and run
		require.NoError(t, c.InputByName("input").PutSignals(signal.New("data")))
		fm := mustFMesh("label-mesh")
		require.NoError(t, fm.AddComponents(c))
		_, err := fm.Run()
		require.NoError(t, err)

		// Verify label manipulation results
		inputPort := c.InputByName("input")
		lbls, err := inputPort.Labels().All()
		require.NoError(t, err)

		assert.Equal(t, "prod", lbls["env"], "env should be updated")
		assert.Equal(t, "123", lbls["build"], "build should be added")
		assert.Equal(t, "true", lbls["processed"], "processed should be added")
		assert.Equal(t, "team-a", lbls["owner"], "owner should remain")
		assert.NotContains(t, lbls, "version", "version should be removed")
	})

	t.Run("incremental port addition", func(t *testing.T) {
		// Demonstrate adding ports one by one
		c := mustComponent("incremental",
			component.WithOutputs("result"),
		)
		require.NoError(t, c.AddInputs("a"))   // Add first input
		require.NoError(t, c.AddInputs("b"))   // Add second input
		require.NoError(t, c.AttachInputPorts( // Add with details
			mustInputPort("c", port.WithDescription("Third input")),
		))
		c.WithActivationFunc(func(this *component.Component) error {
			if !this.Inputs().AllHaveSignals() {
				return nil
			}

			a := this.InputByName("a").Signals().FirstPayloadOrDefault(0).(int)
			b := this.InputByName("b").Signals().FirstPayloadOrDefault(0).(int)
			cv := this.InputByName("c").Signals().FirstPayloadOrDefault(0).(int)

			return this.OutputByName("result").PutSignals(signal.New(a + b + cv))
		})

		// Verify all ports exist and work
		require.NoError(t, c.InputByName("a").PutSignals(signal.New(1)))
		require.NoError(t, c.InputByName("b").PutSignals(signal.New(2)))
		require.NoError(t, c.InputByName("c").PutSignals(signal.New(3)))

		fm := mustFMesh("incremental-mesh")
		require.NoError(t, fm.AddComponents(c))
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
		c := mustComponent("collection-demo",
			component.WithInputs("i1", "i2", "i3"),
			component.WithOutputs("summary"),
		)
		require.NoError(t, c.AttachInputPorts(
			mustInputPort("i4").AddLabel("priority", "high"),
			mustInputPort("i5").AddLabel("priority", "low"),
		))
		c.WithActivationFunc(func(this *component.Component) error {
			inputs := this.Inputs()

			// Count ports with signals
			portsWithSignals := inputs.Filter(func(p *port.Port) bool {
				return p.HasSignals()
			}).Len()

			// Find high priority ports
			highPriorityPorts := inputs.Filter(func(p *port.Port) bool {
				lbls, err := p.Labels().All()
				if err != nil {
					return false
				}
				return lbls["priority"] == "high"
			})

			// Apply operation to all ports (add processing label)
			if err := inputs.ForEach(func(p *port.Port) error {
				p.AddLabel("checked", "true")
				return nil
			}); err != nil {
				return err
			}

			highPriorityCount, _ := highPriorityPorts.All()
			summary := fmt.Sprintf("Total: %d, WithSignals: %d, HighPriority: %d",
				inputs.Len(), portsWithSignals, len(highPriorityCount))

			return this.OutputByName("summary").PutSignals(signal.New(summary))
		})

		// Put signals on some ports
		require.NoError(t, c.InputByName("i1").PutSignals(signal.New(1)))
		require.NoError(t, c.InputByName("i2").PutSignals(signal.New(2)))
		require.NoError(t, c.InputByName("i4").PutSignals(signal.New(4)))

		fm := mustFMesh("collection-mesh")
		require.NoError(t, fm.AddComponents(c))
		_, err := fm.Run()
		require.NoError(t, err)

		summary, err := c.OutputByName("summary").Signals().FirstPayload()
		require.NoError(t, err)
		assert.Equal(t, "Total: 5, WithSignals: 3, HighPriority: 1", summary.(string))

		// Verify all ports were labeled
		require.NoError(t, c.Inputs().ForEach(func(p *port.Port) error {
			assert.True(t, p.Labels().Has("checked"))
			return nil
		}))
	})
}
