// Package profiler provides [Plugin], a mesh plugin that measures a running
// mesh, and the stat types it reports: [Stat] and [ComponentStat] for time,
// [Flow] and [PipeStat] for pipe traffic, [CycleRecord] for the per-cycle
// timeline, and [ResourceStat] for Go runtime counters.
//
// [Mode] selects which of those dimensions are measured. [New] with no arguments
// measures time alone, which is the cheapest and the only one attributable to
// the mesh rather than to the whole process.
package profiler
