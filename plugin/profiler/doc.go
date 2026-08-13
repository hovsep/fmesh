// Package profiler provides [Plugin], a mesh plugin that measures a running
// mesh, and the stat types it reports: [Stat] and [ComponentStat] for time,
// [Flow] and [PipeStat] for pipe traffic, and [CycleRecord] for the per-cycle
// timeline.
//
// [Mode] selects which of those dimensions are measured. [New] with no
// arguments measures time alone, which is the cheapest. Every number is
// mesh-attributable: the profiler measures what the mesh did, never what the
// process around it did.
package profiler
