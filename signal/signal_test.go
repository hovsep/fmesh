package signal

import (
	"errors"
	"testing"

	"github.com/hovsep/fmesh/labels"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	type args struct {
		payload any
	}
	tests := []struct {
		name            string
		args            args
		expectedPayload any
	}{
		{
			name: "nil payload is valid",
			args: args{
				payload: nil,
			},
			expectedPayload: nil,
		},
		{
			name: "with payload",
			args: args{
				payload: []any{123, "hello", []int{1, 2, 3}, map[string]int{"key": 42}, []byte{}, nil},
			},
			expectedPayload: []any{123, "hello", []int{1, 2, 3}, map[string]int{"key": 42}, []byte{}, nil},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sig := New(tt.args.payload)
			require.NotNil(t, sig)
			require.NotNil(t, sig.Labels())
			assert.False(t, sig.HasChainableErr())
			
			payload, err := sig.Payload()
			require.NoError(t, err)
			assert.Equal(t, tt.expectedPayload, payload)
		})
	}
}

func TestSignal_Payload(t *testing.T) {
	tests := []struct {
		name            string
		signal          *Signal
		want            any
		wantErrorString string
	}{
		{
			name:   "nil payload is valid",
			signal: New(nil),
			want:   nil,
		},
		{
			name:   "with payload",
			signal: New(123),
			want:   123,
		},
		{
			name:            "with error in chain",
			signal:          New(123).WithChainableErr(errors.New("some error in chain")),
			want:            nil,
			wantErrorString: "some error in chain",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.signal.Payload()
			if tt.wantErrorString != "" {
				require.Error(t, err)
				require.EqualError(t, err, tt.wantErrorString)
			} else {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestSignal_PayloadOrNil(t *testing.T) {
	tests := []struct {
		name   string
		signal *Signal
		want   any
	}{
		{
			name:   "payload returned",
			signal: New(123),
			want:   123,
		},
		{
			name:   "nil returned",
			signal: New(123).WithChainableErr(errors.New("some error in chain")),
			want:   nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.signal.PayloadOrNil())
		})
	}
}

func TestSignal_Map(t *testing.T) {
	tests := []struct {
		name       string
		signal     *Signal
		mapperFunc Mapper
		want       *Signal
	}{
		{
			name:   "happy path",
			signal: New(1),
			mapperFunc: func(signal *Signal) *Signal {
				return signal.SetLabels(labels.Map{
					"l1": "v1",
				})
			},
			want: New(1).SetLabels(labels.Map{
				"l1": "v1",
			}),
		},
		{
			name:   "with chain error",
			signal: New(1).WithChainableErr(errors.New("some error in chain")),
			mapperFunc: func(signal *Signal) *Signal {
				return signal.SetLabels(labels.Map{
					"l1": "v1",
				})
			},
			want: New(1).WithChainableErr(errors.New("some error in chain")),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.mapperFunc(tt.signal))
		})
	}
}

func TestSignal_MapPayload(t *testing.T) {
	tests := []struct {
		name       string
		signal     *Signal
		mapperFunc PayloadMapper
		want       *Signal
	}{
		{
			name:   "happy path",
			signal: New(1).AddLabel("foo", "bar"),
			mapperFunc: func(payload any) any {
				return payload.(int) * 2
			},
			want: New(2).AddLabel("foo", "bar"),
		},
		{
			name:   "with chain error",
			signal: New(1).WithChainableErr(errors.New("some error in chain")),
			mapperFunc: func(payload any) any {
				return payload.(int) * 2
			},
			want: New(1).WithChainableErr(errors.New("some error in chain")),
		},
		{
			name:   "payload nil",
			signal: New(nil).AddLabel("x", "y"),
			mapperFunc: func(payload any) any {
				if payload == nil {
					return "default"
				}
				return payload
			},
			want: New("default").AddLabel("x", "y"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.signal.MapPayload(tt.mapperFunc)
			assert.Equal(t, tt.want.PayloadOrNil(), got.PayloadOrNil())
			assert.Equal(t, tt.want.Labels(), got.Labels())
			assert.Equal(t, tt.want.HasChainableErr(), got.HasChainableErr())
		})
	}
}

func TestSignal_SetLabels(t *testing.T) {
	tests := []struct {
		name       string
		signal     *Signal
		labels     labels.Map
		assertions func(t *testing.T, signal *Signal)
	}{
		{
			name:   "set labels on new signal",
			signal: New(123),
			labels: labels.Map{
				"l1": "v1",
				"l2": "v2",
			},
			assertions: func(t *testing.T, signal *Signal) {
				assert.Equal(t, 2, signal.Labels().Len())
				assert.True(t, signal.labels.HasAll("l1", "l2"))
			},
		},
		{
			name:   "set labels replaces existing labels",
			signal: New(123).AddLabels(labels.Map{"old": "value"}),
			labels: labels.Map{
				"l1": "v1",
				"l2": "v2",
			},
			assertions: func(t *testing.T, signal *Signal) {
				assert.Equal(t, 2, signal.Labels().Len())
				assert.True(t, signal.labels.HasAll("l1", "l2"))
				assert.False(t, signal.labels.Has("old"))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			signalAfter := tt.signal.SetLabels(tt.labels)
			if tt.assertions != nil {
				tt.assertions(t, signalAfter)
			}
		})
	}
}

func TestSignal_AddLabels(t *testing.T) {
	tests := []struct {
		name       string
		signal     *Signal
		labels     labels.Map
		assertions func(t *testing.T, signal *Signal)
	}{
		{
			name:   "add labels to new signal",
			signal: New(123),
			labels: labels.Map{
				"l1": "v1",
				"l2": "v2",
			},
			assertions: func(t *testing.T, signal *Signal) {
				assert.Equal(t, 2, signal.Labels().Len())
				assert.True(t, signal.labels.HasAll("l1", "l2"))
			},
		},
		{
			name:   "add labels merges with existing",
			signal: New(123).AddLabels(labels.Map{"existing": "label"}),
			labels: labels.Map{
				"l1": "v1",
				"l2": "v2",
			},
			assertions: func(t *testing.T, signal *Signal) {
				assert.Equal(t, 3, signal.Labels().Len())
				assert.True(t, signal.labels.HasAll("existing", "l1", "l2"))
			},
		},
		{
			name:   "add labels updates existing key",
			signal: New(123).AddLabels(labels.Map{"l1": "old"}),
			labels: labels.Map{
				"l1": "new",
			},
			assertions: func(t *testing.T, signal *Signal) {
				assert.Equal(t, 1, signal.Labels().Len())
				assert.True(t, signal.labels.ValueIs("l1", "new"))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			signalAfter := tt.signal.AddLabels(tt.labels)
			if tt.assertions != nil {
				tt.assertions(t, signalAfter)
			}
		})
	}
}

func TestSignal_AddLabel(t *testing.T) {
	tests := []struct {
		name       string
		signal     *Signal
		labelName  string
		labelValue string
		assertions func(t *testing.T, signal *Signal)
	}{
		{
			name:       "add single label to new signal",
			signal:     New(123),
			labelName:  "priority",
			labelValue: "high",
			assertions: func(t *testing.T, signal *Signal) {
				assert.Equal(t, 1, signal.Labels().Len())
				assert.True(t, signal.labels.ValueIs("priority", "high"))
			},
		},
		{
			name:       "add label merges with existing",
			signal:     New(123).AddLabel("existing", "label"),
			labelName:  "priority",
			labelValue: "high",
			assertions: func(t *testing.T, signal *Signal) {
				assert.Equal(t, 2, signal.Labels().Len())
				assert.True(t, signal.labels.HasAll("existing", "priority"))
			},
		},
		{
			name:       "add label updates existing key",
			signal:     New(123).AddLabel("priority", "low"),
			labelName:  "priority",
			labelValue: "high",
			assertions: func(t *testing.T, signal *Signal) {
				assert.Equal(t, 1, signal.Labels().Len())
				assert.True(t, signal.labels.ValueIs("priority", "high"))
			},
		},
		{
			name:       "chainable",
			signal:     New(123),
			labelName:  "l1",
			labelValue: "v1",
			assertions: func(t *testing.T, signal *Signal) {
				result := signal.AddLabel("l2", "v2").AddLabel("l3", "v3")
				assert.Equal(t, 3, result.Labels().Len())
				assert.True(t, result.labels.HasAll("l1", "l2", "l3"))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			signalAfter := tt.signal.AddLabel(tt.labelName, tt.labelValue)
			if tt.assertions != nil {
				tt.assertions(t, signalAfter)
			}
		})
	}
}

func TestSignal_ClearLabels(t *testing.T) {
	tests := []struct {
		name       string
		signal     *Signal
		assertions func(t *testing.T, signal *Signal)
	}{
		{
			name:   "clear labels from signal with labels",
			signal: New(123).AddLabels(labels.Map{"k1": "v1", "k2": "v2"}),
			assertions: func(t *testing.T, signal *Signal) {
				assert.Equal(t, 0, signal.Labels().Len())
				assert.False(t, signal.Labels().Has("k1"))
				assert.False(t, signal.Labels().Has("k2"))
			},
		},
		{
			name:   "clear labels from signal without labels",
			signal: New(123),
			assertions: func(t *testing.T, signal *Signal) {
				assert.Equal(t, 0, signal.Labels().Len())
			},
		},
		{
			name:   "chainable",
			signal: New(123).AddLabels(labels.Map{"k1": "v1"}),
			assertions: func(t *testing.T, signal *Signal) {
				result := signal.ClearLabels().AddLabel("k2", "v2")
				assert.Equal(t, 1, result.Labels().Len())
				assert.False(t, result.Labels().Has("k1"))
				assert.True(t, result.Labels().ValueIs("k2", "v2"))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			signalAfter := tt.signal.ClearLabels()
			if tt.assertions != nil {
				tt.assertions(t, signalAfter)
			}
		})
	}
}

func TestSignal_WithoutLabels(t *testing.T) {
	tests := []struct {
		name           string
		signal         *Signal
		labelsToRemove []string
		assertions     func(t *testing.T, signal *Signal)
	}{
		{
			name:           "remove single label",
			signal:         New(123).AddLabels(labels.Map{"k1": "v1", "k2": "v2", "k3": "v3"}),
			labelsToRemove: []string{"k1"},
			assertions: func(t *testing.T, signal *Signal) {
				assert.Equal(t, 2, signal.Labels().Len())
				assert.False(t, signal.Labels().Has("k1"))
				assert.True(t, signal.Labels().Has("k2"))
				assert.True(t, signal.Labels().Has("k3"))
			},
		},
		{
			name:           "remove multiple labels",
			signal:         New(123).AddLabels(labels.Map{"k1": "v1", "k2": "v2", "k3": "v3"}),
			labelsToRemove: []string{"k1", "k2"},
			assertions: func(t *testing.T, signal *Signal) {
				assert.Equal(t, 1, signal.Labels().Len())
				assert.False(t, signal.Labels().Has("k1"))
				assert.False(t, signal.Labels().Has("k2"))
				assert.True(t, signal.Labels().ValueIs("k3", "v3"))
			},
		},
		{
			name:           "remove non-existent label",
			signal:         New(123).AddLabels(labels.Map{"k1": "v1"}),
			labelsToRemove: []string{"k2"},
			assertions: func(t *testing.T, signal *Signal) {
				assert.Equal(t, 1, signal.Labels().Len())
				assert.True(t, signal.Labels().ValueIs("k1", "v1"))
			},
		},
		{
			name:           "chainable",
			signal:         New(123).AddLabels(labels.Map{"k1": "v1", "k2": "v2", "k3": "v3"}),
			labelsToRemove: []string{"k1"},
			assertions: func(t *testing.T, signal *Signal) {
				result := signal.WithoutLabels("k2").AddLabel("k4", "v4")
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
			signalAfter := tt.signal.WithoutLabels(tt.labelsToRemove...)
			if tt.assertions != nil {
				tt.assertions(t, signalAfter)
			}
		})
	}
}

func TestSignal_Chainability(t *testing.T) {
	t.Run("SetLabels called twice replaces all labels", func(t *testing.T) {
		s := New(123).
			SetLabels(labels.Map{"k1": "v1", "k2": "v2"}).
			SetLabels(labels.Map{"k3": "v3"})

		assert.Equal(t, 1, s.Labels().Len())
		assert.False(t, s.Labels().Has("k1"), "k1 should be replaced")
		assert.False(t, s.Labels().Has("k2"), "k2 should be replaced")
		assert.True(t, s.Labels().ValueIs("k3", "v3"))
	})

	t.Run("AddLabels called twice merges labels", func(t *testing.T) {
		s := New(123).
			AddLabels(labels.Map{"k1": "v1", "k2": "v2"}).
			AddLabels(labels.Map{"k3": "v3", "k2": "v2-updated"})

		assert.Equal(t, 3, s.Labels().Len())
		assert.True(t, s.Labels().ValueIs("k1", "v1"))
		assert.True(t, s.Labels().ValueIs("k2", "v2-updated"), "should update existing key")
		assert.True(t, s.Labels().ValueIs("k3", "v3"))
	})

	t.Run("mixed Set and Add operations", func(t *testing.T) {
		s := New(123).
			AddLabel("k1", "v1").
			AddLabels(labels.Map{"k2": "v2", "k3": "v3"}).
			SetLabels(labels.Map{"k4": "v4"}). // Wipes k1, k2, k3
			AddLabel("k5", "v5")               // Merges with k4

		assert.Equal(t, 2, s.Labels().Len())
		assert.False(t, s.Labels().Has("k1"), "wiped by SetLabels")
		assert.False(t, s.Labels().Has("k2"), "wiped by SetLabels")
		assert.False(t, s.Labels().Has("k3"), "wiped by SetLabels")
		assert.True(t, s.Labels().ValueIs("k4", "v4"))
		assert.True(t, s.Labels().ValueIs("k5", "v5"))
	})

	t.Run("ClearLabels removes all labels", func(t *testing.T) {
		s := New(123).
			AddLabels(labels.Map{"k1": "v1", "k2": "v2"}).
			ClearLabels().
			AddLabel("k3", "v3")

		assert.Equal(t, 1, s.Labels().Len())
		assert.False(t, s.Labels().Has("k1"))
		assert.False(t, s.Labels().Has("k2"))
		assert.True(t, s.Labels().ValueIs("k3", "v3"))
	})

	t.Run("WithoutLabels removes specific labels", func(t *testing.T) {
		s := New(123).
			AddLabels(labels.Map{"k1": "v1", "k2": "v2", "k3": "v3"}).
			WithoutLabels("k1", "k2").
			AddLabel("k4", "v4")

		assert.Equal(t, 2, s.Labels().Len())
		assert.False(t, s.Labels().Has("k1"))
		assert.False(t, s.Labels().Has("k2"))
		assert.True(t, s.Labels().ValueIs("k3", "v3"))
		assert.True(t, s.Labels().ValueIs("k4", "v4"))
	})
}
