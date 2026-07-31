package component

import "context"

// Predicate is a function that tests whether a Component matches a condition.
type Predicate func(component *Component) bool

// Mapper transforms a Component into a new Component.
type Mapper func(component *Component) *Component

// ResultPredicate is a function that tests whether an ActivationResult matches a condition.
type ResultPredicate func(result *ActivationResult) bool

// ParentMesh is an interface for a parent mesh.
type ParentMesh interface {
	Name() string
}

// ActivationFunc is the activation function of a component.
//
// The context is the one passed to FMesh.Run, narrowed by the mesh time limit
// when one is configured. Pass it to anything that can block — HTTP calls,
// database queries — and poll ctx.Err() inside long loops, so the mesh can be
// stopped from outside. Ignoring it is safe: the mesh still stops between
// cycles, it just cannot interrupt an activation that is already running.
type ActivationFunc func(ctx context.Context, this *Component) error

// Option is a functional option for configuring a component during construction.
type Option func(*Component) error
