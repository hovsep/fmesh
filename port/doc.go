// Package port provides [Port], the data endpoint of a component, along with
// [Group] (ordered port list) and [Collection] (name-keyed port map).
//
// Ports are connected output→input with [Port.PipeTo]; [Port.Flush] fans a
// port's signal buffer out through its pipes and clears the source. Unlike
// the copy-on-write signal package, port types mutate in place: Set*, Add*
// and Remove* methods modify the receiver.
package port
