package postgres

import (
	"context"
	"encoding/json"
	"log"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/big"
	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/glifio/graph/pkg/lily"
	"github.com/ipfs/go-cid"
	"github.com/jackc/pgtype"
	_ "github.com/lib/pq"
)

type Message struct {
}

func (t *Message) Get(id string) (*lily.MessageItem, error) {
	// Prepare query, takes a name argument, protects from sql injection
	// stmt, err := t.db.Db.Prepare("select m.id, m.code, m.head, m.nonce, m.balance, m.state_root, m.height from actors m where m.id = $1")
	// if err != nil {
	// 	fmt.Println("Get Actor Preperation Err: ", err)
	// }
	var msg lily.MessageItem

	// Make query with our stmt, passing in name argument
	// err = stmt.QueryRow(id).Scan(&msg.Cid,
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

func GetMessagesInTipset(height uint) ([]api.Message, error) {
	db := GetInstanceDB().pgx
	rows, err := db.Query(context.Background(), "select m.cid, 1, m.to, m.from, m.nonce, m.value, m.gas_limit, m.gas_fee_cap, m.gas_premium, m.method, p.params from messages m left join parsed_messages p using (height, cid) where m.height = $1", height)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer rows.Close()

	var res []api.Message

	for rows.Next() {
		var _cid string
		var _version uint64
		var _to address.Address
		var _from address.Address
		var _nonce uint64
		var _value pgtype.Numeric
		var _gaslimit int64
		var _gasfeecap pgtype.Numeric
		var _gaspremium pgtype.Numeric
		var _method uint64
		var _params []byte

		if err := rows.Scan(&_cid, &_version, &_to, &_from, &_nonce, &_value, &_gaslimit, &_gasfeecap, &_gaspremium, &_method, &_params); err != nil {
			log.Fatal(err)
		}

		bytes, _ := json.Marshal(string(_params))

		newcid, _ := cid.Decode(_cid)
		newmsg := &types.Message{
			Version:    _version,
			To:         _to,
			From:       _from,
			Nonce:      _nonce,
			Value:      big.NewFromGo(_value.Int),
			GasLimit:   _gaslimit,
			GasFeeCap:  big.NewFromGo(_gasfeecap.Int),
			GasPremium: big.NewFromGo(_gaspremium.Int),
			Method:     abi.MethodNum(_method),
			Params:     bytes,
		}
		res = append(res, api.Message{Cid: newcid, Message: newmsg})
	}

	return res, rows.Err()
}

func (t *Message) List(limit int, offset int) ([]lily.MessageItem, error) {
	// Prepare query, takes a name argument, protects from sql injection
	// stmt, err := t.db.Db.Prepare("select m.cid, m.height, m.from, m.to, m.value, m.method, m.params from parsed_messages m limit 5 offset $2")
	// if err != nil {
	// 	fmt.Println("GetMessages Preperation Err: ", err)
	// }

	//	rows, err := t.db.Db.Query("select m.cid, m.height, m.from, m.to, m.value, m.method, m.params from parsed_messages m limit $1 offset $2", limit, offset)
	// Make query with our stmt
	//rows, err := stmt.Query(offset)
	// if err != nil {
	// 	fmt.Println("Get Messages Query Err: ", err)
	// }

	// if err != nil {
	// 	return nil, err
	// }

	messages := []lily.MessageItem{}

	// for rows.Next() {
	// 	msg := lily.MessageItem{}

	// 	err := rows.Scan(
	// 		&msg.Cid,
	// 		&msg.Height,
	// 		&msg.From,
	// 		&msg.To,
	// 		&msg.Value,
	// 		&msg.Method,
	// 		&msg.Params,
	// 	)

	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	messages = append(messages, msg)
	// }
	// if rows.Err() != nil {
	// 	return nil, err
	// }
	return messages, nil
}
