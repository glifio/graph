package postgres

import (
	"github.com/filecoin-project/lily/model/blocks"
	_ "github.com/lib/pq"
)

type BlockHeader struct {
	db 	Database
}

func (t *BlockHeader) Init(db Database) error {
	t.db = db;
	return nil
}

func (t *BlockHeader) GetByMessage(height int64, id string) (*blocks.BlockHeader, error) {

	var block []blocks.BlockHeader

    var err = t.db.Db.Model(&block).
		Join("JOIN block_messages AS bm").
		JoinOn("block_header.height = bm.height and block_header.cid = bm.block").
		Where("block_header.height = ?", height).
		Select()
	if err != nil {
		return nil, err
	}

	if len(block) == 0 {
		return nil, nil
	}

	return &block[0], nil
}

func (t *BlockHeader) List(address *string, limit *int, offset *int) ([]blocks.BlockHeader, error) {

	// t.db.Db.AddQueryHook(pgdebug.DebugHook{
	// 	// Print all queries.
	// 	Verbose: true,
	// })

	// Select messages
    var msgs []blocks.BlockHeader
    var err = t.db.Db.Model(&msgs).Where("gas_outputs.from = ?", *address).WhereOr("gas_outputs.to = ?", *address).Limit(*limit).Offset(*offset).Select()
	if err != nil {
		return nil, err
	}

	return msgs, nil
}
