package signal

import "strings"

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

// HasLabel returns a predicate that is true when the signal has a label with the given name.
func HasLabel(name string) Predicate {
	return func(s *Signal) bool {
		return s.Labels().Has(name)
	}
}

// LabelEquals returns a predicate that is true when the signal has a label with the given name and exact value.
func LabelEquals(name, value string) Predicate {
	return func(s *Signal) bool {
		return s.Labels().ValueIs(name, value)
	}
}

// LabelContains returns a predicate that is true when the signal has a label with the given name
// and the label's value contains the given substring.
func LabelContains(name, substr string) Predicate {
	return func(s *Signal) bool {
		v, err := s.Labels().Value(name)
		if err != nil {
			return false
		}
		return strings.Contains(v, substr)
	}
}

// HasAllLabels returns a predicate that is true when the signal has all of the given label names.
func HasAllLabels(names ...string) Predicate {
	return func(s *Signal) bool {
		return s.Labels().HasAll(names...)
	}
}

// HasAnyLabel returns a predicate that is true when the signal has at least one of the given label names.
func HasAnyLabel(names ...string) Predicate {
	return func(s *Signal) bool {
		return s.Labels().HasAny(names...)
	}
}
