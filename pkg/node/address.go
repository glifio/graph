package node

import (
	"context"
	"log"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/glifio/graph/pkg/graph"
	"github.com/glifio/graph/pkg/kvdb"
	"google.golang.org/protobuf/proto"
)

func GetAddressFromID(ctx context.Context, id address.Address, tsk types.TipSetKey) (address.Address, error) {
	if id.Protocol() != address.ID {
		// already robust
		return id, nil
	}

	idval, _ := address.IDFromAddress(id)
	if idval <= 99 {
		// reserved address with no robust addr
		return id, nil
	}

	// see if we have the address in kv
	val, err := kvdb.Open().Get([]byte("ir:" + id.String()))
	if err == nil {
		return DecodeAddress(val)
	}

	// lookup address in lotus
	res, err := LookupRobustAddress(ctx, id, tsk)
	if err != nil {
		return res, err
	}

	log.Printf("address lookup: %s -> %s\n", id, res)

	// save the address in kv
	err = SetAddressFromID(id, res)

	return res, err
}

func GetIdAddress(ctx context.Context, addr address.Address, tsk types.TipSetKey) (address.Address, error) {
	if addr.Protocol() == address.ID {
		// already robust
		return addr, nil
	}

	// see if we have the address in kv
	val, err := kvdb.Open().Get(append([]byte("ai:"), addr.Payload()...))
	if err == nil {
		return DecodeIdAddress(val)
	}

	// lookup address in lotus
	res, err := LookupIdAddress(ctx, addr, types.EmptyTSK)
	if err != nil {
		log.Printf("id lookup: %s -> %s\n", addr, err)
		return res, err
	}

	//log.Printf("id lookup: %s -> %s\n", addr, res)

	// save the address in kv
	err = SetAddressToID(addr, res)

	return res, err
}

func DecodeAddress(val []byte) (address.Address, error) {
	a := &graph.Address{}
	if err := proto.Unmarshal(val, a); err != nil {
		log.Println("Failed to parse address:", err)
		return address.Address{}, err
	}
	return address.NewFromBytes(a.Address)
}

func DecodeIdAddress(val []byte) (address.Address, error) {
	a := &graph.Address{}
	if err := proto.Unmarshal(val, a); err != nil {
		log.Println("Failed to parse address:", err)
		return address.Address{}, err
	}
	return address.NewFromBytes(a.Id)
}

func LookupRobustAddress(ctx context.Context, addr address.Address, tsk types.TipSetKey) (address.Address, error) {
	if addr.Protocol() != address.ID {
		// already robust
		return addr, nil
	}

	// lookup robust
	return lotus.api.StateLookupRobustAddress(ctx, addr, tsk)
}

func LookupIdAddress(ctx context.Context, addr address.Address, tsk types.TipSetKey) (address.Address, error) {
	if addr.Protocol() == address.ID {
		// already id
		return addr, nil
	}

	// lookup robust
	return lotus.api.StateLookupID(ctx, addr, tsk)
}

func SetAddressFromID(id address.Address, robust address.Address) error {
	db := kvdb.Open()

	// key is ir:[id]
	key := []byte("ir:" + id.String())

	// encode address
	a := &graph.Address{}
	a.Id = id.Bytes()
	a.Address = robust.Bytes()

	val, err := proto.Marshal(a)
	if err != nil {
		return err
	}

	// store adddress
	return db.SetNXs(key, val)
}

func SetAddressToID(robust address.Address, id address.Address) error {
	db := kvdb.Open()

	// key is ai:[robust]
	key := append([]byte("ai:"), robust.Payload()...)

	// encode address
	a := &graph.Address{}
	a.Id = id.Bytes()
	a.Address = robust.Bytes()

	val, err := proto.Marshal(a)
	if err != nil {
		return err
	}

	// store adddress
	return db.SetNXs(key, val)
}
