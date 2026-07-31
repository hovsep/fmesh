package signal

import (
	"errors"
	"fmt"
)

// Typed reads of a payload, which is an any because a mesh carries mixed types
// down one pipe. Two shapes: one reports what went wrong, one carries on with a
// default. Neither panics.
//
// Reach for As. AsOrDefault is the one that turns "upstream now sends int64" into
// a silent zero — the mismatch and the genuinely-absent value are indistinguishable
// in its result, and nothing anywhere else in the mesh will mention it. Use it only
// where a fallback is the correct answer rather than the convenient one.

// As returns the payload as T, failing rather than panicking when the payload is
// another type — a component cannot know that something upstream changed its
// payload type, and taking down the mesh is not a useful way to be told.
func As[T any](s *Signal) (T, error) {
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

// AsOrDefault returns the payload as T, or the default if it is missing or of
// another type.
//
// It cannot tell those two cases apart, and says nothing when it substitutes: a
// component whose upstream changed from int to int64 keeps computing, on zero.
// Prefer As unless the default is genuinely the right answer for a wrong type.
//
// T is inferred from defaultValue, so an untyped 0 means int: pass 0.0, or use
// AsFloat64OrDefault, when the payload is a float64.
func AsOrDefault[T any](s *Signal, defaultValue T) T {
	if s == nil {
		return defaultValue
	}

	value, ok := s.Payload().(T)
	if !ok {
		return defaultValue
	}
	return value
}

// AsFloat64 returns the payload as a float64.
func AsFloat64(s *Signal) (float64, error) { return As[float64](s) }

// AsFloat64OrDefault returns the payload as a float64, or the default. The other
// types need no such shorthand: AsOrDefault(s, 0) already infers int, while
// AsOrDefault(s, 0) for a float64 payload infers int and returns the default.
func AsFloat64OrDefault(s *Signal, defaultValue float64) float64 {
	return AsOrDefault(s, defaultValue)
}

// AsInt returns the payload as an int.
func AsInt(s *Signal) (int, error) { return As[int](s) }

// AsString returns the payload as a string.
func AsString(s *Signal) (string, error) { return As[string](s) }

// AsBool returns the payload as a bool.
func AsBool(s *Signal) (bool, error) { return As[bool](s) }

// AsGroup returns the payload as a group, for a signal carrying other signals.
func AsGroup(s *Signal) (*Group, error) { return As[*Group](s) }

// AsNumber reports the payload as a float64 when it carries float64, float32,
// int, int64 or uint64, or a bool as 1 and 0. Narrower integer types are not
// covered: widen at the source rather than adding cases here.
//
// Loose on purpose: it answers "is this a measurement at all" for code that does
// not know which components produce which payloads.
func AsNumber(s *Signal) (float64, bool) {
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
