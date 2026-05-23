package component

import (
	"testing"

	"github.com/hovsep/fmesh/labels"
	"github.com/hovsep/fmesh/port"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewComponent(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "empty name is valid",
			args: args{name: ""},
		},
		{
			name: "with name",
			args: args{name: "multiplier"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c1, err := New(tt.args.name)
			require.NoError(t, err)
			c2, err := New(tt.args.name)
			require.NoError(t, err)
			assert.Equal(t, c1, c2)
		})
	}
}

func TestComponent_WithDescription(t *testing.T) {
	type args struct {
		description string
	}
	tests := []struct {
		name      string
		component *Component
		args      args
		want      *Component
	}{
		{
			name:      "happy path",
			component: mustNew("c1"),
			args: args{
				description: "descr",
			},
			want: &Component{
				name:        "c1",
				description: "descr",
				labels:      labels.NewCollection(),
				inputPorts:  port.NewCollection(),
				outputPorts: port.NewCollection(),
				f:           nil,
				state:       NewState(),
				hooks:       NewHooks(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.component.WithDescription(tt.args.description))
		})
	}
}

func TestComponent_SetLabels(t *testing.T) {
	tests := []struct {
		name       string
		component  *Component
		labels     labels.Map
		assertions func(t *testing.T, component *Component)
	}{
		{
			name:      "set labels on new component",
			component: mustNew("c1"),
			labels: labels.Map{
				"l1": "v1",
				"l2": "v2",
			},
			assertions: func(t *testing.T, component *Component) {
				assert.Equal(t, 2, component.Labels().Len())
				assert.True(t, component.labels.HasAll("l1", "l2"))
			},
		},
		{
			name:      "set labels replaces existing labels",
			component: mustNew("c1").AddLabels(labels.Map{"old": "value"}),
			labels: labels.Map{
				"l1": "v1",
				"l2": "v2",
			},
			assertions: func(t *testing.T, component *Component) {
				assert.Equal(t, 2, component.Labels().Len())
				assert.True(t, component.labels.HasAll("l1", "l2"))
				assert.False(t, component.labels.Has("old"))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			componentAfter := tt.component.SetLabels(tt.labels)
			if tt.assertions != nil {
				tt.assertions(t, componentAfter)
			}
		})
	}
}

func TestComponent_AddLabels(t *testing.T) {
	tests := []struct {
		name       string
		component  *Component
		labels     labels.Map
		assertions func(t *testing.T, component *Component)
	}{
		{
			name:      "add labels to new component",
			component: mustNew("c1"),
			labels: labels.Map{
				"l1": "v1",
				"l2": "v2",
			},
			assertions: func(t *testing.T, component *Component) {
				assert.Equal(t, 2, component.Labels().Len())
				assert.True(t, component.labels.HasAll("l1", "l2"))
			},
		},
		{
			name:      "add labels merges with existing",
			component: mustNew("c1").AddLabels(labels.Map{"existing": "label"}),
			labels: labels.Map{
				"l1": "v1",
				"l2": "v2",
			},
			assertions: func(t *testing.T, component *Component) {
				assert.Equal(t, 3, component.Labels().Len())
				assert.True(t, component.labels.HasAll("existing", "l1", "l2"))
			},
		},
		{
			name:      "add labels updates existing key",
			component: mustNew("c1").AddLabels(labels.Map{"l1": "old"}),
			labels: labels.Map{
				"l1": "new",
			},
			assertions: func(t *testing.T, component *Component) {
				assert.Equal(t, 1, component.Labels().Len())
				assert.True(t, component.labels.ValueIs("l1", "new"))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			componentAfter := tt.component.AddLabels(tt.labels)
			if tt.assertions != nil {
				tt.assertions(t, componentAfter)
			}
		})
	}
}

func TestComponent_AddLabel(t *testing.T) {
	tests := []struct {
		name       string
		component  *Component
		labelName  string
		labelValue string
		assertions func(t *testing.T, component *Component)
	}{
		{
			name:       "add single label to new component",
			component:  mustNew("c1"),
			labelName:  "env",
			labelValue: "prod",
			assertions: func(t *testing.T, component *Component) {
				assert.Equal(t, 1, component.Labels().Len())
				assert.True(t, component.labels.ValueIs("env", "prod"))
			},
		},
		{
			name:       "add label merges with existing",
			component:  mustNew("c1").AddLabel("existing", "label"),
			labelName:  "env",
			labelValue: "prod",
			assertions: func(t *testing.T, component *Component) {
				assert.Equal(t, 2, component.Labels().Len())
				assert.True(t, component.labels.HasAll("existing", "env"))
			},
		},
		{
			name:       "add label updates existing key",
			component:  mustNew("c1").AddLabel("env", "dev"),
			labelName:  "env",
			labelValue: "prod",
			assertions: func(t *testing.T, component *Component) {
				assert.Equal(t, 1, component.Labels().Len())
				assert.True(t, component.labels.ValueIs("env", "prod"))
			},
		},
		{
			name:       "chainable",
			component:  mustNew("c1"),
			labelName:  "l1",
			labelValue: "v1",
			assertions: func(t *testing.T, component *Component) {
				result := component.AddLabel("l2", "v2").AddLabel("l3", "v3")
				assert.Equal(t, 3, result.Labels().Len())
				assert.True(t, result.labels.HasAll("l1", "l2", "l3"))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			componentAfter := tt.component.AddLabel(tt.labelName, tt.labelValue)
			if tt.assertions != nil {
				tt.assertions(t, componentAfter)
			}
		})
	}
}

func TestComponent_ClearLabels(t *testing.T) {
	tests := []struct {
		name       string
		component  *Component
		assertions func(t *testing.T, component *Component)
	}{
		{
			name:      "clear labels from component with labels",
			component: mustNew("c1").AddLabels(labels.Map{"k1": "v1", "k2": "v2"}),
			assertions: func(t *testing.T, component *Component) {
				assert.Equal(t, 0, component.Labels().Len())
				assert.False(t, component.Labels().Has("k1"))
				assert.False(t, component.Labels().Has("k2"))
			},
		},
		{
			name:      "clear labels from component without labels",
			component: mustNew("c1"),
			assertions: func(t *testing.T, component *Component) {
				assert.Equal(t, 0, component.Labels().Len())
			},
		},
		{
			name:      "chainable",
			component: mustNew("c1").AddLabels(labels.Map{"k1": "v1"}),
			assertions: func(t *testing.T, component *Component) {
				result := component.ClearLabels().AddLabel("k2", "v2")
				assert.Equal(t, 1, result.Labels().Len())
				assert.False(t, result.Labels().Has("k1"))
				assert.True(t, result.Labels().ValueIs("k2", "v2"))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			componentAfter := tt.component.ClearLabels()
			if tt.assertions != nil {
				tt.assertions(t, componentAfter)
			}
		})
	}
}

func TestComponent_RemoveLabels(t *testing.T) {
	tests := []struct {
		name           string
		component      *Component
		labelsToRemove []string
		assertions     func(t *testing.T, component *Component)
	}{
		{
			name:           "remove single label",
			component:      mustNew("c1").AddLabels(labels.Map{"k1": "v1", "k2": "v2", "k3": "v3"}),
			labelsToRemove: []string{"k1"},
			assertions: func(t *testing.T, component *Component) {
				assert.Equal(t, 2, component.Labels().Len())
				assert.False(t, component.Labels().Has("k1"))
				assert.True(t, component.Labels().Has("k2"))
				assert.True(t, component.Labels().Has("k3"))
			},
		},
		{
			name:           "remove multiple labels",
			component:      mustNew("c1").AddLabels(labels.Map{"k1": "v1", "k2": "v2", "k3": "v3"}),
			labelsToRemove: []string{"k1", "k2"},
			assertions: func(t *testing.T, component *Component) {
				assert.Equal(t, 1, component.Labels().Len())
				assert.False(t, component.Labels().Has("k1"))
				assert.False(t, component.Labels().Has("k2"))
				assert.True(t, component.Labels().ValueIs("k3", "v3"))
			},
		},
		{
			name:           "remove non-existent label",
			component:      mustNew("c1").AddLabels(labels.Map{"k1": "v1"}),
			labelsToRemove: []string{"k2"},
			assertions: func(t *testing.T, component *Component) {
				assert.Equal(t, 1, component.Labels().Len())
				assert.True(t, component.Labels().ValueIs("k1", "v1"))
			},
		},
		{
			name:           "chainable",
			component:      mustNew("c1").AddLabels(labels.Map{"k1": "v1", "k2": "v2", "k3": "v3"}),
			labelsToRemove: []string{"k1"},
			assertions: func(t *testing.T, component *Component) {
				result := component.RemoveLabels("k2").AddLabel("k4", "v4")
				assert.Equal(t, 2, result.Labels().Len())
				assert.False(t, result.Labels().Has("k1"))
				assert.False(t, result.Labels().Has("k2"))
				assert.True(t, result.Labels().ValueIs("k3", "v3"))
				assert.True(t, result.Labels().ValueIs("k4", "v4"))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			componentAfter := tt.component.RemoveLabels(tt.labelsToRemove...)
			if tt.assertions != nil {
				tt.assertions(t, componentAfter)
			}
		})
	}
}

func TestComponent_Chainability(t *testing.T) {
	t.Run("SetLabels called twice replaces all labels", func(t *testing.T) {
		c := mustNew("c1").
			SetLabels(labels.Map{"k1": "v1", "k2": "v2"}).
			SetLabels(labels.Map{"k3": "v3"})

		assert.Equal(t, 1, c.Labels().Len())
		assert.False(t, c.Labels().Has("k1"), "k1 should be replaced")
		assert.False(t, c.Labels().Has("k2"), "k2 should be replaced")
		assert.True(t, c.Labels().ValueIs("k3", "v3"))
	})

	t.Run("AddLabels called twice merges labels", func(t *testing.T) {
		c := mustNew("c1").
			AddLabels(labels.Map{"k1": "v1", "k2": "v2"}).
			AddLabels(labels.Map{"k3": "v3", "k2": "v2-updated"})

		assert.Equal(t, 3, c.Labels().Len())
		assert.True(t, c.Labels().ValueIs("k1", "v1"))
		assert.True(t, c.Labels().ValueIs("k2", "v2-updated"), "should update existing key")
		assert.True(t, c.Labels().ValueIs("k3", "v3"))
	})

	t.Run("mixed Set and Add operations", func(t *testing.T) {
		c := mustNew("c1").
			AddLabel("k1", "v1").
			AddLabels(labels.Map{"k2": "v2", "k3": "v3"}).
			SetLabels(labels.Map{"k4": "v4"}). // Wipes k1, k2, k3
			AddLabel("k5", "v5")               // Merges with k4

		assert.Equal(t, 2, c.Labels().Len())
		assert.False(t, c.Labels().Has("k1"), "wiped by SetLabels")
		assert.False(t, c.Labels().Has("k2"), "wiped by SetLabels")
		assert.False(t, c.Labels().Has("k3"), "wiped by SetLabels")
		assert.True(t, c.Labels().ValueIs("k4", "v4"))
		assert.True(t, c.Labels().ValueIs("k5", "v5"))
	})

	t.Run("WithDescription replaces previous value", func(t *testing.T) {
		c := mustNew("c1").
			WithDescription("first").
			WithDescription("second")

		assert.Equal(t, "second", c.Description())
	})

	t.Run("AddInputs called twice adds ports without duplicates", func(t *testing.T) {
		c := mustNew("c1")
		require.NoError(t, c.AddInputs("in1", "in2"))
		require.NoError(t, c.AddInputs("in3")) // in1 already exists - would error, skip in1
		// Add in1 separately to test duplicate skipping
		_ = c.AddInputs("in1") // duplicate - ignore error

		assert.Equal(t, 3, c.Inputs().Len())
		assert.NotNil(t, c.Inputs().ByName("in1"))
		assert.NotNil(t, c.Inputs().ByName("in2"))
		assert.NotNil(t, c.Inputs().ByName("in3"))
	})

	t.Run("AddOutputs called twice adds ports without duplicates", func(t *testing.T) {
		c := mustNew("c1")
		require.NoError(t, c.AddOutputs("out1", "out2"))
		require.NoError(t, c.AddOutputs("out3"))
		_ = c.AddOutputs("out1") // duplicate - ignore error

		assert.Equal(t, 3, c.Outputs().Len())
		assert.NotNil(t, c.Outputs().ByName("out1"))
		assert.NotNil(t, c.Outputs().ByName("out2"))
		assert.NotNil(t, c.Outputs().ByName("out3"))
	})

	t.Run("ClearLabels removes all labels", func(t *testing.T) {
		c := mustNew("c1").
			AddLabels(labels.Map{"k1": "v1", "k2": "v2"}).
			ClearLabels().
			AddLabel("k3", "v3")

		assert.Equal(t, 1, c.Labels().Len())
		assert.False(t, c.Labels().Has("k1"))
		assert.False(t, c.Labels().Has("k2"))
		assert.True(t, c.Labels().ValueIs("k3", "v3"))
	})

	t.Run("RemoveLabels removes specific labels", func(t *testing.T) {
		c := mustNew("c1").
			AddLabels(labels.Map{"k1": "v1", "k2": "v2", "k3": "v3"}).
			RemoveLabels("k1", "k2").
			AddLabel("k4", "v4")

		assert.Equal(t, 2, c.Labels().Len())
		assert.False(t, c.Labels().Has("k1"))
		assert.False(t, c.Labels().Has("k2"))
		assert.True(t, c.Labels().ValueIs("k3", "v3"))
		assert.True(t, c.Labels().ValueIs("k4", "v4"))
	})
}
