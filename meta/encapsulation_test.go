package meta

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Contract tests: Labels.All must return a defensive copy of the internal map,
// not a live reference. The collection itself is mutable by design; this file
// guards only against callers reaching inside and corrupting internal state.

func TestLabelsCollection_All_returnsDefensiveCopy(t *testing.T) {
	c := NewLabels().Set("k", "v")
	m := c.All()

	m["k"] = "mutated"

	assert.True(t, c.ValueIs("k", "v"),
		"mutating the map from All() must not change the collection (#203)")
}

func TestLabelsCollection_All_map_not_shared_with_AddMany(t *testing.T) {
	c1 := NewLabels().SetMany(map[string]string{"a": "1", "b": "2"})
	m := c1.All()

	c2 := NewLabels().SetMany(m)

	m["a"] = "hacked"
	m["b"] = "hacked"

	assert.True(t, c1.ValueIs("a", "1"), "c1 must not observe mutation of map returned from All()")
	assert.True(t, c1.ValueIs("b", "2"))
	assert.True(t, c2.ValueIs("a", "1"), "c2 built from All() map must not share live map with c1")
	assert.True(t, c2.ValueIs("b", "2"))
}
