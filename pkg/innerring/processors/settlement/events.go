package settlement

import (
	"github.com/nspcc-dev/neofs-node/pkg/morph/event"
)

// BasicIncomeEvent is an event of the beginning of basic income cash settlements.
type BasicIncomeEvent struct {
	epoch      uint64
	blockIndex uint32
}

// MorphEvent implements Neo: FS chain event.
func (e BasicIncomeEvent) MorphEvent() {}

// Epoch returns the number of the epoch
// in which the event was generated.
func (e BasicIncomeEvent) Epoch() uint64 {
	return e.epoch
}

// BlockIndex returns the index of the FS chain block
// that triggered the event.
func (e BasicIncomeEvent) BlockIndex() uint32 {
	return e.blockIndex
}

// NewBasicIncomeEvent for epoch triggered by the block with the given index.
func NewBasicIncomeEvent(epoch uint64, blockIndex uint32) event.Event {
	return BasicIncomeEvent{
		epoch:      epoch,
		blockIndex: blockIndex,
	}
}
