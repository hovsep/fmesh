package plugin

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// target is a stand-in for whatever a plugin initializes (*FMesh, *Component).
type target struct {
	seen []string
}

type recorder struct {
	name string
	err  error
}

func (r recorder) GetName() string { return r.name }

func (r recorder) Init(t *target) error {
	t.seen = append(t.seen, r.name)
	return r.err
}

func TestRegistry(t *testing.T) {
	t.Run("empty registry initializes nothing", func(t *testing.T) {
		tgt := &target{}
		r := NewRegistry[*target]()

		require.NoError(t, r.InitAll(tgt))
		assert.Empty(t, tgt.seen)
		assert.False(t, r.Has("anything"))
	})

	t.Run("plugins initialize in name order", func(t *testing.T) {
		// Plugins register hooks and hooks fire in registration order, so this
		// order is observable behavior. Ranging the map would make it vary
		// between runs of the same binary.
		tgt := &target{}
		r := NewRegistry[*target]()
		require.NoError(t, r.Add(recorder{name: "zulu"}, recorder{name: "alpha"}, recorder{name: "mike"}))

		require.NoError(t, r.InitAll(tgt))
		assert.Equal(t, []string{"alpha", "mike", "zulu"}, tgt.seen)
	})

	t.Run("a registered plugin is queryable by name", func(t *testing.T) {
		r := NewRegistry[*target]()
		require.NoError(t, r.Add(recorder{name: "alpha"}))

		assert.True(t, r.Has("alpha"))
		assert.False(t, r.Has("Alpha"), "names are compared exactly")
	})

	t.Run("a duplicate name is rejected", func(t *testing.T) {
		r := NewRegistry[*target]()
		require.NoError(t, r.Add(recorder{name: "alpha"}))

		err := r.Add(recorder{name: "alpha"})
		require.Error(t, err)
		assert.ErrorContains(t, err, "plugin alpha already registered")
	})

	t.Run("a failing plugin stops initialization", func(t *testing.T) {
		tgt := &target{}
		r := NewRegistry[*target]()
		require.NoError(t, r.Add(
			recorder{name: "alpha"},
			recorder{name: "mike", err: errors.New("no")},
			recorder{name: "zulu"},
		))

		err := r.InitAll(tgt)
		require.Error(t, err)
		require.ErrorContains(t, err, "plugin mike initialization failed: no")
		assert.Equal(t, []string{"alpha", "mike"}, tgt.seen, "zulu never ran")
	})
}
