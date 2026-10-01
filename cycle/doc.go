// Package cycle provides [Cycle], one synchronized execution tick of a mesh,
// and [Group], the history of cycles kept during a run (unbounded unless a
// length limit is set).
//
// A cycle collects the activation results of the components that had input in
// that tick; a component with no input has no result. Inspect them after a run
// through the mesh's runtime info. Cycle types mutate in place.
package cycle
