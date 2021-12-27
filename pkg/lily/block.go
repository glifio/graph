package lily

import "github.com/filecoin-project/lily/model/blocks"

type BlockInterface interface {
	GetByMessage(height int64, id string) (*blocks.BlockHeader, error)
}
