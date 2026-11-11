package timers

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEpochTimers_DeltaTick(t *testing.T) {
	var (
		newEpochTicks int
		deltaBlocks   []uint32
	)
	et := NewTimers(EpochTicks{
		NewEpochTicks: []Tick{func() { newEpochTicks++ }},
		DeltaTicks: []SubEpochTick{{
			Tick:     func(blockIndex uint32) { deltaBlocks = append(deltaBlocks, blockIndex) },
			EpochMul: 1,
			EpochDiv: 2,
		}},
	})

	et.Reset(1000, 100)

	et.UpdateTime(1010, 10)
	et.UpdateTime(1049, 11)
	require.Empty(t, deltaBlocks)

	et.UpdateTime(1050, 12)
	et.UpdateTime(1060, 13)
	require.Equal(t, []uint32{12}, deltaBlocks)
	require.Zero(t, newEpochTicks)

	et.UpdateTime(1100, 14)
	require.Equal(t, 1, newEpochTicks)
	require.Equal(t, []uint32{12}, deltaBlocks)

	et.Reset(1100, 100)
	et.UpdateTime(1155, 20)
	require.Equal(t, []uint32{12, 20}, deltaBlocks)
}
