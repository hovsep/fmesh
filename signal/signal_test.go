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
		name string
		args args
		want *Signal
	}{
		{
			name: "nil payload is valid",
			args: args{
				payload: nil,
			},
			want: &Signal{
				payload:      []any{nil},
				chainableErr: nil,
				labels:       labels.NewCollection(nil),
			},
		},
		{
			name: "with payload",
			args: args{
				payload: []any{123, "hello", []int{1, 2, 3}, map[string]int{"key": 42}, []byte{}, nil},
			},
			want: &Signal{
				payload:      []any{[]any{123, "hello", []int{1, 2, 3}, map[string]int{"key": 42}, []byte{}, nil}},
				chainableErr: nil,
				labels:       labels.NewCollection(nil),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, New(tt.args.payload))
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
			signal: New(1),
			mapperFunc: func(payload any) any {
				return payload.(int) * 2
			},
			want: New(2),
		},
		{
			name:   "with chain error",
			signal: New(1).WithChainableErr(errors.New("some error in chain")),
			mapperFunc: func(payload any) any {
				return payload.(int) * 2
			},
			want: New(1).WithChainableErr(errors.New("some error in chain")),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.signal.MapPayload(tt.mapperFunc))
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
