package postgres

import (
	"fmt"

	"github.com/glifio/graph/pkg/lily"
	_ "github.com/lib/pq"
)

type Message struct {
	db 	Database
}

func (t *Message) Init(db Database) error {
	t.db = db;
	return nil
}

func (t *Message) Get(id string) (*lily.MessageItem, error) {
	// Prepare query, takes a name argument, protects from sql injection
	stmt, err := t.db.Db.Prepare("select m.id, m.code, m.head, m.nonce, m.balance, m.state_root, m.height from actors m where m.id = $1")
	if err != nil {
		fmt.Println("Get Actor Preperation Err: ", err)
	}
	var msg lily.MessageItem
	
	// Make query with our stmt, passing in name argument
	err = stmt.QueryRow(id).Scan(&msg.Cid,
		&msg.Height,
		&msg.From,
		&msg.To,
		&msg.Value,
		&msg.Method,
		&msg.Params,)

	if err != nil {
		fmt.Println("Get Messages Query Err: ", err)
		return nil, err
	}

	return &msg, nil
}

func (t *Message) List(limit int, offset int) ([]lily.MessageItem, error) {
	// Prepare query, takes a name argument, protects from sql injection
	// stmt, err := t.db.Db.Prepare("select m.cid, m.height, m.from, m.to, m.value, m.method, m.params from parsed_messages m limit 5 offset $2")
	// if err != nil {
	// 	fmt.Println("GetMessages Preperation Err: ", err)
	// }

	rows, err := t.db.Db.Query("select m.cid, m.height, m.from, m.to, m.value, m.method, m.params from parsed_messages m limit $1 offset $2", limit, offset)
	// Make query with our stmt
	//rows, err := stmt.Query(offset)
	if err != nil {
		fmt.Println("Get Messages Query Err: ", err)
	}

	if rows != nil {
		defer rows.Close()
	}
	if err != nil {
		return nil, err
	}

	messages := []lily.MessageItem{}

	for rows.Next() {
		msg := lily.MessageItem{}

		err := rows.Scan(
			&msg.Cid,
			&msg.Height,
			&msg.From,
			&msg.To,
			&msg.Value,
			&msg.Method,
			&msg.Params,
		)

		if err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}
	if rows.Err() != nil {
		return nil, err
	}
	return messages, nil
}

