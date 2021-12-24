package postgres

import (
	"github.com/glifio/graph/pkg/lily"
	_ "github.com/lib/pq"
)

type Actor struct {
	db 	Database
}

func (t *Actor) Init(db Database) error {
	t.db = db;
	return nil
}

func (t *Actor) Get(id string) (*lily.ActorItem, error) {
	// Prepare query, takes a name argument, protects from sql injection
	// stmt, err := t.db.Db.Prepare("select m.id, m.code, m.head, m.nonce, m.balance, m.state_root, m.height from actors m where m.id = $1")
	// if err != nil {
	// 	fmt.Println("Get Actor Preperation Err: ", err)
	// }
	var actor lily.ActorItem
	// Make query with our stmt, passing in name argument
	// err = stmt.QueryRow(id).Scan(&actor.ID,
	// 	&actor.Code,
	// 	&actor.Head,
	// 	&actor.Nonce,
	// 	&actor.Balance,
	// 	&actor.StateRoot,
	// 	&actor.Height)
	// if err != nil {
	// 	fmt.Println("GetMessages Query Err: ", err)
	// }

	// if err != nil {
	// 	return nil, err
	// }
	return &actor, nil
}

func (t *Actor) List() ([]lily.ActorItem, error) {
	// Prepare query, takes a name argument, protects from sql injection
	// stmt, err := t.db.Db.Prepare("select m.id, m.code, m.head, m.nonce, m.balance, m.state_root, m.height from actors m limit 5")
	// if err != nil {
	// 	fmt.Println("Get Actors Preperation Err: ", err)
	// }

	// // Make query with our stmt
	// rows, err := stmt.Query()
	// if err != nil {
	// 	fmt.Println("Get Actors Query Err: ", err)
	// }

	// if rows != nil {
	// 	defer rows.Close()
	// }
	// if err != nil {
	// 	return nil, err
	// }

	// Create slice of Users for our response
	actors := []lily.ActorItem{}

	// for rows.Next() {
	// 	actor := lily.ActorItem{}

	// 	err = rows.Scan(&actor.ID,
	// 		&actor.Code,
	// 		&actor.Head,
	// 		&actor.Nonce,
	// 		&actor.Balance,
	// 		&actor.StateRoot,
	// 		&actor.Height)

	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	actors = append(actors, actor)
	// }
	// if rows.Err() != nil {
	// 	return nil, err
	// }
	return actors, nil
}

