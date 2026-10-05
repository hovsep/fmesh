// Package meta provides [Meta], the key→value metadata store carried by
// signals, ports, components, cycles, groups, collections and the mesh.
//
// A value is a string, a bool or a Go number — see [Value]. Reads are typed at
// the call site: m.Value[float64]("temp"); ValueOrDefault and ValueIs infer
// the type from their argument. A value reads back only as the type it was
// written with. Meta mutates in place; Keys returns a sorted slice for
// determinism, and Clone and Filter are the non-mutating methods.
package meta
