// Package signal provides [Signal] and [Group], the mesh's core data carriers.
//
// [Signal] and [Group] use copy-on-write: methods that appear to mutate return
// new values and never modify the receiver. [Signal.Meta] returns a defensive
// copy of the metadata store. Ranging over [Group.All] cannot change the group;
// change grouped signals with [Group.Map] or [Group.MapPayloads] so the replaced
// signals are stored in the new group.
//
// A payload is an any. [Signal.As] and [Signal.PayloadOrDefault] read one back out
// as a concrete type without the panic a bare assertion risks.
package signal
