package fmesh

import (
	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/cycle"
	"reflect"
	"testing"
)

func TestNew(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name string
		args args
		want *FMesh
	}{
		{
			name: "empty name is valid",
			args: args{
				name: "",
			},
			want: &FMesh{},
		},
		{
			name: "with name",
			args: args{
				name: "fm1",
			},
			want: &FMesh{name: "fm1"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := New(tt.args.name); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFMesh_WithDescription(t *testing.T) {
	type args struct {
		description string
	}
	tests := []struct {
		name string
		fm   *FMesh
		args args
		want *FMesh
	}{
		{
			name: "empty description",
			fm:   New("fm1"),
			args: args{
				description: "",
			},
			want: &FMesh{
				name:                  "fm1",
				description:           "",
				components:            nil,
				errorHandlingStrategy: 0,
			},
		},
		{
			name: "with description",
			fm:   New("fm1"),
			args: args{
				description: "descr",
			},
			want: &FMesh{
				name:                  "fm1",
				description:           "descr",
				components:            nil,
				errorHandlingStrategy: 0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.fm.WithDescription(tt.args.description); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WithDescription() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFMesh_WithErrorHandlingStrategy(t *testing.T) {
	type args struct {
		strategy ErrorHandlingStrategy
	}
	tests := []struct {
		name string
		fm   *FMesh
		args args
		want *FMesh
	}{
		{
			name: "default strategy",
			fm:   New("fm1"),
			args: args{
				strategy: 0,
			},
			want: &FMesh{
				name:                  "fm1",
				errorHandlingStrategy: StopOnFirstError,
			},
		},
		{
			name: "custom strategy",
			fm:   New("fm1"),
			args: args{
				strategy: IgnoreAll,
			},
			want: &FMesh{
				name:                  "fm1",
				errorHandlingStrategy: IgnoreAll,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.fm.WithErrorHandlingStrategy(tt.args.strategy); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WithErrorHandlingStrategy() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFMesh_WithComponents(t *testing.T) {
	type args struct {
		components []*component.Component
	}
	tests := []struct {
		name string
		fm   *FMesh
		args args
		want *FMesh
	}{
		{
			name: "no components",
			fm:   New("fm1"),
			args: args{
				components: nil,
			},
			want: &FMesh{
				name:                  "fm1",
				description:           "",
				components:            component.Components{},
				errorHandlingStrategy: 0,
			},
		},
		{
			name: "with single component",
			fm:   New("fm1"),
			args: args{
				components: []*component.Component{
					component.NewComponent("c1"),
				},
			},
			want: &FMesh{
				name: "fm1",
				components: component.Components{
					"c1": component.NewComponent("c1"),
				},
			},
		},
		{
			name: "with multiple components",
			fm:   New("fm1"),
			args: args{
				components: []*component.Component{
					component.NewComponent("c1"),
					component.NewComponent("c2"),
				},
			},
			want: &FMesh{
				name: "fm1",
				components: component.Components{
					"c1": component.NewComponent("c1"),
					"c2": component.NewComponent("c2"),
				},
			},
		},
		{
			name: "components with duplicating name are collapsed",
			fm:   New("fm1"),
			args: args{
				components: []*component.Component{
					component.NewComponent("c1").WithDescription("descr1"),
					component.NewComponent("c2").WithDescription("descr2"),
					component.NewComponent("c2").WithDescription("descr3"), //This will overwrite the previous one
					component.NewComponent("c4").WithDescription("descr4"),
				},
			},
			want: &FMesh{
				name: "fm1",
				components: component.Components{
					"c1": component.NewComponent("c1").WithDescription("descr1"),
					"c2": component.NewComponent("c2").WithDescription("descr3"),
					"c4": component.NewComponent("c4").WithDescription("descr4"),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.fm.WithComponents(tt.args.components...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WithComponents() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFMesh_Run(t *testing.T) {
	type fields struct {
		name                  string
		description           string
		components            component.Components
		errorHandlingStrategy ErrorHandlingStrategy
	}
	tests := []struct {
		name    string
		fields  fields
		want    []*cycle.Result
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fm := &FMesh{
				name:                  tt.fields.name,
				description:           tt.fields.description,
				components:            tt.fields.components,
				errorHandlingStrategy: tt.fields.errorHandlingStrategy,
			}
			got, err := fm.Run()
			if (err != nil) != tt.wantErr {
				t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Run() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFMesh_activateComponents(t *testing.T) {
	tests := []struct {
		name string
		fm   *FMesh
		want *cycle.Result
	}{
		//@TODO
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.fm.activateComponents(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("activateComponents() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFMesh_drainComponents(t *testing.T) {
	type fields struct {
		name                  string
		description           string
		components            component.Components
		errorHandlingStrategy ErrorHandlingStrategy
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fm := &FMesh{
				name:                  tt.fields.name,
				description:           tt.fields.description,
				components:            tt.fields.components,
				errorHandlingStrategy: tt.fields.errorHandlingStrategy,
			}
			fm.drainComponents()
		})
	}
}
