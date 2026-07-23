// Package cycle provides [Cycle], one synchronized execution tick of a mesh,
// and [Group], the bounded history of cycles kept during a run.
//
// A cycle collects the activation results of every component that was
// considered in that tick; inspect them after a run through the mesh's
// runtime info. Cycle types mutate in place.
package cycle
