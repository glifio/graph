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
	"github.com/filecoin-project/go-state-types/big"
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
	SearchState(ctx context.Context, match Match, limit *int, offset *int, height int) ([]*SearchStateStruct, int, error)
	StateListMessages(ctx context.Context, addr string, lookback int)([]*lotusapi.InvocResult, error)
	StateDecodeParams(id address.Address, p2 abi.MethodNum, p3 []byte) (string, error)
	StateReplay(ctx context.Context, p1 types.TipSetKey, p2 cid.Cid) (*lotusapi.InvocResult, error)

	ChainHeadSub(ctx context.Context) (<-chan []*lotusapi.HeadChange, error)
	MpoolSub(ctx context.Context) (<-chan lotusapi.MpoolUpdate, error)
	Node() *Node
}

type Node struct {
	//api1 lotusapi.FullNodeStruct
	closer jsonrpc.ClientCloser
	api v0api.FullNodeStruct
	cache *ristretto.Cache
}

type SearchStateStruct struct {
	Tipset *types.TipSet
	Message lotusapi.InvocResult
}

func (state *SearchStateStruct) ConfirmedMessage(t *Node) model.MessageConfirmed {
	var item model.MessageConfirmed

	if state.Message.Msg != nil {
		item.Cid = state.Message.MsgCid.String()
		item.Height = int64(state.Tipset.Height())
		item.Version = int(state.Message.Msg.Version)
		item.From =  state.Message.Msg.From.String()
		item.To = state.Message.Msg.To.String()
		item.Nonce = state.Message.Msg.Nonce
		item.Value = state.Message.Msg.Value.String()
		item.GasLimit = state.Message.Msg.GasLimit
		item.GasFeeCap = state.Message.Msg.GasFeeCap.String()
		item.GasPremium = state.Message.Msg.GasPremium.String()
		item.Method = uint64(state.Message.Msg.Method)
	}

	item.GasUsed = state.Message.GasCost.GasUsed.Int64()
	item.GasBurned = state.Message.GasCost.OverEstimationBurn.Int64()
	item.MinerTip = state.Message.GasCost.MinerTip.String()
	item.BaseFeeBurn = state.Message.GasCost.BaseFeeBurn.String()
	item.OverEstimationBurn = state.Message.GasCost.OverEstimationBurn.String()
	item.Refund = state.Message.GasCost.Refund.String()
	item.MinerPenalty = state.Message.GasCost.MinerPenalty.String()
	item.MinerTip = state.Message.GasCost.MinerTip.String()

	//todo add tipsetkey
	params, err := t.StateDecodeParams(state.Message.Msg.To, state.Message.Msg.Method, state.Message.Msg.Params)

	if err == nil && params != "" {
		item.Params = &params
	}

	return item
}

func (t *Node) Init(cache *ristretto.Cache) error {
	t.cache = cache;
	return nil
}

func (t *Node) Node() *Node {
	return t
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

func (t *Node) StartCache(maxheight int){
	fmt.Println("cache -> init")

	// listen for new chainhead
	go func() {
		var current abi.ChainEpoch = 0

		for {
			for {
				log.Printf("cache -> subscribe to chainhead\n")
				chain, err := t.api.ChainNotify(context.Background())

				if err == nil {
					for headchanges := range chain {
						for _, elem := range headchanges {
							if current < elem.Val.Height() {
								current = elem.Val.Height()
								log.Printf("cache -> add tipset %s %s\n", elem.Val.Height()-1, elem.Type)
								t.ChainGetMessagesInTipset(context.Background(), elem.Val.Parents(), 1)
							}
						}
					}
				}

				log.Printf("cache -> subscription failed: %s\n", err)
				time.Sleep(15 * time.Second)
			}
		}
	}()	

	// backfill the cache
	log.Printf("cache -> backfill\n")
	ts, _ := t.api.ChainHead(context.Background())
	for i:=0; i<viper.GetInt("confidence"); i++ {
		if abi.ChainEpoch(maxheight) >= ts.Height() {
			break;
		}
		go func() {
			log.Printf("cache -> add tipset %s %d\n", ts.Height(), i)
			t.ChainGetMessagesInTipset(context.Background(), ts.Key(), i)
		}()
		ts, _ = t.ChainGetTipSet(context.Background(), ts.Parents())
	}
}


func (t *Node) GetActor(id string) (*types.Actor, error) {
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
	
	return t.api.MsigGetPending(context.Background(), res, types.EmptyTSK)
}

// match types take an *types.Message and return a bool value.
type Match func(*lotusapi.InvocResult) bool


func (t *Node) SearchState(ctx context.Context, match Match, limit *int, offset *int, height int) ([]*SearchStateStruct, int, error) {
	ts, _ := t.api.ChainHead(ctx)

	var t1 []*SearchStateStruct
	for i := 0; i < viper.GetInt("confidence"); i++ {
		//log.Printf("search state: %d height: %d\n", i, ts.Height()-1)
		msgs, err := t.ChainGetMessagesInTipset(ctx, ts.Key(), i)
		if err != nil {
			return nil, 0, xerrors.Errorf("failed to get messages for tipset (%s): %w", ts.Key(), err)
		}

		for _, iter := range msgs {
			if match(iter) {
				t1 = append(t1, &SearchStateStruct{Message: *iter, Tipset: ts})
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
	var res []*SearchStateStruct

	if(_offset > len(t1)){
		return res, _count, nil
	}
	if(_offset + _limit > len(t1)){
		_limit = len(t1)-_offset
	}

	res = t1[_offset:_offset+_limit]
	return res, _count, nil
}

func (t *Node) ChainGetMessagesInTipset(p0 context.Context, p1 types.TipSetKey, p3 int) ([]*lotusapi.InvocResult, error) {
	var key = "node/chain/tipset/messages/" + p1.String()

	// look in cache
	value, found := t.cache.Get(key)
	if found {
		res := value.([]*lotusapi.InvocResult)
		return res, nil
	}

	// get messages in tipset
	msgs, err := t.api.ChainGetMessagesInTipset(p0, p1)
	if err != nil {
		log.Printf("error tipset %s\n", err)
		return nil, err
	}

	// if we are close to chainhead don't run state compute
	if p3 < 1 {
		log.Printf("cache -> ignore t=%d\n", p3)
		var tinvoc []*lotusapi.InvocResult
		for _, iter := range msgs {
			tmp := lotusapi.InvocResult{
				Msg: iter.Message,
				MsgRct: nil,
				MsgCid: iter.Cid,
				GasCost: api.MsgGasCost{
					Message:            iter.Cid,
					GasUsed:            big.NewInt(0),
					BaseFeeBurn:        big.NewInt(0),
					OverEstimationBurn: big.NewInt(0),
					MinerPenalty:       big.NewInt(0),
					MinerTip:           big.NewInt(0),
					Refund:             big.NewInt(0),
					TotalCost:          big.NewInt(0),
				},
			}

			tinvoc = append(tinvoc, &tmp)
		}
		return tinvoc, nil
	}

	var tmsg []*types.Message
	for _, iter := range msgs {
		tmsg = append(tmsg, iter.Message)
	}

	res, err := t.api.StateCompute(p0, api.LookbackNoLimit, tmsg, p1)

	if err != nil {
		return nil, err
	}

	// add to cache
	log.Printf("cache -> done t=%d\n", p3)
	t.cache.SetWithTTL(key, res.Trace, 1, 60*time.Minute)

	return res.Trace, err
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
	t.cache.SetWithTTL(key, *tipset, 1, 60*time.Minute)

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
		replay, err := t.StateReplay(ctx, types.EmptyTSK, iter)
		if err != nil {
			log.Printf("StateListMessages: %s\n", err)
		} else {
			invoc = append(invoc, replay)
		}
	}

	return invoc, err 
}

func (t *Node) StateReplay(ctx context.Context, p1 types.TipSetKey, p2 cid.Cid) (*lotusapi.InvocResult, error) {
	var key = "node/state/replay/" + p2.String()

	value, found := t.cache.Get(key)
	if found {
		log.Printf("hit -> state replay cache: %s\n", key)
		res := value.(lotusapi.InvocResult)
		return &res, nil
	}

	res, err := t.api.StateReplay(ctx, p1, p2)
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
