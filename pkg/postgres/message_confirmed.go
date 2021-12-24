package postgres

import (
	"github.com/filecoin-project/lily/model/derived"
	"github.com/glifio/graph/pkg/lily"
	_ "github.com/lib/pq"
)

type MessageConfirmed struct {
	db 	Database
}

func (t *MessageConfirmed) Init(db Database) error {
	t.db = db;
	return nil
}

func (t *MessageConfirmed) Get(id string) (*lily.MessageConfirmedItem, error) {

	// Select message by primary key.
    // user := &derived.GasOutputs{Id: user1.Id}
    // err = db.Model(user).WherePK().Select()
    // if err != nil {
    //     panic(err)
    // }

	// Prepare query, takes a name argument, protects from sql injection
	// stmt, err := t.db.Db.Prepare("select m.id, m.code, m.head, m.nonce, m.balance, m.state_root, m.height from actors m where m.id = $1")
	// if err != nil {
	// 	fmt.Println("Get Actor Preperation Err: ", err)
	// }
	var msg lily.MessageConfirmedItem
	
	// Make query with our stmt, passing in name argument
	// err = stmt.QueryOne(id).Scan(&msg.Cid,
	// 	&msg.Height,
	// 	&msg.From,
	// 	&msg.To,
	// 	&msg.Value,
	// 	&msg.Method,
	// 	&msg.Params,)

	// if err != nil {
	// 	fmt.Println("Get Messages Query Err: ", err)
	// 	return nil, err
	// }

	return &msg, nil
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
