package node

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"log"
	"math"

	badger "github.com/dgraph-io/badger/v3"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/glifio/graph/gql/model"
	"github.com/glifio/graph/internal/ulid"
	"github.com/glifio/graph/pkg/graph"
	"github.com/glifio/graph/pkg/kvdb"
	"github.com/ipfs/go-cid"
	gocid "github.com/ipfs/go-cid"
	"google.golang.org/protobuf/proto"
)

func GetLotusMessage(cid gocid.Cid, ts *types.TipSet, wb *badger.WriteBatch) (*api.Message, error) {
	val, err := kvdb.Open().Get([]byte("cid:" + cid.String()))
	if err != nil {
		log.Printf("fetching missing message: %s\n", cid)
		res, err := lotus.api.ChainGetMessage(context.Background(), cid)
		if err != nil {
			log.Printf("get message kv error %s %s\n", cid, err)
			return nil, err
		}
		msg := api.Message{Cid: cid, Message: res}
		_, err = SetMessage(msg, ts, wb)
		return &msg, err
	}

	msg, err := DecodeLotusMessage(val)
	if err != nil {
		log.Printf("get message decode error %s %s\n", cid, err)
		return nil, err
	}
	msg.Cid = cid

	return msg, err
}

func DecodeLotusMessage(val []byte) (*api.Message, error) {
	gmsg := &graph.Message{}
	if err := proto.Unmarshal(val, gmsg); err != nil {
		log.Fatalln("Failed to parse message:", err)
		return nil, err
	}
	var msg types.Message
	if err := msg.UnmarshalCBOR(bytes.NewReader(gmsg.MessageCbor)); err != nil {
		return nil, err
	}
	res := &api.Message{Message: &msg}
	return res, nil
}

func GetMessage(cid string) (*model.Message, error) {
	val, err := kvdb.Open().Get([]byte("cid:" + cid))
	if err != nil {
		return nil, err
	}

	msg, err := DecodeMessage(val)
	msg.Cid = cid

	return msg, err
}

func DecodeMessage(val []byte) (*model.Message, error) {
	gmsg := &graph.Message{}
	if err := proto.Unmarshal(val, gmsg); err != nil {
		log.Fatalln("Failed to parse message:", err)
		return nil, err
	}
	msg := types.Message{}
	if err := msg.UnmarshalCBOR(bytes.NewReader(gmsg.MessageCbor)); err != nil {
		return nil, err
	}
	res := &SearchStateStruct{Message: api.InvocResult{Msg: &msg}}
	item := res.CreateMessage()
	item.Height = gmsg.Height
	return &item, nil
}

func SetMessage(message api.Message, ts *types.TipSet, wb *badger.WriteBatch) (bool, error) {
	db := kvdb.Open()

	// key is cid:[cid]
	keyMsg := []byte("cid:" + message.Cid.String())

	// encode message
	buf := new(bytes.Buffer)
	if err := message.Message.MarshalCBOR(buf); err != nil {
		return false, err
	}

	val, err := proto.Marshal(&graph.Message{MessageCbor: buf.Bytes(), Height: uint64(ts.Height())})

	// newMsg := &graph.Message{}
	// newMsg.Serialize(message.Message, uint64(ts.Height()))
	// val, err := proto.Marshal(newMsg)

	if err != nil {
		return false, err
	}

	// store message
	created, err := db.SetNxWb(keyMsg, val, wb)
	if err != nil {
		return false, err
	}

	// add search index
	AddAddressToMessageIndex(context.Background(), &message, ts, wb)

	return created, err
}

func SetMessageNoIndex(message api.Message, ts *types.TipSet, wb *badger.WriteBatch) error {
	db := kvdb.Open()

	// key is cid:[cid]
	keyMsg := []byte("cid:" + message.Cid.String())

	// encode message
	buf := new(bytes.Buffer)
	if err := message.Message.MarshalCBOR(buf); err != nil {
		return err
	}

	val, err := proto.Marshal(&graph.Message{MessageCbor: buf.Bytes(), Height: uint64(ts.Height())})
	if err != nil {
		return err
	}

	// store message
	_, err = db.SetNxWb(keyMsg, val, wb)
	if err != nil {
		return err
	}

	return err
}

func generateULID(height uint64, message *types.Message) (ulid.ULID, error) {
	// use reverse height as the main sorting id
	id, err := ulid.NewReverse(height)
	if err != nil {
		return id, err
	}

	// use nonce as the secondary sorting id
	entropy := id.Entropy()
	binary.BigEndian.PutUint64(entropy, math.MaxUint64-message.Nonce)

	// use last bytes from address as the third sorting id
	from := message.From.Bytes()
	if len(from) > 2 {
		entropy[8] = from[len(from)-2]
		entropy[9] = from[len(from)-1]
	}

	id.SetEntropy(entropy)
	return id, err
}

func AddAddressToMessageIndex(ctx context.Context, msg *api.Message, ts *types.TipSet, wb *badger.WriteBatch) {
	// (a)ddr (m)essage (i)ndex  ami:[id]:[ulid]
	ulid, _ := generateULID(uint64(ts.Height()), msg.Message)

	SetAddressIndex(ctx, msg.Message.To, ulid, msg.Cid, ts.Key(), wb)

	// only save "from" reference if from != to
	if msg.Message.From.String() != msg.Message.To.String() {
		SetAddressIndex(ctx, msg.Message.From, ulid, msg.Cid, ts.Key(), wb)
	}
}

func SetAddressIndex(ctx context.Context, addr address.Address, ulid ulid.ULID, cid cid.Cid, tsk types.TipSetKey, wb *badger.WriteBatch) {
	db := kvdb.Open()
	id, err := GetIdAddress(ctx, addr, tsk, wb)
	if err == nil {
		key := "ami:" + id.String() + ":" + ulid.String()
		_, _ = db.SetNxWb([]byte(key), []byte(cid.String()), wb)
	}
}

func SearchMessagesByHeight(ctx context.Context, height uint64, limit *int, offset *int) ([]*model.Message, error) {
	_limit := 10
	_offset := 0

	if limit != nil {
		_limit = *limit
	}
	if offset != nil {
		_offset = *offset
	}

	var msgs []*model.Message
	log.Printf("search -> messages %d %d %d", height, _limit, _offset)

	// get the tipset key for height
	key := []byte(fmt.Sprintf("h:%d", height))
	b, err := kvdb.Open().Get(key)
	if err != nil {
		return nil, err
	}

	// get tipset messages
	key = []byte(fmt.Sprintf("tm:%s", string(b)))
	b, err = kvdb.Open().Get(key)
	if err != nil {
		return nil, err
	}

	tpsmsg := &graph.TipsetMessages{}
	if err := proto.Unmarshal(b, tpsmsg); err != nil {
		log.Fatalln("Failed to parse tipset messages:", err)
		return nil, err
	}

	for _, c := range tpsmsg.Cids {
		id, _ := gocid.Cast(c)
		msg, err := GetMessage(id.String())
		if err != nil {
			return nil, err
		}
		msgs = append(msgs, msg)
	}
	return msgs, nil
}

func SearchMessagesByAddress(ctx context.Context, address string, limit *int, offset *int) ([]*model.Message, error) {
	_limit := 10
	_offset := 0

	if limit != nil {
		_limit = *limit
	}
	if offset != nil {
		_offset = *offset
	}

	var msgs []*model.Message
	log.Printf("search -> messages %d %d", _limit, _offset)

	prefix := []byte("ami:" + address + ":")
	items, err := kvdb.Open().Search(prefix, uint(_limit), uint(_offset))
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		msg, err := GetMessage(string(item))
		if err != nil {
			return nil, err
		}
		msgs = append(msgs, msg)
	}
	return msgs, nil
}
