package hooks

import (
	"errors"
	"testing"

	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/port"
	"github.com/hovsep/fmesh/signal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComponentHooks_PracticalErrorLogging(t *testing.T) {
	// Practical example: Error logging with context
	type ErrorLog struct {
		ComponentName string
		ErrorMessage  string
		ErrorType     string
	}
	var errorLog ErrorLog

	validationErr := errors.New("validation failed: negative value")

	c := mustComponent("validator",
		component.WithInputs("data"),
	).SetupHooks(func(h *component.Hooks) {
		h.OnError(func(ctx *component.ActivationContext) error {
			// Access the error and log it with component context
			errorLog = ErrorLog{
				ComponentName: ctx.Component.Name(),
				ErrorMessage:  ctx.Result.ActivationError().Error(),
				ErrorType:     "validation_error",
			}
			return nil
		})
	}).SetActivationFunc(func(c *component.Component) error {
		// Simulate validation logic
		inputVal := c.InputByName("data").Signals().First().PayloadOrDefault(0).(int)
		if inputVal < 0 {
			return validationErr
		}
		return nil
	})

	require.NoError(t, c.InputByName("data").PutSignals(signal.New(-5)))
	result := c.MaybeActivate()

	require.True(t, result.IsError())
	assert.Equal(t, "validator", errorLog.ComponentName)
	assert.Contains(t, errorLog.ErrorMessage, "validation failed")
	assert.Equal(t, "validation_error", errorLog.ErrorType)
}

func TestComponentHooks_PracticalOutputValidation(t *testing.T) {
	// Practical example: Validate output data in hooks
	var outputIsValid bool
	var outputValue int

	c := mustComponent("calculator",
		component.WithInputs("x", "y"),
		component.WithOutputs("result"),
	).SetupHooks(func(h *component.Hooks) {
		h.OnSuccess(func(ctx *component.ActivationContext) error {
			// Access and validate output signals
			resultPort := ctx.Component.OutputByName("result")
			if resultPort.Signals().Len() == 1 {
				val := resultPort.Signals().FirstPayloadOrDefault(0).(int)
				outputValue = val
				outputIsValid = val >= 0 && val <= 100 // Valid range check
			}
			return nil
		})
	}).SetActivationFunc(func(c *component.Component) error {
		x := c.InputByName("x").Signals().First().PayloadOrDefault(0).(int)
		y := c.InputByName("y").Signals().First().PayloadOrDefault(0).(int)
		result := x + y
		return c.OutputByName("result").PutSignals(signal.New(result))
	})

	require.NoError(t, c.InputByName("x").PutSignals(signal.New(30)))
	require.NoError(t, c.InputByName("y").PutSignals(signal.New(20)))
	c.MaybeActivate()

	assert.True(t, outputIsValid, "Output should be in valid range")
	assert.Equal(t, 50, outputValue)
}

func TestComponentHooks_PracticalMetricsCollection(t *testing.T) {
	// Practical example: Collect metrics about component execution
	type ComponentMetrics struct {
		SuccessCount       int
		ErrorCount         int
		PanicCount         int
		TotalActivations   int
		LastError          error
		OutputSignalCounts map[string]int
	}
	metrics := ComponentMetrics{
		OutputSignalCounts: make(map[string]int),
	}

	c := mustComponent("processor",
		component.WithInputs("in"),
		component.WithOutputs("success", "failure"),
	).SetupHooks(func(h *component.Hooks) {
		h.OnSuccess(func(ctx *component.ActivationContext) error {
			metrics.SuccessCount++
			// Track output signal counts
			if err := ctx.Component.Outputs().ForEach(func(p *port.Port) error {
				metrics.OutputSignalCounts[p.Name()] = p.Signals().Len()
				return nil
			}); err != nil {
				return err
			}
			return nil
		})

		h.OnError(func(ctx *component.ActivationContext) error {
			metrics.ErrorCount++
			metrics.LastError = ctx.Result.ActivationError()
			return nil
		})

		h.OnPanic(func(ctx *component.ActivationContext) error {
			metrics.PanicCount++
			return nil
		})

		h.AfterActivation(func(ctx *component.ActivationContext) error {
			metrics.TotalActivations++
			return nil
		})
	}).SetActivationFunc(func(c *component.Component) error {
		val := c.InputByName("in").Signals().First().PayloadOrDefault(0).(int)
		if val > 0 {
			return c.OutputByName("success").PutSignals(signal.New(val * 2))
		}
		return errors.New("invalid input")
	})

	// First activation: success
	require.NoError(t, c.InputByName("in").PutSignals(signal.New(5)))
	c.MaybeActivate()

	// Second activation: error
	require.NoError(t, c.InputByName("in").Clear())
	require.NoError(t, c.InputByName("in").PutSignals(signal.New(-1)))
	c.MaybeActivate()

	assert.Equal(t, 1, metrics.SuccessCount)
	assert.Equal(t, 1, metrics.ErrorCount)
	assert.Equal(t, 0, metrics.PanicCount)
	assert.Equal(t, 2, metrics.TotalActivations)
	require.Error(t, metrics.LastError)
	assert.Equal(t, 1, metrics.OutputSignalCounts["success"])
}

func TestComponentHooks_PracticalDataTransformation(t *testing.T) {
	// Practical example: Transform or enrich output data in hooks
	var enrichedOutput []map[string]any

	c := mustComponent("enricher",
		component.WithInputs("raw"),
		component.WithOutputs("enriched"),
	).SetupHooks(func(h *component.Hooks) {
		h.OnSuccess(func(ctx *component.ActivationContext) error {
			// Access output and create enriched version with metadata
			_, err := ctx.Component.OutputByName("enriched").Signals().ForEach(func(s *signal.Signal) error {
				enriched := map[string]any{
					"value":         s.PayloadOrDefault(nil),
					"component":     ctx.Component.Name(),
					"timestamp":     "2024-01-01", // In real code, use time.Now()
					"activation_ok": ctx.Result.Code() == component.ActivationCodeOK,
				}
				enrichedOutput = append(enrichedOutput, enriched)
				return nil
			})
			return err
		})
	}).SetActivationFunc(func(c *component.Component) error {
		// Process and output data
		_, err := c.InputByName("raw").Signals().ForEach(func(s *signal.Signal) error {
			processed := s.PayloadOrDefault(0).(int) * 10
			return c.OutputByName("enriched").PutSignals(signal.New(processed))
		})
		return err
	})

	require.NoError(t, c.InputByName("raw").PutSignals(signal.New(3), signal.New(7)))
	c.MaybeActivate()

	require.Len(t, enrichedOutput, 2)
	assert.Equal(t, 30, enrichedOutput[0]["value"])
	assert.Equal(t, "enricher", enrichedOutput[0]["component"])
	assert.Equal(t, true, enrichedOutput[0]["activation_ok"])
	assert.Equal(t, 70, enrichedOutput[1]["value"])
}

func TestComponentHooks_PracticalErrorRecovery(t *testing.T) {
	// Practical example: Attempt recovery or fallback on error
	var recoveryAttempted bool
	var fallbackValueProvided bool

	c := mustComponent("resilient",
		component.WithInputs("in"),
		component.WithOutputs("out"),
	).SetupHooks(func(h *component.Hooks) {
		h.OnError(func(ctx *component.ActivationContext) error {
			recoveryAttempted = true
			return nil
		})

		h.AfterActivation(func(ctx *component.ActivationContext) error {
			// Check if error occurred and no output was produced
			if ctx.Result.IsError() &&
				ctx.Component.OutputByName("out").Signals().IsEmpty() {
				fallbackValueProvided = true
			}
			return nil
		})
	}).SetActivationFunc(func(c *component.Component) error {
		val := c.InputByName("in").Signals().First().PayloadOrDefault(0).(int)
		if val == 0 {
			return errors.New("division by zero")
		}
		return c.OutputByName("out").PutSignals(signal.New(100 / val))
	})

	require.NoError(t, c.InputByName("in").PutSignals(signal.New(0)))
	result := c.MaybeActivate()

	require.True(t, result.IsError())
	assert.True(t, recoveryAttempted)
	assert.True(t, fallbackValueProvided)
}

func TestComponentHooks_PracticalInputOutputInspection(t *testing.T) {
	// Practical example: Inspect input/output relationship for debugging
	type ActivationTrace struct {
		InputCount  int
		OutputCount int
		InputValues []int
		OutputSum   int
	}
	var trace ActivationTrace

	c := mustComponent("aggregator",
		component.WithInputs("numbers"),
		component.WithOutputs("sum"),
	).SetupHooks(func(h *component.Hooks) {
		h.BeforeActivation(func(c *component.Component) error {
			// Capture input state
			trace.InputCount = c.InputByName("numbers").Signals().Len()
			_, err := c.InputByName("numbers").Signals().ForEach(func(s *signal.Signal) error {
				trace.InputValues = append(trace.InputValues, s.PayloadOrDefault(0).(int))
				return nil
			})
			return err
		})

		h.AfterActivation(func(ctx *component.ActivationContext) error {
			// Capture output state
			trace.OutputCount = ctx.Component.OutputByName("sum").Signals().Len()
			if trace.OutputCount > 0 {
				trace.OutputSum = ctx.Component.OutputByName("sum").
					Signals().First().PayloadOrDefault(0).(int)
			}
			return nil
		})
	}).SetActivationFunc(func(c *component.Component) error {
		sum := 0
		if _, err := c.InputByName("numbers").Signals().ForEach(func(s *signal.Signal) error {
			sum += s.PayloadOrDefault(0).(int)
			return nil
		}); err != nil {
			return err
		}
		return c.OutputByName("sum").PutSignals(signal.New(sum))
	})

	require.NoError(t, c.InputByName("numbers").PutSignals(signal.New(10), signal.New(20), signal.New(30)))
	c.MaybeActivate()

	assert.Equal(t, 3, trace.InputCount)
	assert.Equal(t, []int{10, 20, 30}, trace.InputValues)
	assert.Equal(t, 1, trace.OutputCount)
	assert.Equal(t, 60, trace.OutputSum)
}
