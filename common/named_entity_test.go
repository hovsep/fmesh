package common

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewNamedEntity(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name string
		args args
		want NamedEntity
	}{
		{
			name: "empty name is valid",
			args: args{
				name: "",
			},
			want: NamedEntity{
				name: "",
			},
		},
		{
			name: "with name",
			args: args{
				name: "component1",
			},
			want: NamedEntity{
				name: "component1",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, NewNamedEntity(tt.args.name))
		})
	}
}

func TestNamedEntity_Name(t *testing.T) {
	tests := []struct {
		name        string
		namedEntity NamedEntity
		want        string
	}{
		{
			name:        "empty name",
			namedEntity: NewNamedEntity(""),
			want:        "",
		},
		{
			name:        "with name",
			namedEntity: NewNamedEntity("port2"),
			want:        "port2",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.namedEntity.Name())
		})
	}
}
