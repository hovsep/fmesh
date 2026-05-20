package signal

import (
	"testing"

	"github.com/hovsep/fmesh/labels"
	"github.com/stretchr/testify/assert"
)

func TestNot(t *testing.T) {
	alwaysTrue := func(s *Signal) bool { return true }
	alwaysFalse := func(s *Signal) bool { return false }

	assert.False(t, Not(alwaysTrue)(New(1)))
	assert.True(t, Not(alwaysFalse)(New(1)))
}

func TestAnd(t *testing.T) {
	isPositive := func(s *Signal) bool {
		v, _ := s.Payload()
		return v.(int) > 0
	}
	isEven := func(s *Signal) bool {
		v, _ := s.Payload()
		return v.(int)%2 == 0
	}

	assert.True(t, And(isPositive, isEven)(New(4)))
	assert.False(t, And(isPositive, isEven)(New(3)))
	assert.False(t, And(isPositive, isEven)(New(-2)))
}

func TestOr(t *testing.T) {
	isZero := func(s *Signal) bool {
		v, _ := s.Payload()
		return v.(int) == 0
	}
	isNeg := func(s *Signal) bool {
		v, _ := s.Payload()
		return v.(int) < 0
	}

	assert.True(t, Or(isZero, isNeg)(New(0)))
	assert.True(t, Or(isZero, isNeg)(New(-5)))
	assert.False(t, Or(isZero, isNeg)(New(1)))
}

func TestHasLabel(t *testing.T) {
	s := New(1).WithLabel("env", "prod")

	assert.True(t, HasLabel("env")(s))
	assert.False(t, HasLabel("region")(s))
}

func TestLabelEquals(t *testing.T) {
	s := New(1).WithLabel("env", "prod")

	assert.True(t, LabelEquals("env", "prod")(s))
	assert.False(t, LabelEquals("env", "staging")(s))
	assert.False(t, LabelEquals("region", "us")(s))
}

func TestLabelContains(t *testing.T) {
	s := New(1).WithLabel("tag", "urgent-request")

	assert.True(t, LabelContains("tag", "urgent")(s))
	assert.True(t, LabelContains("tag", "request")(s))
	assert.False(t, LabelContains("tag", "critical")(s))
	assert.False(t, LabelContains("missing", "x")(s))
}

func TestHasAllLabels(t *testing.T) {
	s := New(1).WithLabels(labels.Map{"a": "1", "b": "2", "c": "3"})

	assert.True(t, HasAllLabels("a", "b")(s))
	assert.True(t, HasAllLabels("a", "b", "c")(s))
	assert.False(t, HasAllLabels("a", "d")(s))
	assert.True(t, HasAllLabels()(s)) // vacuous
}

func TestHasAnyLabel(t *testing.T) {
	s := New(1).WithLabels(labels.Map{"a": "1", "b": "2"})

	assert.True(t, HasAnyLabel("a", "z")(s))
	assert.True(t, HasAnyLabel("b")(s))
	assert.False(t, HasAnyLabel("x", "y")(s))
}

func TestPredicateCombinators_composition(t *testing.T) {
	g := NewGroup(1, 2, 3, 4, 5, 6).Map(func(s *Signal) *Signal {
		v, _ := s.Payload()
		if v.(int)%2 == 0 {
			return s.WithLabel("even", "true")
		}
		return s.WithLabel("odd", "true")
	})

	// Keep only even signals using combinator
	evens := g.Filter(HasLabel("even"))
	assert.Equal(t, 3, evens.Len())

	// Keep odd signals via Not
	odds := g.Filter(Not(HasLabel("even")))
	assert.Equal(t, 3, odds.Len())

	// And: even AND payload > 3  → 4, 6
	bigEvens := g.Filter(And(HasLabel("even"), func(s *Signal) bool {
		v, _ := s.Payload()
		return v.(int) > 3
	}))
	assert.Equal(t, 2, bigEvens.Len())

	// Or: has "odd" label OR payload == 6  → 1,3,5,6
	oddOrSix := g.Filter(Or(HasLabel("odd"), func(s *Signal) bool {
		v, _ := s.Payload()
		return v.(int) == 6
	}))
	assert.Equal(t, 4, oddOrSix.Len())
}
