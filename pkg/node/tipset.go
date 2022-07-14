package node

import (
	"bytes"
	"context"
	"fmt"
	"log"

	badger "github.com/dgraph-io/badger/v3"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/glifio/graph/pkg/graph"
	"github.com/glifio/graph/pkg/kvdb"
	gocid "github.com/ipfs/go-cid"
	"google.golang.org/protobuf/proto"
)

func SetTipSet(ts *types.TipSet, wb *badger.WriteBatch) error {
	db := kvdb.Open()
	key := []byte("t:" + ts.Key().String())

	buf := new(bytes.Buffer)
	if err := ts.MarshalCBOR(buf); err != nil {
		return err
	}
	val := buf.Bytes()

	// store tipset
	if _, err := db.SetNxWb(key, val, wb); err != nil {
		return err
	}

	// store reference to tipset height
	keyHeight := "h:" + ts.Height().String()
	_, err := db.SetNxWb([]byte(keyHeight), []byte(ts.Key().String()), wb)
	return err
}

func GetTipSetByHeight(height uint64) (*types.TipSet, error) {
	ts := &types.TipSet{}
	db := kvdb.Open()
	key := []byte(fmt.Sprintf("h:%d", height))

	val, err := db.Get(key)
	if err == badger.ErrKeyNotFound {
		// get tipset from node
		if ts, err = lotus.api.ChainGetTipSetByHeight(context.Background(), abi.ChainEpoch(height), types.EmptyTSK); err != nil {
			log.Printf("tipset -> error: %s\n", err)
			return nil, err
		}

		// add to badger
		// err = SetTipSet(ts)
		return ts, err
	}

	// get tipset from kv store
	tskey := []byte(fmt.Sprintf("t:%s", string(val)))
	tsval, err := db.Get(tskey)
	if err != nil {
		return nil, err
	}

	// unmarshal tipset
	if err := ts.UnmarshalCBOR(bytes.NewReader(tsval)); err != nil {
		return nil, err
	}

	return ts, nil
}

func ExistsTipSet(tsk types.TipSetKey) bool {
	db := kvdb.Open()

	key := []byte("t:" + tsk.String())

	if _, err := db.Get(key); err == badger.ErrKeyNotFound {
		return false
	}
	return true
}

func GetTipSet(tsk types.TipSetKey, wb *badger.WriteBatch) (*types.TipSet, error) {
	db := kvdb.Open()

	ts := &types.TipSet{}
	key := []byte("t:" + tsk.String())

	if val, err := db.Get(key); err == badger.ErrKeyNotFound {
		// get tipset from node
		if ts, err = lotus.api.ChainGetTipSet(context.Background(), tsk); err != nil {
			log.Printf("sync -> error: %s\n", err)
			return nil, err
		}
		// add to badger
		SetTipSet(ts, wb)
	} else {
		if err := ts.UnmarshalCBOR(bytes.NewReader(val)); err != nil {
			return nil, err
		}
	}
	return ts, nil
}

func GetTipSetMessages(ts *types.TipSet, wb *badger.WriteBatch) ([]api.Message, error) {
	db := kvdb.Open()
	key := []byte(fmt.Sprintf("tm:%s", ts.Key()))

	// get tipset messages
	val, err := db.Get(key)
	if err == badger.ErrKeyNotFound {
		// get tipset from node
		tsm, err := lotus.api.ChainGetMessagesInTipset(context.Background(), ts.Key())
		if err != nil {
			log.Printf("sync -> error: %s\n", err)
			return nil, err
		}
		// store messages
		for _, msg := range tsm {
			SetMessage(msg, ts, wb)
		}

		// store list of tipset messages
		SetTipSetMessages(ts.Key(), tsm, wb)

		return tsm, err
	}

	tpsmsgs := &graph.TipsetMessages{}
	if err := proto.Unmarshal(val, tpsmsgs); err != nil {
		log.Fatalln("Failed to parse tipset messages:", err)
		return nil, err
	}

	var msgs []api.Message
	for _, c := range tpsmsgs.Cids {
		cid, _ := gocid.Cast(c)
		msg, err := GetLotusMessage(cid, ts, wb)
		if err != nil {
			log.Printf("error tipset messages: %s\n", cid.String())
		} else {
			// make sure address index is up to date
			//AddAddressToMessageIndex(context.Background(), msg, ts)
			msgs = append(msgs, *msg)
		}
	}
	return msgs, nil
}

func UpdateTipSetMessages(ts *types.TipSet, wb *badger.WriteBatch) (bool, error) {
	db := kvdb.Open()
	key := []byte(fmt.Sprintf("tm:%s", ts.Key()))

	// get tipset messages
	if !db.Exists(key) {
		// get tipset from node
		tsm, err := lotus.api.ChainGetMessagesInTipset(context.Background(), ts.Key())
		if err != nil {
			log.Printf("sync -> error: %s\n", err)
			return false, err
		}

		// store messages
		for _, msg := range tsm {
			SetMessage(msg, ts, wb)
		}

		// store list of tipset messages
		return SetTipSetMessages(ts.Key(), tsm, wb)
	}
	return false, nil
}

func SetTipSetMessages(tsk types.TipSetKey, messages []api.Message, wb *badger.WriteBatch) (bool, error) {
	db := kvdb.Open()
	key := []byte(fmt.Sprintf("tm:%s", tsk.String()))

	// grab the cids
	cids := [][]byte{}
	for _, msg := range messages {
		cids = append(cids, msg.Cid.Bytes())
	}

	// store list of cids
	tm := &graph.TipsetMessages{Cids: cids}
	val, err := proto.Marshal(tm)
	if err != nil {
		return false, err
	}
	return db.SetNxWb(key, val, wb)
}
