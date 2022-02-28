package postgres

import (
	"sort"

	"github.com/filecoin-project/lily/model/derived"
	"github.com/glifio/graph/gql/model"
	"github.com/glifio/graph/pkg/lily"
	_ "github.com/lib/pq"
)

type MessageConfirmed struct {
	db 	Database
}

func (t *MessageConfirmed) Init(db Database) error {
	t.db = db;

	// t.db.Db.AddQueryHook(pgdebug.DebugHook{
	// 	// Print all queries.
	// 	Verbose: true,
	// })

	return nil
}

func (t *MessageConfirmed) Get(id string, height *int) (*lily.GasOutputs, error) {
	// Select message
    var msgs []lily.GasOutputs
	var err error = nil

	if height != nil {
		err = t.db.Db.Model(&msgs).
			Relation("ParsedMessage").  // left join parsed msg to get params
			Where("gas_outputs.cid = ? and gas_outputs.height = ?", id, *height).
			Select()
	} else {
		err = t.db.Db.Model(&msgs).
			Relation("ParsedMessage").  // left join parsed msg to get params
			Where("gas_outputs.cid = ?", id).
			Select()
	}
	if err != nil {
		return nil, err
	}

	if len(msgs) == 0 {
		return nil, nil 
	}
	
	return &msgs[0], nil
}

func (t *MessageConfirmed) List(address *string, limit *int, offset *int) ([]derived.GasOutputs, error) {
	// Select messages
    var msgs []derived.GasOutputs
    var err = t.db.Db.Model(&msgs).
		Where("gas_outputs.from = ?", *address).
		WhereOr("gas_outputs.to = ?", *address).
		Select()
	if err != nil {
		return nil, err
	}

	// sort the result by height desc
	sort.Slice(msgs, func(i, j int) bool {
		return msgs[i].Height > msgs[j].Height
	})

	// limit and offset
	_limit := *limit
	_offset := *offset
	var res []derived.GasOutputs
	if(_offset > len(msgs)){
		return res, nil
	}
	if(_offset + _limit > len(msgs)){
		_limit = len(msgs)-_offset
	}
    res = msgs[_offset:_offset+_limit]

	return res, nil
}

func (t *MessageConfirmed) Search(address *model.Address, limit *int, offset *int) ([]derived.GasOutputs, error) {
	// Select messages
    var msgs []derived.GasOutputs
    var err = t.db.Db.Model(&msgs).
		Where("gas_outputs.from = ?", address.ID).
		WhereOr("gas_outputs.from = ?", address.Robust).
		WhereOr("gas_outputs.to = ?", address.ID).
		WhereOr("gas_outputs.to = ?", address.Robust).
		Select()
	if err != nil {
		return nil, err
	}

	// sort the result by height desc
	sort.Slice(msgs, func(i, j int) bool {
		return msgs[i].Height > msgs[j].Height
	})

	// limit and offset
	_limit := *limit
	_offset := *offset
	var res []derived.GasOutputs
	if(_offset > len(msgs)){
		return res, nil
	}
	if(_offset + _limit > len(msgs)){
		_limit = len(msgs)-_offset
	}
    res = msgs[_offset:_offset+_limit]

	return res, nil
}
