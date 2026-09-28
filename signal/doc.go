// Package signal provides [Signal] and [Group], the mesh's core data carriers.
//
// [Signal] and [Group] use copy-on-write: methods that appear to mutate return
// new values and never modify the receiver. [Group.All] returns a copy of the
// slice; [Signal.Meta] returns a defensive copy of the metadata store.
// [Group.ForEach] does not update the receiver on success; metadata changes on
// grouped signals should be done with [Group.Map] / [Group.MapPayloads] so
// replaced signals are stored in the new group (github.com/hovsep/fmesh#203).
//
// A payload is an any. [Signal.As] and [Signal.PayloadOrDefault] read one back out
// as a concrete type without the panic a bare assertion risks.
package signal
