package postgres

import (
	"log"
	"sort"
	"time"

	"github.com/dgraph-io/ristretto"
	"github.com/filecoin-project/lily/model/derived"
	"github.com/glifio/graph/gql/model"
	"github.com/glifio/graph/pkg/lily"
	_ "github.com/lib/pq"
)

type MessageConfirmed struct {
	db 	Database
	cache *ristretto.Cache
}

func (t *MessageConfirmed) Init(db Database, cache *ristretto.Cache) error {
	t.db = db;
	t.cache = cache

	// t.db.Db.AddQueryHook(pgdebug.DebugHook{
	// 	// Print all queries.
	// 	Verbose: true,
	// })

	return nil
}

func (t *MessageConfirmed) GetMaxHeight() (int, error) {
	var res struct {
		MaxHeight int
	}

	var msgs []derived.GasOutputs
	err := t.db.Db.Model(&msgs).ColumnExpr("max(height) AS max_height").Select(&res)

	if err != nil {
		return 0, err
	}

	return res.MaxHeight, nil
}

func (t *MessageConfirmed) Get(id string, height *int) (*lily.GasOutputs, *lily.ParsedMessage, error) {
	// Select message
    var msgs []lily.GasOutputs
    var parsed_msgs []lily.ParsedMessage
	var err error = nil

	if height != nil {
		err = t.db.Db.Model(&msgs).
			Where("gas_outputs.cid = ? and gas_outputs.height = ?", id, *height).
			Select()
	} else {
		err = t.db.Db.Model(&msgs).
			Where("gas_outputs.cid = ?", id).
			Select()
	}
	if err != nil {
		return nil, nil, err
	}

	if len(msgs) == 0 {
		return nil, nil, nil 
	}

	err = t.db.Db.Model(&parsed_msgs).
	Where("parsed_message.cid = ? and parsed_message.height=?", msgs[0].Cid, msgs[0].Height).
	Select()
	if err != nil {
		return &msgs[0], nil, nil
	}
	return &msgs[0], &parsed_msgs[0], nil
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

func (t *MessageConfirmed) Search(address *model.Address, limit int, offset int) ([]derived.GasOutputs, error) {
	var msgs []derived.GasOutputs

	value, found := t.cache.Get("msg:confirm:search:" + address.Robust)
	if found {
		log.Println("cache hit")
		msgs = value.([]derived.GasOutputs)
	} else {
		var err = t.db.Db.Model(&msgs).
			Where("gas_outputs.from = ?", address.ID).
			WhereOr("gas_outputs.from = ?", address.Robust).
			WhereOr("gas_outputs.to = ?", address.ID).		
			WhereOr("gas_outputs.to = ?", address.Robust).
			Order("height desc").
			// Limit(limit).
			// Offset(offset).
			Select()
		if err != nil {
			return nil, err
		}
		// set cache
		t.cache.SetWithTTL("msg:confirm:search:" + address.Robust, msgs, 1, 1*time.Minute)
	}

	// // sort the result by height desc
	// sort.Slice(msgs, func(i, j int) bool {
	// 	return msgs[i].Height > msgs[j].Height
	// })

	// limit and offset
	_limit := limit
	_offset := offset
	var res []derived.GasOutputs

	// no results 
	if len(msgs) == 0 {
		return res, nil
	}

	// offset bigger than results
	if(_offset > len(msgs)){
		return res, nil
	}

	// partial results 
	if(_offset + _limit > len(msgs)){
		_limit = len(msgs)-_offset
	}

    res = msgs[_offset:_offset+_limit]

	return res, nil
}
