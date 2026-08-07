// Package plugin groups the mesh-level plugins that ship with F-Mesh. It holds
// no code of its own; each plugin is a package under it.
//
// Both are plain fmesh.Plugin implementations -- an initialization bundle that
// registers hooks and then gets out of the way -- and both reach every component
// through the OnComponentAdded hook, so they are things you add to a mesh rather
// than things components opt into.
//
//	profiler -- measures where a mesh spends its time, which pipes carry its
//	            traffic, and what the Go runtime did while it ran
//	autowire -- connects components by naming convention instead of by hand
package plugin
