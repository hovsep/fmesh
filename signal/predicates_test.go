package signal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNot(t *testing.T) {
	t.Parallel()
	alwaysTrue := func(s *Signal) bool { return true }
	alwaysFalse := func(s *Signal) bool { return false }

	assert.False(t, Not(alwaysTrue)(New(1)))
	assert.True(t, Not(alwaysFalse)(New(1)))
}

func TestAnd(t *testing.T) {
	t.Parallel()
	isPositive := func(s *Signal) bool {
		v := s.Payload()
		return v.(int) > 0
	}
	isEven := func(s *Signal) bool {
		v := s.Payload()
		return v.(int)%2 == 0
	}

	assert.True(t, And(isPositive, isEven)(New(4)))
	assert.False(t, And(isPositive, isEven)(New(3)))
	assert.False(t, And(isPositive, isEven)(New(-2)))
}

func TestOr(t *testing.T) {
	t.Parallel()
	isZero := func(s *Signal) bool {
		v := s.Payload()
		return v.(int) == 0
	}
	isNeg := func(s *Signal) bool {
		v := s.Payload()
		return v.(int) < 0
	}

	assert.True(t, Or(isZero, isNeg)(New(0)))
	assert.True(t, Or(isZero, isNeg)(New(-5)))
	assert.False(t, Or(isZero, isNeg)(New(1)))
}

func TestMetaEquals(t *testing.T) {
	t.Parallel()
	s := New(1).WithMeta("env", "prod")

	assert.True(t, MetaEquals("env", "prod")(s))
	assert.False(t, MetaEquals("env", "staging")(s))
	assert.False(t, MetaEquals("region", "us")(s))
}

func TestMetaContains(t *testing.T) {
	t.Parallel()
	s := New(1).WithMeta("tag", "urgent-request")

	assert.True(t, MetaContains("tag", "urgent")(s))
	assert.True(t, MetaContains("tag", "request")(s))
	assert.False(t, MetaContains("tag", "critical")(s))
	assert.False(t, MetaContains("missing", "x")(s))
}

func TestHasMeta(t *testing.T) {
	t.Parallel()
	s := New(1).WithMetaMany(map[string]string{"a": "1", "b": "2", "c": "3"})

	assert.True(t, HasMeta("a")(s))
	assert.False(t, HasMeta("z")(s))
	assert.True(t, HasMeta("a", "b")(s))
	assert.True(t, HasMeta("a", "b", "c")(s))
	assert.False(t, HasMeta("a", "d")(s))
	assert.True(t, HasMeta()(s)) // vacuous
}

func TestHasAnyMeta(t *testing.T) {
	t.Parallel()
	s := New(1).WithMetaMany(map[string]string{"a": "1", "b": "2"})

	assert.True(t, HasAnyMeta("a", "z")(s))
	assert.True(t, HasAnyMeta("b")(s))
	assert.False(t, HasAnyMeta("x", "y")(s))
}

// One key space holds both types, so a predicate typed for one must not match
// an entry of the other: "1" is not 1.0, and a float never "contains" a digit.
func TestMetaPredicates_typeMismatch(t *testing.T) {
	t.Parallel()
	s := New(1).WithMeta("n", 1.0).WithMeta("tag", "1")

	assert.True(t, MetaEquals("n", 1.0)(s))
	assert.False(t, MetaEquals("n", "1")(s))
	assert.False(t, MetaEquals("tag", 1.0)(s))
	assert.False(t, MetaContains("n", "1")(s))
}

// A zero-value signal has a nil store; predicates must read it as empty, not panic.
func TestMetaPredicates_zeroValueSignal(t *testing.T) {
	t.Parallel()
	s := &Signal{}

	assert.False(t, HasMeta("a")(s))
	assert.False(t, HasAnyMeta("a")(s))
	assert.False(t, MetaEquals("a", "b")(s))
	assert.False(t, MetaContains("a", "b")(s))
}

func TestPredicateCombinators_composition(t *testing.T) {
	t.Parallel()
	g := NewGroup(1, 2, 3, 4, 5, 6).Map(func(s *Signal) *Signal {
		v := s.Payload()
		if v.(int)%2 == 0 {
			return s.WithMeta("even", "true")
		}
		return s.WithMeta("odd", "true")
	})

	// Keep only even signals using combinator
	evens := g.Filter(HasMeta("even"))
	assert.Equal(t, 3, evens.Len())

	// Keep odd signals via Not
	odds := g.Filter(Not(HasMeta("even")))
	assert.Equal(t, 3, odds.Len())

	// And: even AND payload > 3  → 4, 6
	bigEvens := g.Filter(And(HasMeta("even"), func(s *Signal) bool {
		v := s.Payload()
		return v.(int) > 3
	}))
	assert.Equal(t, 2, bigEvens.Len())

	// Or: has "odd" entry OR payload == 6  → 1,3,5,6
	oddOrSix := g.Filter(Or(HasMeta("odd"), func(s *Signal) bool {
		v := s.Payload()
		return v.(int) == 6
	}))
	assert.Equal(t, 4, oddOrSix.Len())
}
