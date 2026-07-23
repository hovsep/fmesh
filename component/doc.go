// Package component provides [Component], the building block of a mesh.
//
// A component owns named input and output ports and an activation function
// that runs once per cycle when inputs are ready. Components are built with
// [New] and functional options (WithInputs, WithActivationFunc, ...); after a
// run, each cycle exposes an [ActivationResult] per component describing
// whether and how it activated. Component types mutate in place.
package component
