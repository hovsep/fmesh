package port

import (
	"github.com/hovsep/fmesh/common"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewGroup(t *testing.T) {
	type args struct {
		names []string
	}
	tests := []struct {
		name string
		args args
		want *Group
	}{
		{
			name: "empty group",
			args: args{
				names: nil,
			},
			want: &Group{
				Chainable: common.NewChainable(),
				ports:     Ports{},
			},
		},
		{
			name: "non-empty group",
			args: args{
				names: []string{"p1", "p2"},
			},
			want: &Group{
				Chainable: common.NewChainable(),
				ports: Ports{New("p1"),
					New("p2")},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, NewGroup(tt.args.names...))
		})
	}
}

func TestNewIndexedGroup(t *testing.T) {
	type args struct {
		prefix     string
		startIndex int
		endIndex   int
	}
	tests := []struct {
		name string
		args args
		want *Group
	}{
		{
			name: "empty prefix is valid",
			args: args{
				prefix:     "",
				startIndex: 0,
				endIndex:   3,
			},
			want: NewGroup("0", "1", "2", "3"),
		},
		{
			name: "with prefix",
			args: args{
				prefix:     "in_",
				startIndex: 4,
				endIndex:   5,
			},
			want: NewGroup("in_4", "in_5"),
		},
		{
			name: "with invalid start index",
			args: args{
				prefix:     "",
				startIndex: 999,
				endIndex:   5,
			},
			want: NewGroup().WithChainError(ErrInvalidRangeForIndexedGroup),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, NewIndexedGroup(tt.args.prefix, tt.args.startIndex, tt.args.endIndex))
		})
	}
}

func TestGroup_With(t *testing.T) {
	type args struct {
		ports Ports
	}
	tests := []struct {
		name       string
		group      *Group
		args       args
		assertions func(t *testing.T, group *Group)
	}{
		{
			name:  "adding nothing to empty group",
			group: NewGroup(),
			args: args{
				ports: nil,
			},
			assertions: func(t *testing.T, group *Group) {
				assert.Zero(t, group.Len())
			},
		},
		{
			name:  "adding to empty group",
			group: NewGroup(),
			args: args{
				ports: NewGroup("p1", "p2", "p3").PortsOrNil(),
			},
			assertions: func(t *testing.T, group *Group) {
				assert.Equal(t, group.Len(), 3)
			},
		},
		{
			name:  "adding to non-empty group",
			group: NewIndexedGroup("p", 1, 3),
			args: args{
				ports: NewGroup("p4", "p5", "p6").PortsOrNil(),
			},
			assertions: func(t *testing.T, group *Group) {
				assert.Equal(t, group.Len(), 6)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			groupAfter := tt.group.With(tt.args.ports...)
			if tt.assertions != nil {
				tt.assertions(t, groupAfter)
			}
		})
	}
}
