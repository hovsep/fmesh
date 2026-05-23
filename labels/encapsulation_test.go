package labels

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Contract tests: Collection.All must return a defensive copy of the internal map,
// not a live reference. The collection itself is mutable by design; this file
// guards only against callers reaching inside and corrupting internal state.

func TestLabelsCollection_All_returnsDefensiveCopy(t *testing.T) {
	c := NewCollection().Add("k", "v")
	m, err := c.All()
	require.NoError(t, err)

	m["k"] = "mutated"

	assert.True(t, c.ValueIs("k", "v"),
		"mutating the map from All() must not change the collection (#203)")
}

func TestLabelsCollection_All_map_not_shared_with_AddMany(t *testing.T) {
	c1 := NewCollection().AddMany(map[string]string{"a": "1", "b": "2"})
	m, err := c1.All()
	require.NoError(t, err)

	c2 := NewCollection().AddMany(m)

	m["a"] = "hacked"
	m["b"] = "hacked"

	assert.True(t, c1.ValueIs("a", "1"), "c1 must not observe mutation of map returned from All()")
	assert.True(t, c1.ValueIs("b", "2"))
	assert.True(t, c2.ValueIs("a", "1"), "c2 built from All() map must not share live map with c1")
	assert.True(t, c2.ValueIs("b", "2"))
}
