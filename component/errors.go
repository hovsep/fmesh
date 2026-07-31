package component

import (
	"errors"
	"fmt"
)

// These are control-flow signals, not actual failures.
// They instruct the scheduler how to proceed with the current component.
//
// For now we only have two variants, so sentinel errors are sufficient.
// If more behaviors are introduced, consider switching to a typed error.
var (
	// ErrWaitingForInputs is the sentinel both waiting modes wrap. Test for it with
	// errors.Is when you want to know "is this component waiting" regardless of mode.
	// Returning it directly still works and means the same as ErrWaitDroppingInputs.
	ErrWaitingForInputs = errors.New("component is waiting for some inputs")

	// ErrWaitDroppingInputs suspends the component and **discards** the input signals
	// that arrived this cycle. Correct when the partial input is worthless without the
	// rest — a request whose response never came.
	ErrWaitDroppingInputs = fmt.Errorf("%w: dropping the signals received this cycle", ErrWaitingForInputs)

	// ErrWaitKeepingInputs suspends the component and **keeps** its input signals for
	// the next cycle. Correct when the component is accumulating — a join waiting for
	// its second operand.
	//
	// Both modes are named, because which one is right depends on the component and
	// the wrong one loses data silently.
	ErrWaitKeepingInputs = fmt.Errorf("%w: keeping the signals for the next cycle", ErrWaitingForInputs)
)
