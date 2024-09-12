package component

import (
	"errors"
	"testing"
)

func TestIsWaitingForInputError(t *testing.T) {
	type args struct {
		err error
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "no",
			args: args{
				err: errors.New("test error"),
			},
			want: false,
		},
		{
			name: "yes",
			args: args{
				err: ErrWaitingForInputKeepInputs,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsWaitingForInputError(tt.args.err); got != tt.want {
				t.Errorf("IsWaitingForInputError() = %v, want %v", got, tt.want)
			}
		})
	}
}
