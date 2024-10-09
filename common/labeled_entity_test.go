package common

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewLabeledEntity(t *testing.T) {
	type args struct {
		labels LabelsCollection
	}
	tests := []struct {
		name string
		args args
		want LabeledEntity
	}{
		{
			name: "empty labels",
			args: args{
				labels: nil,
			},
			want: LabeledEntity{
				labels: nil,
			},
		},
		{
			name: "with labels",
			args: args{
				labels: LabelsCollection{
					"label1": "value1",
				},
			},
			want: LabeledEntity{
				labels: LabelsCollection{
					"label1": "value1",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, NewLabeledEntity(tt.args.labels))
		})
	}
}

func TestLabeledEntity_Labels(t *testing.T) {
	tests := []struct {
		name          string
		labeledEntity LabeledEntity
		want          LabelsCollection
	}{
		{
			name:          "no labels",
			labeledEntity: NewLabeledEntity(nil),
			want:          nil,
		},
		{
			name: "with labels",
			labeledEntity: NewLabeledEntity(LabelsCollection{
				"l1": "v1",
				"l2": "v2",
			}),
			want: LabelsCollection{
				"l1": "v1",
				"l2": "v2",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.labeledEntity.Labels())
		})
	}
}

func TestLabeledEntity_SetLabels(t *testing.T) {
	type args struct {
		labels LabelsCollection
	}
	tests := []struct {
		name          string
		labeledEntity LabeledEntity
		args          args
		assertions    func(t *testing.T, labeledEntity LabeledEntity)
	}{
		{
			name:          "setting to empty labels collection",
			labeledEntity: NewLabeledEntity(nil),
			args: args{
				labels: LabelsCollection{
					"l1": "v1",
					"l2": "v2",
					"l3": "v3",
				},
			},
			assertions: func(t *testing.T, labeledEntity LabeledEntity) {
				assert.Equal(t, LabelsCollection{
					"l1": "v1",
					"l2": "v2",
					"l3": "v3",
				}, labeledEntity.Labels())
			},
		},
		{
			name: "setting to non-empty labels collection",
			labeledEntity: NewLabeledEntity(LabelsCollection{
				"l1":  "v1",
				"l2":  "v2",
				"l3":  "v3",
				"l99": "val1",
			}),
			args: args{
				labels: LabelsCollection{
					"l4":  "v4",
					"l5":  "v5",
					"l6":  "v6",
					"l99": "val2",
				},
			},
			assertions: func(t *testing.T, labeledEntity LabeledEntity) {
				assert.Equal(t, LabelsCollection{
					"l4":  "v4",
					"l5":  "v5",
					"l6":  "v6",
					"l99": "val2",
				}, labeledEntity.Labels())
			},
		},
		{
			name: "setting nil",
			labeledEntity: NewLabeledEntity(LabelsCollection{
				"l1":  "v1",
				"l2":  "v2",
				"l3":  "v3",
				"l99": "val1",
			}),
			args: args{
				labels: nil,
			},
			assertions: func(t *testing.T, labeledEntity LabeledEntity) {
				assert.Nil(t, labeledEntity.Labels())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.labeledEntity.SetLabels(tt.args.labels)
			if tt.assertions != nil {
				tt.assertions(t, tt.labeledEntity)
			}
		})
	}
}

func TestLabeledEntity_AddLabel(t *testing.T) {
	type args struct {
		label string
		value string
	}
	tests := []struct {
		name          string
		labeledEntity LabeledEntity
		args          args
		assertions    func(t *testing.T, labeledEntity LabeledEntity)
	}{
		{
			name:          "adding to empty labels collection",
			labeledEntity: NewLabeledEntity(nil),
			args: args{
				label: "l1",
				value: "v1",
			},
			assertions: func(t *testing.T, labeledEntity LabeledEntity) {
				assert.Equal(t, LabelsCollection{
					"l1": "v1",
				}, labeledEntity.Labels())
			},
		},
		{
			name: "adding to non-empty labels collection",
			labeledEntity: NewLabeledEntity(LabelsCollection{
				"l1": "v1",
			}),
			args: args{
				label: "l2",
				value: "v2",
			},
			assertions: func(t *testing.T, labeledEntity LabeledEntity) {
				assert.Equal(t, LabelsCollection{
					"l1": "v1",
					"l2": "v2",
				}, labeledEntity.Labels())
			},
		},
		{
			name: "overwriting a label",
			labeledEntity: NewLabeledEntity(LabelsCollection{
				"l1": "v1",
				"l2": "v2",
			}),
			args: args{
				label: "l2",
				value: "v3",
			},
			assertions: func(t *testing.T, labeledEntity LabeledEntity) {
				assert.Equal(t, LabelsCollection{
					"l1": "v1",
					"l2": "v3",
				}, labeledEntity.Labels())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.labeledEntity.AddLabel(tt.args.label, tt.args.value)
			if tt.assertions != nil {
				tt.assertions(t, tt.labeledEntity)
			}
		})
	}
}

func TestLabeledEntity_AddLabels(t *testing.T) {
	type args struct {
		labels LabelsCollection
	}
	tests := []struct {
		name          string
		labeledEntity LabeledEntity
		args          args
		assertions    func(t *testing.T, labeledEntity LabeledEntity)
	}{
		{
			name: "adding to non-empty labels collection",
			labeledEntity: NewLabeledEntity(LabelsCollection{
				"l1": "v1",
				"l2": "v2",
				"l3": "v3",
			}),
			args: args{
				labels: LabelsCollection{
					"l3": "v100",
					"l4": "v4",
					"l5": "v5",
				},
			},
			assertions: func(t *testing.T, labeledEntity LabeledEntity) {
				assert.Equal(t, LabelsCollection{
					"l1": "v1",
					"l2": "v2",
					"l3": "v100",
					"l4": "v4",
					"l5": "v5",
				}, labeledEntity.Labels())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.labeledEntity.AddLabels(tt.args.labels)
			if tt.assertions != nil {
				tt.assertions(t, tt.labeledEntity)
			}
		})
	}
}

func TestLabeledEntity_DeleteLabel(t *testing.T) {
	type args struct {
		label string
	}
	tests := []struct {
		name          string
		labeledEntity LabeledEntity
		args          args
		assertions    func(t *testing.T, labeledEntity LabeledEntity)
	}{
		{
			name: "label found and deleted",
			labeledEntity: NewLabeledEntity(LabelsCollection{
				"l1": "v1",
				"l2": "v2",
			}),
			args: args{
				label: "l1",
			},
			assertions: func(t *testing.T, labeledEntity LabeledEntity) {
				assert.Equal(t, LabelsCollection{
					"l2": "v2",
				}, labeledEntity.Labels())
			},
		},
		{
			name: "label not found, no-op",
			labeledEntity: NewLabeledEntity(LabelsCollection{
				"l1": "v1",
				"l2": "v2",
			}),
			args: args{
				label: "l3",
			},
			assertions: func(t *testing.T, labeledEntity LabeledEntity) {
				assert.Equal(t, LabelsCollection{
					"l1": "v1",
					"l2": "v2",
				}, labeledEntity.Labels())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.labeledEntity.DeleteLabel(tt.args.label)
			if tt.assertions != nil {
				tt.assertions(t, tt.labeledEntity)
			}
		})
	}
}

func TestLabeledEntity_HasAllLabels(t *testing.T) {
	type args struct {
		label []string
	}
	tests := []struct {
		name          string
		labeledEntity LabeledEntity
		args          args
		want          bool
	}{
		{
			name:          "empty collection",
			labeledEntity: NewLabeledEntity(nil),
			args: args{
				label: []string{"l1"},
			},
			want: false,
		},
		{
			name: "has all labels",
			labeledEntity: NewLabeledEntity(LabelsCollection{
				"l1": "v1",
				"l2": "v2",
				"l3": "v3",
			}),
			args: args{
				label: []string{"l1", "l2"},
			},
			want: true,
		},
		{
			name: "does not have all labels",
			labeledEntity: NewLabeledEntity(LabelsCollection{
				"l1": "v1",
				"l2": "v2",
				"l3": "v3",
			}),
			args: args{
				label: []string{"l1", "l2", "l4"},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.labeledEntity.HasAllLabels(tt.args.label...))
		})
	}
}

func TestLabeledEntity_HasAnyLabel(t *testing.T) {
	type args struct {
		label []string
	}
	tests := []struct {
		name          string
		labeledEntity LabeledEntity
		args          args
		want          bool
	}{
		{
			name:          "empty collection",
			labeledEntity: NewLabeledEntity(nil),
			args: args{
				label: []string{"l1"},
			},
			want: false,
		},
		{
			name: "has some labels",
			labeledEntity: NewLabeledEntity(LabelsCollection{
				"l1": "v1",
				"l2": "v2",
				"l3": "v3",
			}),
			args: args{
				label: []string{"l1", "l10"},
			},
			want: true,
		},
		{
			name: "does not have any of labels",
			labeledEntity: NewLabeledEntity(LabelsCollection{
				"l1": "v1",
				"l2": "v2",
				"l3": "v3",
			}),
			args: args{
				label: []string{"l10", "l20", "l4"},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.labeledEntity.HasAnyLabel(tt.args.label...))
		})
	}
}

func TestLabeledEntity_Label(t *testing.T) {
	type args struct {
		label string
	}
	tests := []struct {
		name          string
		labeledEntity LabeledEntity
		args          args
		want          string
		wantErr       bool
	}{
		{
			name: "no labels",
			labeledEntity: LabeledEntity{
				labels: nil,
			},
			args: args{
				label: "l1",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "label found",
			labeledEntity: NewLabeledEntity(LabelsCollection{
				"l1": "v1",
				"l2": "v2",
			}),
			args: args{
				label: "l2",
			},
			want:    "v2",
			wantErr: false,
		},
		{
			name: "label not found",
			labeledEntity: NewLabeledEntity(LabelsCollection{
				"l1": "v1",
				"l2": "v2",
			}),
			args: args{
				label: "l3",
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.labeledEntity.Label(tt.args.label)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.want, got)
		})
	}
}
