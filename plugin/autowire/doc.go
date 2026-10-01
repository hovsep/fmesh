// Package autowire provides [Plugin], a mesh plugin that pipes components
// together by naming convention instead of by hand.
//
// [Broadcast], [BroadcastAs] and [Prefixed] are the ready-made conventions;
// set [Plugin.InputNameFor] to a rule of your own. Wiring happens as each component
// arrives and goes in both directions, so the order of AddComponents does not
// matter.
package autowire
