package graph

import "github.com/filecoin-project/lotus/chain/types"

func (m *Message) Serialize(msg *types.Message, height uint64) {
	val, _ := msg.Serialize()
	*m = Message{
		Height:      height,
		MessageCbor: val,
	}
}
