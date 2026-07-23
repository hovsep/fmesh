// Package meta provides [Labels] (string key/value) and [Scalars]
// (string→float64) metadata carried by signals, ports, components, cycles
// and the mesh itself.
//
// Both types mutate in place; Keys and Values return sorted slices for
// determinism, and Merge is the one non-mutating method on each type.
package meta
