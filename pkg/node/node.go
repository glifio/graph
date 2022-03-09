package node

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/dgraph-io/ristretto"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-jsonrpc"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/lotus/api"
	lotusapi "github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/api/v0api"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/glifio/graph/gql/model"
	"github.com/ipfs/go-cid"
	"github.com/spf13/viper"
	"golang.org/x/xerrors"
)

type NodeInterface interface {
	GetActor(id string) (*types.Actor, error)
	GetPending() ([]*types.SignedMessage, error)
	GetMessage(cidcc string) (*types.Message, error)
	StateSearchMsg(id string) (*lotusapi.MsgLookup, error)
	AddressLookup(id string) (*model.Address, error)
	MsigGetPending(addr string) ([]*lotusapi.MsigTransaction, error)
	SearchState(ctx context.Context, addr address.Address, limit *int, offset *int, height int) ([]*model.MessageConfirmed, int, error)
	StateListMessages(ctx context.Context, addr string, lookback int)([]*lotusapi.InvocResult, error)
	StateDecodeParams(id address.Address, p2 abi.MethodNum, p3 []byte) (string, error)
	StateReplay(ctx context.Context, id string) (*lotusapi.InvocResult, error)

	ChainHeadSub(ctx context.Context) (<-chan []*lotusapi.HeadChange, error)
	MpoolSub(ctx context.Context) (<-chan lotusapi.MpoolUpdate, error)
}

type Node struct {
	//api1 lotusapi.FullNodeStruct
	closer jsonrpc.ClientCloser
	api v0api.FullNodeStruct
	cache *ristretto.Cache
}

func (t *Node) Init(cache *ristretto.Cache) error {
	t.cache = cache;
	return nil
}

func (t *Node) Connect(address1 string, token string){
	head := http.Header{}

	if token != "" {
		head.Set("Authorization","Bearer " + token)
	}
	
	var err error
	t.closer, err = jsonrpc.NewMergeClient(context.Background(), 
		address1, 
		"Filecoin", 
		api.GetInternalStructs(&t.api), 
		head)
	if err != nil {
		log.Fatalf("connecting with lotus failed: %s", err)
	}

	name, _ := t.api.StateNetworkName(context.Background())
	fmt.Println("network name: ", name)
	if name == "mainnet" {
		address.CurrentNetwork = address.Mainnet
		fmt.Println("address network : mainnet")
	} else {
		address.CurrentNetwork = address.Testnet
		fmt.Println("address network : testnet")
	}
}

func (t *Node) Close(){
	t.closer()
}


func (t *Node) GetActor(id string) (*types.Actor, error) {
	addr, err := address.NewFromString(id)
	if err != nil {
		log.Fatal(err)
	}
	
	tipset, err := t.api.ChainHead(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	actor, err := t.api.StateGetActor(context.Background(), addr, tipset.Key() )
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("actor cid: ", actor.Code)
	fmt.Println("actor bal: ", actor.Balance)

	return actor, nil
}

func (t *Node) GetMessage(id string) (*types.Message, error) {
	c, err := cid.Decode(id)
	if err != nil {
		log.Fatal(err)
	}
	
	msg, err := t.api.ChainGetMessage(context.Background(), c )
	if err != nil {
		log.Fatal(err)
	}

	return msg, nil
}

func (t *Node) StateSearchMsg(id string) (*lotusapi.MsgLookup, error){
	c, err := cid.Decode(id)	
	if err != nil {
		log.Print(err)
		return nil, err
	}

	var key = "node/state/search/" + id

	value, found := t.cache.Get(key)
	if found {
		log.Println(key)
		res := value.(lotusapi.MsgLookup)
		return &res, nil
	}

	confidence := viper.GetInt("confidence")
	msg, err := t.api.StateSearchMsgLimited(context.Background(), c, abi.ChainEpoch(confidence))
	if err != nil {
		log.Print(err)
		return nil, err
	}

	// set cache
	if msg != nil {
		t.cache.SetWithTTL(key, *msg, 1, 5*time.Minute)
	}

	return msg, err
}

func (t *Node) StateDecodeParams(id address.Address, p2 abi.MethodNum, p3 []byte) (string, error){
	var res string
	if p3 == nil {
		return res, nil
	}
	obj, err := t.api.StateDecodeParams(context.Background(), id, p2, p3, types.EmptyTSK )
	if err != nil {
		return res, err
	}
	parambytes, err := json.Marshal(obj)
	if err != nil {
		return res, err
	}
	res = string(parambytes)

	return res, err
}

func (t *Node) MsigGetPending(addr string) ([]*lotusapi.MsigTransaction, error) {
	res, err := address.NewFromString(addr)
	if err != nil {
		return nil, err
	}
	
	pending, err := t.api.MsigGetPending(context.Background(), res, types.EmptyTSK)

	return pending, err
}

func (t *Node) SearchState(ctx context.Context, addr address.Address, limit *int, offset *int, height int) ([]*model.MessageConfirmed, int, error) {
	ts, _ := t.api.ChainHead(ctx)

	match, _ := t.AddressLookup(addr.String())

	if match != nil {
		_, err := t.api.StateLookupID(ctx, addr, types.EmptyTSK)

		// if the recipient doesn't exist at the start point, we're not gonna find any matches
		if xerrors.Is(err, types.ErrActorNotFound) {
			return nil, 0, nil
		}

		if err != nil {
			return nil, 0, xerrors.Errorf("looking up match: %w", err)
		}
	}

	matchFunc := func(msg *types.Message) bool {		
		if match.ID == msg.From.String() || match.ID == msg.To.String() || 
		match.Robust == msg.From.String() || match.Robust == msg.To.String() {
			return true
		}
		return false
	}

	type SearchStateStruct struct {
		message *types.Message
		ts *types.TipSet
	}

	var t1 []*SearchStateStruct
	for i := 0; i < viper.GetInt("confidence"); i++ {
		msgs, err := t.ChainGetMessagesInTipset(ctx, ts.Key())
		if err != nil {
			return nil, 0, xerrors.Errorf("failed to get messages for tipset (%s): %w", ts.Key(), err)
		}

		for _, iter := range msgs {
			if matchFunc(iter.Message) {
				// todo create custom struct
				t1 = append(t1, &SearchStateStruct{message: iter.Message, ts: ts})
			}
		}

		if ts.Height() == 0 {
			break
		}

		if len(t1) >= *offset+*limit {
			log.Printf("search state limit hit: %d dist:%d\n", ts.Height(), i)
			break
		}		

		next, err := t.ChainGetTipSet(ctx, ts.Parents())
		if err != nil {
			return nil, 0, xerrors.Errorf("loading next tipset: %w", err)
		}

		// check if we reached the lookback height
		if next.Height() <= abi.ChainEpoch(height) {
			log.Printf("search state lookback hit: %d dist:%d msgs:%d\n", next.Height(), i, len(t1))
			break
		}
		ts = next
	}

	_limit := *limit
	_offset := *offset
	_count := len(t1)
	var res []*model.MessageConfirmed

	if(_offset > len(t1)){
		return res, _count, nil
	}
	if(_offset + _limit > len(t1)){
		_limit = len(t1)-_offset
	}

	t2 := t1[_offset:_offset+_limit]

	for _, iter := range t2 {
		var item model.MessageConfirmed
		item.Cid = iter.message.Cid().String()
		item.Height = int64(iter.ts.Height())
		item.Value = iter.message.Value.String()
		item.From = iter.message.From.String()
		item.To = iter.message.To.String()
		item.Nonce = iter.message.Nonce
		item.Version = int(iter.message.Version)
		item.GasFeeCap = iter.message.GasFeeCap.String()
		item.GasLimit = iter.message.GasLimit
		item.GasPremium = iter.message.GasPremium.String()
		item.Method = uint64(iter.message.Method)
		obj, err := t.StateDecodeParams(iter.message.To, iter.message.Method, iter.message.Params)
		if err == nil && obj != "" {
			item.Params = &obj
		}

		// replay, err := t.api.StateReplay(ctx, iter.ts.Key(), iter.message.Cid)
		// if err == nil {
		// 	item.MinerTip = replay.GasCost.MinerTip.String()
		// 	item.BaseFeeBurn = replay.GasCost.BaseFeeBurn.String()
		// 	item.OverEstimationBurn = replay.GasCost.OverEstimationBurn.String()		
		// }
		res = append(res, &item)
	}

	return res, _count, nil
}

func (t *Node) ChainGetMessagesInTipset(p0 context.Context, p1 types.TipSetKey) ([]lotusapi.Message, error) {
	var key = "node/chain/messages/tipset/" + p1.String()

	// look in cache
	value, found := t.cache.Get(key)
	if found {
		res := value.([]lotusapi.Message)
		return res, nil
	}

	// get messages in tipset
	msgs, err := t.api.ChainGetMessagesInTipset(p0, p1)
	if err != nil {
		return nil, err
	}

	// add to cache
	t.cache.SetWithTTL(key, msgs, 1, 30*time.Minute)

	return msgs, err
}

func (t *Node) ChainGetTipSet(p0 context.Context, p1 types.TipSetKey) (*types.TipSet, error){
	var key = "node/chain/tipset/" + p1.String()

	// look in cache
	value, found := t.cache.Get(key)
	if found {
		res := value.(types.TipSet)
		return &res, nil
	}

	// get messages in tipset
	tipset, err := t.api.ChainGetTipSet(p0, p1)
	if err != nil {
		return nil, err
	}

	// add to cache
	t.cache.SetWithTTL(key, *tipset, 1, 30*time.Minute)

	return tipset, err
}

func (t *Node) StateListMessages(ctx context.Context, addr string, lookback int)([]*lotusapi.InvocResult, error){
	var out []cid.Cid
	var res []cid.Cid

	tipset, err := t.api.ChainHead(ctx)
	if err != nil {
		return nil, err
	}

	//lookback := 5

	robust, _ := t.AddressGetRobust(addr)
	id, _ := t.AddressGetID(addr)

	if !id.Empty() {
		res, err = t.api.StateListMessages(ctx, &lotusapi.MessageMatch{From: id}, types.EmptyTSK, tipset.Height()-abi.ChainEpoch(lookback))
		if err == nil {
			out = append(out, res...)
		}
		res, err = t.api.StateListMessages(ctx, &lotusapi.MessageMatch{To: id}, types.EmptyTSK, tipset.Height()-abi.ChainEpoch(lookback))
		if err == nil {
			out = append(out, res...)
		}
	}

	if !robust.Empty() {
		res, err = t.api.StateListMessages(ctx, &lotusapi.MessageMatch{From: robust}, types.EmptyTSK, tipset.Height()-abi.ChainEpoch(lookback))
		if err == nil {
			out = append(out, res...)
		}

		res, err = t.api.StateListMessages(ctx, &lotusapi.MessageMatch{To: robust}, types.EmptyTSK, tipset.Height()-abi.ChainEpoch(lookback))
		if err == nil {
			out = append(out, res...)
		}	
	}

	var invoc []*lotusapi.InvocResult
	for _, iter := range out {
		replay, err := t.StateReplay(ctx, iter.String())
		if err != nil {
			log.Printf("StateListMessages: %s\n", err)
		} else {
			invoc = append(invoc, replay)
		}
	}

	return invoc, err 
}

func (t *Node) StateReplay(ctx context.Context, id string) (*lotusapi.InvocResult, error) {
	var key = "node/state/replay/" + id

	value, found := t.cache.Get(key)
	if found {
		log.Println(key)
		res := value.(lotusapi.InvocResult)
		return &res, nil
	}

	c, err := cid.Decode(id)
	if err != nil {
		return nil, err
	}

	res, err := t.api.StateReplay(ctx, types.EmptyTSK, c)
	if err != nil {
		return nil, err
	}

	t.cache.SetWithTTL(key, *res, 1, 5*time.Minute)

	return res, err
}

func (t *Node) ChainHead(ctx context.Context) (*types.TipSet, error) {
	tipset, err := t.api.ChainHead(ctx)
	return tipset, err
}

func (t *Node) ChainHeadSub(ctx context.Context) (<-chan []*lotusapi.HeadChange, error) {
	headchange, err := t.api.ChainNotify(ctx)
	return headchange, err
}

func (t *Node) MpoolSub(ctx context.Context) (<-chan lotusapi.MpoolUpdate, error) {
	mpool, err := t.api.MpoolSub(ctx)
	return mpool, err
}

func (t *Node) GetPending() ([]*types.SignedMessage, error) {

	tipset, err := t.api.ChainHead(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	status, err := t.api.MpoolPending(context.Background(), tipset.Key())
	if err != nil {
		log.Fatal(err)
	}

	return status, nil
}

func (t *Node) AddressLookup(id string) (*model.Address, error){
	var key = "node/addr/lookup/" + id

	value, found := t.cache.Get(key)
	if found {
		res := value.(model.Address)
		return &res, nil
	}

	result := &model.Address{ID: "", Robust: ""}
	addr, err := address.NewFromString(id)
	if err != nil {
		log.Println(err)
		return nil, err		
	}

	var rs address.Address
	switch(addr.Protocol()){
		case address.ID:
			//protocol = ID
			result.ID = addr.String()
			rs, err = t.api.StateAccountKey(context.Background(), addr, types.EmptyTSK)
			if err == nil {
				result.Robust = rs.String()
			}
		default:
			result.Robust = addr.String()
			rs, err = t.api.StateLookupID(context.Background(), addr, types.EmptyTSK)
			if err == nil {
				result.ID = rs.String()
			}
	}

	// set a value with a cost of 1
	t.cache.SetWithTTL(key, *result, 1, 30*time.Minute)

	return result, nil
}

func (t *Node) AddressGetID(id string) (address.Address, error){
	addr, err := address.NewFromString(id)
	if err != nil {
		return addr, err
	}
	var rs address.Address
	switch(addr.Protocol()){
		case address.ID:
			//protocol = ID
			return addr, nil
		default:
			rs, err = t.api.StateLookupID(context.Background(), addr, types.EmptyTSK)
			if err != nil {
				return rs, err
			}
			return rs, nil
	}
}

func (t *Node) AddressGetRobust(id string) (address.Address, error){
	addr, err := address.NewFromString(id)
	if err != nil {
		return addr, err
	}
	var rs address.Address
	switch(addr.Protocol()){
		case address.ID:
			//protocol = ID
			rs, err = t.api.StateAccountKey(context.Background(), addr, types.EmptyTSK)
			if err != nil {
				return rs, err
			}
			return rs, nil
		default:
			return addr, nil
	}
}
