// Package fmesh orchestrates a mesh of components exchanging signals through
// ports, following the Flow-Based Programming model.
//
// Build a mesh with [New], add components, and start it with [FMesh.Run].
// Execution proceeds in discrete synchronized cycles: every cycle all ready
// components activate concurrently, then outputs are flushed through pipes to
// downstream inputs. The mesh stops naturally when no component activates in a
// cycle, or on the cycle limit, time limit, error-handling strategy or livelock
// detection configured via the With* options, or when its context is canceled.
package fmesh
