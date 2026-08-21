package signal

import (
	"errors"
	"fmt"
)

// Typed payload accessors: As reports a wrong or missing type, PayloadOrDefault
// substitutes a default. Neither panics, and all are safe on a nil receiver,
// which is what Group.First returns for an empty group. Prefer As — see
// PayloadOrDefault's footgun.

// As returns the payload as T, failing rather than panicking when the payload is
// another type — a component cannot know that something upstream changed its
// payload type, and taking down the mesh is not a useful way to be told.
func (s *Signal) As[T any]() (T, error) {
	var zero T
	if s == nil {
		return zero, errors.New("signal is nil")
	}

	payload := s.Payload()
	typed, ok := payload.(T)
	if !ok {
		return zero, fmt.Errorf("signal payload is %T, not %T", payload, zero)
	}
	return typed, nil
}

// PayloadOrDefault returns the payload as T, or the default if it is missing or
// of another type.
//
// It cannot tell those two cases apart, and says nothing when it substitutes: a
// component whose upstream changed from int to int64 keeps computing, on zero.
// Prefer As unless the default is genuinely the right answer for a wrong type.
//
// T is inferred from defaultValue, so an untyped 0 means int: pass 0.0, or use
// Float64OrDefault, when the payload is a float64.
func (s *Signal) PayloadOrDefault[T any](defaultValue T) T {
	if s == nil {
		return defaultValue
	}

	value, ok := s.Payload().(T)
	if !ok {
		return defaultValue
	}
	return value
}

// Float64OrDefault returns the payload as a float64, or the default. It is the
// one per-type shorthand kept: PayloadOrDefault infers T from the default, and
// an untyped 0 makes that int, so s.PayloadOrDefault(0) on a float64 payload
// silently returns the default. Every other type spells out as s.As[T]().
func (s *Signal) Float64OrDefault(defaultValue float64) float64 {
	return s.PayloadOrDefault(defaultValue)
}

// AsGroup returns the payload as a group, for a signal carrying other signals.
func (s *Signal) AsGroup() (*Group, error) { return s.As[*Group]() }

// AsNumber reports the payload as a float64 when it carries float64, float32,
// int, int64 or uint64, or a bool as 1 and 0. Narrower integer types are not
// covered: widen at the source rather than adding cases here.
//
// Loose on purpose: it answers "is this a measurement at all" for code that does
// not know which components produce which payloads.
func (s *Signal) AsNumber() (float64, bool) {
	if s == nil {
		return 0, false
	}

	switch v := s.Payload().(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	case uint64:
		return float64(v), true
	case bool:
		if v {
			return 1, true
		}
		return 0, true
	default:
		return 0, false
	}
}
