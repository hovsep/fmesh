// Package signal provides [Signal] and [Group], the mesh's core data carriers.
//
// [Signal] and [Group] use copy-on-write: methods that appear to mutate return
// new values and never modify the receiver. [Group.All] returns a copy of the
// slice; [Signal.Labels] returns a defensive copy of the label collection.
// [Group.ForEach] does not update the receiver on success; label changes on
// grouped signals should be done with [Group.Map] / [Group.MapPayloads] so
// replaced signals are stored in the new group (github.com/hovsep/fmesh#203).
package signal
