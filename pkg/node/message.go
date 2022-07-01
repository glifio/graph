package node

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"log"
	"math"

	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/glifio/graph/gql/model"
	"github.com/glifio/graph/internal/ulid"
	"github.com/glifio/graph/pkg/graph"
	"github.com/glifio/graph/pkg/kvdb"
	gocid "github.com/ipfs/go-cid"
	"google.golang.org/protobuf/proto"
)

func GetLotusMessage(cid gocid.Cid, ts *types.TipSet) (*api.Message, error) {
	val, err := kvdb.Open().Get([]byte("cid:" + cid.String()))
	if err != nil {
		log.Printf("fetching missing message: %s\n", cid)
		res, err := lotus.api.ChainGetMessage(context.Background(), cid)
		if err != nil {
			log.Printf("get message kv error %s %s\n", cid, err)
			return nil, err
		}
		msg := api.Message{Cid: cid, Message: res}
		err = SetMessage(msg, ts)
		return &msg, err
	}

	msg, err := DecodeLotusMessage(val)
	if err != nil {
		log.Printf("get message decode error %s %s\n", cid, err)
		return nil, err
	}
	msg.Cid = cid

	// add search index
	AddAddressToMessageIndex(context.Background(), msg, ts)

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
	var msg types.Message
	if err := msg.UnmarshalCBOR(bytes.NewReader(gmsg.MessageCbor)); err != nil {
		return nil, err
	}
	res := &SearchStateStruct{Message: api.InvocResult{Msg: &msg}}
	item := res.CreateMessage()
	item.Height = gmsg.Height
	return &item, nil
}

func SetMessage(message api.Message, ts *types.TipSet) error {
	db := kvdb.Open()

	// key is cid:[cid]
	keyMsg := []byte("cid:" + message.Cid.String())

	// encode message
	newMsg := &graph.Message{}
	newMsg.CopyMessage(message.Message, uint64(ts.Height()))
	val, err := proto.Marshal(newMsg)
	if err != nil {
		return err
	}

	// store message
	err = db.SetNX(keyMsg, val)
	if err != nil {
		return err
	}

	// add search index
	AddAddressToMessageIndex(context.Background(), &message, ts)

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

func AddAddressToMessageIndex(ctx context.Context, msg *api.Message, ts *types.TipSet) {
	db := kvdb.Open()

	// (a)ddr (m)essage (i)ndex  ami:[id]:[ulid]

	id, _ := generateULID(uint64(ts.Height()), msg.Message)

	to, err := GetIdAddress(ctx, msg.Message.To, ts.Key())
	if err == nil {
		keyTo := "ami:" + to.String() + ":" + id.String()
		_ = db.SetNX([]byte(keyTo), []byte(msg.Cid.String()))
	}

	// only save "from" reference if from != to
	if msg.Message.From.String() != msg.Message.To.String() {
		from, err := GetIdAddress(ctx, msg.Message.From, ts.Key())
		if err != nil {
			keyFrom := "ami:" + from.String() + ":" + id.String()
			_ = db.SetNX([]byte(keyFrom), []byte(msg.Cid.String()))
		}
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
