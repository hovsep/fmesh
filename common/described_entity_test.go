package common

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewDescribedEntity(t *testing.T) {
	type args struct {
		description string
	}
	tests := []struct {
		name string
		args args
		want DescribedEntity
	}{
		{
			name: "empty description",
			args: args{
				description: "",
			},
			want: DescribedEntity{
				description: "",
			},
		},
		{
			name: "with description",
			args: args{
				description: "component1 is used to generate logs",
			},
			want: DescribedEntity{
				description: "component1 is used to generate logs",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, NewDescribedEntity(tt.args.description))
		})
	}
}

func TestDescribedEntity_Description(t *testing.T) {
	tests := []struct {
		name            string
		describedEntity DescribedEntity
		want            string
	}{
		{
			name:            "empty description",
			describedEntity: NewDescribedEntity(""),
			want:            "",
		},
		{
			name:            "with description",
			describedEntity: NewDescribedEntity("component2 is used to handle errors"),
			want:            "component2 is used to handle errors",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.describedEntity.Description())
		})
	}
}
