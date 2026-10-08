package component

import "context"

// Predicate is a function that tests whether a Component matches a condition.
type Predicate func(component *Component) bool

// ResultPredicate is a function that tests whether an ActivationResult matches a condition.
type ResultPredicate func(result *ActivationResult) bool

// ParentMesh is an interface for a parent mesh.
type ParentMesh interface {
	Name() string
}

// ActivationFunc is the activation function of a component.
//
// ctx is the one passed to FMesh.Run, narrowed by the mesh time limit. Pass it
// to anything that blocks: the mesh cannot interrupt a running activation.
type ActivationFunc func(ctx context.Context, this *Component) error

// Option is a functional option for configuring a component during construction.
type Option func(*Component) error
