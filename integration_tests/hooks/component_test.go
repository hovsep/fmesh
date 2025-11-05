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

	c := component.New("validator").
		AddInputs("data").
		SetupHooks(func(h *component.Hooks) {
			h.OnError(func(ctx *component.ActivationContext) {
				// Access the error and log it with component context
				errorLog = ErrorLog{
					ComponentName: ctx.Component.Name(),
					ErrorMessage:  ctx.Result.ActivationError().Error(),
					ErrorType:     "validation_error",
				}
			})
		}).
		WithActivationFunc(func(c *component.Component) error {
			// Simulate validation logic
			inputVal := c.InputByName("data").Signals().First().PayloadOrDefault(0).(int)
			if inputVal < 0 {
				return validationErr
			}
			return nil
		})

	c.InputByName("data").PutSignals(signal.New(-5))
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

	c := component.New("calculator").
		AddInputs("x", "y").
		AddOutputs("result").
		SetupHooks(func(h *component.Hooks) {
			h.OnSuccess(func(ctx *component.ActivationContext) {
				// Access and validate output signals
				resultPort := ctx.Component.OutputByName("result")
				if resultPort.Signals().Len() == 1 {
					val := resultPort.Signals().First().PayloadOrDefault(0).(int)
					outputValue = val
					outputIsValid = val >= 0 && val <= 100 // Valid range check
				}
			})
		}).
		WithActivationFunc(func(c *component.Component) error {
			x := c.InputByName("x").Signals().First().PayloadOrDefault(0).(int)
			y := c.InputByName("y").Signals().First().PayloadOrDefault(0).(int)
			result := x + y
			c.OutputByName("result").PutSignals(signal.New(result))
			return nil
		})

	c.InputByName("x").PutSignals(signal.New(30))
	c.InputByName("y").PutSignals(signal.New(20))
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

	c := component.New("processor").
		AddInputs("in").
		AddOutputs("success", "failure").
		SetupHooks(func(h *component.Hooks) {
			h.OnSuccess(func(ctx *component.ActivationContext) {
				metrics.SuccessCount++
				// Track output signal counts
				ctx.Component.Outputs().ForEach(func(p *port.Port) {
					metrics.OutputSignalCounts[p.Name()] = p.Signals().Len()
				})
			})

			h.OnError(func(ctx *component.ActivationContext) {
				metrics.ErrorCount++
				metrics.LastError = ctx.Result.ActivationError()
			})

			h.OnPanic(func(ctx *component.ActivationContext) {
				metrics.PanicCount++
			})

			h.AfterActivation(func(ctx *component.ActivationContext) {
				metrics.TotalActivations++
			})
		}).
		WithActivationFunc(func(c *component.Component) error {
			val := c.InputByName("in").Signals().First().PayloadOrDefault(0).(int)
			if val > 0 {
				c.OutputByName("success").PutSignals(signal.New(val * 2))
				return nil
			}
			return errors.New("invalid input")
		})

	// First activation: success
	c.InputByName("in").PutSignals(signal.New(5))
	c.MaybeActivate()

	// Second activation: error
	c.InputByName("in").Clear()
	c.InputByName("in").PutSignals(signal.New(-1))
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
	var enrichedOutput []map[string]interface{}

	c := component.New("enricher").
		AddInputs("raw").
		AddOutputs("enriched").
		SetupHooks(func(h *component.Hooks) {
			h.OnSuccess(func(ctx *component.ActivationContext) {
				// Access output and create enriched version with metadata
				ctx.Component.OutputByName("enriched").Signals().ForEach(func(s *signal.Signal) {
					enriched := map[string]interface{}{
						"value":         s.PayloadOrDefault(nil),
						"component":     ctx.Component.Name(),
						"timestamp":     "2024-01-01", // In real code, use time.Now()
						"activation_ok": ctx.Result.Code() == component.ActivationCodeOK,
					}
					enrichedOutput = append(enrichedOutput, enriched)
				})
			})
		}).
		WithActivationFunc(func(c *component.Component) error {
			// Process and output data
			c.InputByName("raw").Signals().ForEach(func(s *signal.Signal) {
				processed := s.PayloadOrDefault(0).(int) * 10
				c.OutputByName("enriched").PutSignals(signal.New(processed))
			})
			return nil
		})

	c.InputByName("raw").PutSignals(signal.New(3), signal.New(7))
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

	c := component.New("resilient").
		AddInputs("in").
		AddOutputs("out").
		SetupHooks(func(h *component.Hooks) {
			h.OnError(func(ctx *component.ActivationContext) {
				recoveryAttempted = true
				// In a real scenario, you might log, alert, or trigger retry logic
				// Note: Can't modify output in hook, but can trigger side effects
			})

			h.AfterActivation(func(ctx *component.ActivationContext) {
				// Check if error occurred and no output was produced
				if ctx.Result.IsError() &&
					ctx.Component.OutputByName("out").Signals().IsEmpty() {
					// Could trigger fallback mechanism, send to dead letter queue, etc.
					fallbackValueProvided = true
				}
			})
		}).
		WithActivationFunc(func(c *component.Component) error {
			val := c.InputByName("in").Signals().First().PayloadOrDefault(0).(int)
			if val == 0 {
				return errors.New("division by zero")
			}
			c.OutputByName("out").PutSignals(signal.New(100 / val))
			return nil
		})

	c.InputByName("in").PutSignals(signal.New(0))
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

	c := component.New("aggregator").
		AddInputs("numbers").
		AddOutputs("sum").
		SetupHooks(func(h *component.Hooks) {
			h.BeforeActivation(func(c *component.Component) {
				// Capture input state
				trace.InputCount = c.InputByName("numbers").Signals().Len()
				c.InputByName("numbers").Signals().ForEach(func(s *signal.Signal) {
					trace.InputValues = append(trace.InputValues, s.PayloadOrDefault(0).(int))
				})
			})

			h.AfterActivation(func(ctx *component.ActivationContext) {
				// Capture output state
				trace.OutputCount = ctx.Component.OutputByName("sum").Signals().Len()
				if trace.OutputCount > 0 {
					trace.OutputSum = ctx.Component.OutputByName("sum").
						Signals().First().PayloadOrDefault(0).(int)
				}
			})
		}).
		WithActivationFunc(func(c *component.Component) error {
			sum := 0
			c.InputByName("numbers").Signals().ForEach(func(s *signal.Signal) {
				sum += s.PayloadOrDefault(0).(int)
			})
			c.OutputByName("sum").PutSignals(signal.New(sum))
			return nil
		})

	c.InputByName("numbers").PutSignals(signal.New(10), signal.New(20), signal.New(30))
	c.MaybeActivate()

	assert.Equal(t, 3, trace.InputCount)
	assert.Equal(t, []int{10, 20, 30}, trace.InputValues)
	assert.Equal(t, 1, trace.OutputCount)
	assert.Equal(t, 60, trace.OutputSum)
}
