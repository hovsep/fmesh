package component

import (
	"fmt"

	"github.com/hovsep/fmesh/signal"
)

// Combinators over ActivationFunc: they build the single function a component is
// constructed with. Use OnActivation hooks instead when the behavior is added
// from outside the component rather than composed by it.

// Sequential runs activation functions in order, stopping at the first error.
//
// Errors pass through as the component's own, so a stage returning
// ErrWaitingForInputs suspends the component.
func Sequential(funcs ...ActivationFunc) ActivationFunc {
	return func(this *Component) error {
		for _, f := range funcs {
			if err := f(this); err != nil {
				return err
			}
		}
		return nil
	}
}

// When runs fn only if the predicate holds.
func When(predicate func(*Component) bool, fn ActivationFunc) ActivationFunc {
	return func(this *Component) error {
		if !predicate(this) {
			return nil
		}
		return fn(this)
	}
}

// HasSignalsOn reports whether every named input port exists and carries
// something. A name no port has is never ready; RequireInputs reports it.
func HasSignalsOn(portNames ...string) func(*Component) bool {
	return func(this *Component) bool {
		for _, name := range portNames {
			p := this.InputByName(name)
			if p == nil || !p.HasSignals() {
				return false
			}
		}
		return true
	}
}

// RequireInputs suspends the component until every named port has signals,
// keeping whatever has already arrived. Contrast When, which skips the
// activation and so lets the partial inputs be cleared.
//
// A name no port has fails the activation instead of suspending it forever, and
// every name is checked before any signal, or an empty port hides the typo
// behind a wait.
func RequireInputs(portNames ...string) ActivationFunc {
	return func(this *Component) error {
		for _, name := range portNames {
			if this.InputByName(name) == nil {
				return fmt.Errorf("required input port %q does not exist", name)
			}
		}

		if !this.Inputs().ByNames(portNames...).AllHaveSignals() {
			return ErrWaitingForInputsKeep
		}
		return nil
	}
}

// PipelineStage transforms a group of signals on its way through a pipeline.
type PipelineStage func(signals *signal.Group) (*signal.Group, error)

// Pipeline reads the named input ports in the order given, passes their signals
// through each stage in turn, and writes the result to one output port.
//
// A port name no port has, and a nil group from a stage, are errors rather than
// nil dereferences. A stage that drops everything returns an empty group.
func Pipeline(inputPortNames []string, outputPortName string, stages ...PipelineStage) ActivationFunc {
	return func(this *Component) error {
		out := this.OutputByName(outputPortName)
		if out == nil {
			return fmt.Errorf("pipeline output port %q does not exist", outputPortName)
		}

		signals := signal.NewGroup()
		for _, name := range inputPortNames {
			in := this.InputByName(name)
			if in == nil {
				return fmt.Errorf("pipeline input port %q does not exist", name)
			}
			signals = signals.With(in.Signals().All()...)
		}

		for i, stage := range stages {
			next, err := stage(signals)
			if err != nil {
				return fmt.Errorf("pipeline stage %d failed: %w", i, err)
			}
			if next == nil {
				return fmt.Errorf("pipeline stage %d returned a nil group", i)
			}
			signals = next
		}

		return out.PutSignalGroups(signals)
	}
}
