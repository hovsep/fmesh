package signal

import (
	"strings"

	"github.com/hovsep/fmesh/meta"
)

// Not returns a predicate that is the logical negation of p.
func Not(p Predicate) Predicate {
	return func(s *Signal) bool {
		return !p(s)
	}
}

// And returns a predicate that is true only when both p1 and p2 are true.
func And(p1, p2 Predicate) Predicate {
	return func(s *Signal) bool {
		return p1(s) && p2(s)
	}
}

// Or returns a predicate that is true when at least one of p1 or p2 is true.
func Or(p1, p2 Predicate) Predicate {
	return func(s *Signal) bool {
		return p1(s) || p2(s)
	}
}

// HasMeta returns a predicate that is true when the signal carries every named metadata key.
func HasMeta(keys ...string) Predicate {
	return func(s *Signal) bool {
		return s.Meta().Has(keys...)
	}
}

// HasAnyMeta returns a predicate that is true when the signal carries at least one of the keys.
func HasAnyMeta(keys ...string) Predicate {
	return func(s *Signal) bool {
		return s.Meta().HasAny(keys...)
	}
}

// MetaEquals returns a predicate that is true when the signal's metadata holds exactly value under key.
func MetaEquals[T meta.Value](key string, value T) Predicate {
	return func(s *Signal) bool {
		return s.Meta().ValueIs(key, value)
	}
}

// MetaContains returns a predicate that is true when key holds a string containing substr.
func MetaContains(key, substr string) Predicate {
	return func(s *Signal) bool {
		v, err := s.Meta().Value[string](key)
		return err == nil && strings.Contains(v, substr)
	}
}
