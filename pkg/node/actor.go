package node

import (
	"context"
	"log"

	"github.com/filecoin-project/go-address"
	lotusapi "github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/chain/types"
	_ "github.com/lib/pq"
)

type Actor struct {
	api 	lotusapi.FullNodeStruct
}

func (t *Actor) Init(api lotusapi.FullNodeStruct) error {
	t.api = api;
	return nil
}

func (t *Actor) Get(id string) (*types.Actor, error) {
	addr, err := address.NewFromString(id)
	
	if err != nil {
		log.Println(err)
		return nil, err
	}
	
	actor, err := t.api.StateGetActor(context.Background(), addr, types.EmptyTSK )
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return actor, nil
}

// func (t *Actor) List() ([]lily.ActorItem, error) {
// 	// Prepare query, takes a name argument, protects from sql injection
// 	stmt, err := t.db.Db.Prepare("select m.id, m.code, m.head, m.nonce, m.balance, m.state_root, m.height from actors m limit 5")
// 	if err != nil {
// 		fmt.Println("Get Actors Preperation Err: ", err)
// 	}

// 	// Make query with our stmt
// 	rows, err := stmt.Query()
// 	if err != nil {
// 		fmt.Println("Get Actors Query Err: ", err)
// 	}

// 	if rows != nil {
// 		defer rows.Close()
// 	}
// 	if err != nil {
// 		return nil, err
// 	}

// 	// Create slice of Users for our response
// 	actors := []lily.ActorItem{}

// 	for rows.Next() {
// 		actor := lily.ActorItem{}

// 		err = rows.Scan(&actor.ID,
// 			&actor.Code,
// 			&actor.Head,
// 			&actor.Nonce,
// 			&actor.Balance,
// 			&actor.StateRoot,
// 			&actor.Height)

// 		if err != nil {
// 			return nil, err
// 		}
// 		actors = append(actors, actor)
// 	}
// 	if rows.Err() != nil {
// 		return nil, err
// 	}
// 	return actors, nil
// }

