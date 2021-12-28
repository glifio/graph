package postgres

import (
	"github.com/filecoin-project/lily/model/derived"
	_ "github.com/lib/pq"
)

type MessageConfirmed struct {
	db 	Database
}

func (t *MessageConfirmed) Init(db Database) error {
	t.db = db;
	return nil
}

func (t *MessageConfirmed) Get(id string) (*derived.GasOutputs, error) {

	// Select message
    var msgs []derived.GasOutputs
    var err = t.db.Db.Model(&msgs).Where("gas_outputs.cid = ?", id).Select()
	if err != nil {
		return nil, err
	}
	
	return &msgs[0], nil
}

func (t *MessageConfirmed) List(address *string, limit *int, offset *int) ([]derived.GasOutputs, error) {

	// t.db.Db.AddQueryHook(pgdebug.DebugHook{
	// 	// Print all queries.
	// 	Verbose: true,
	// })

	// Select messages
    var msgs []derived.GasOutputs
    var err = t.db.Db.Model(&msgs).Where("gas_outputs.from = ?", *address).WhereOr("gas_outputs.to = ?", *address).Limit(*limit).Offset(*offset).Select()
	if err != nil {
		return nil, err
	}

	return msgs, nil
}
