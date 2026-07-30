package port

import (
	"context"
	"testing"

	"github.com/hovsep/fmesh/signal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMultiPipe(t *testing.T) {
	t.Run("wires every edge", func(t *testing.T) {
		out1, out2 := mustOutput("o1"), mustOutput("o2")
		in1, in2 := mustInput("i1"), mustInput("i2")

		require.NoError(t, MultiPipe(
			Pipe{From: out1, To: in1},
			Pipe{From: out2, To: in2},
		))

		require.NoError(t, out1.PutSignals(signal.New(1)))
		require.NoError(t, out1.Flush(context.Background()))
		assert.True(t, in1.HasSignals())
	})

	t.Run("names the edge that failed", func(t *testing.T) {
		// A bare error from PipeTo gives no clue which of a dozen lines
		// produced it.
		err := MultiPipe(Pipe{From: mustInput("wrong_way"), To: mustInput("i1")})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "wrong_way -> i1")
	})

	t.Run("a missing port is reported, not panicked on", func(t *testing.T) {
		err := MultiPipe(Pipe{From: nil, To: mustInput("i1")})

		require.ErrorContains(t, err, "<missing port> -> i1")
	})
}

func TestMultiForward(t *testing.T) {
	t.Run("forwards every pair", func(t *testing.T) {
		src1, src2 := mustInput("s1"), mustInput("s2")
		dst1, dst2 := mustInput("d1"), mustInput("d2")
		require.NoError(t, src1.PutSignals(signal.New(1)))
		require.NoError(t, src2.PutSignals(signal.New(2)))

		require.NoError(t, MultiForward(context.Background(), Pair{From: src1, To: dst1}, Pair{From: src2, To: dst2}))

		assert.True(t, dst1.HasSignals())
		assert.True(t, dst2.HasSignals())
	})

	t.Run("a missing port is reported, not panicked on", func(t *testing.T) {
		err := MultiForward(context.Background(), Pair{From: nil, To: mustInput("d1")})

		require.ErrorContains(t, err, "<missing port> -> d1")
	})
}
