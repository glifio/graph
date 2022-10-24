package node

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"time"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/big"
	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/lotus/node/modules/dtypes"
	"github.com/fxamacker/cbor"
	"github.com/glifio/graph/gql/model"
	"github.com/glifio/graph/pkg/kvdb"
	"github.com/glifio/graph/pkg/postgres"
	"github.com/ipfs/go-cid"
	"github.com/spf13/viper"
	"golang.org/x/xerrors"
)

type NodeInterface interface {
	GetActor(id string) (*types.Actor, error)
	GetPending() ([]*types.SignedMessage, error)
	GetMessage(cidcc string) (*types.Message, error)
	StateSearchMsg(id string) (*api.MsgLookup, error)
	MsigGetPending(addr string) ([]*api.MsigTransaction, error)
	SearchState(ctx context.Context, match Match, limit *int, offset *int, height int) ([]*SearchStateStruct, int, error)
	StateListMessages(ctx context.Context, addr string, lookback int) ([]*api.InvocResult, error)
	StateReplay(ctx context.Context, p1 types.TipSetKey, p2 cid.Cid) (*api.InvocResult, error)

	ChainHeadSub(ctx context.Context) (<-chan []*api.HeadChange, error)
	MpoolSub(ctx context.Context) (<-chan api.MpoolUpdate, error)
	Node() *Node
}

type Node struct {
	//api1 api.FullNodeStruct
	// closer jsonrpc.ClientCloser
	// api    v0api.FullNodeStruct
	// cache *ristretto.Cache
	// db     postgres.Database
	ticker *time.Ticker
}

type SearchStateStruct struct {
	Tipset  *types.TipSet
	Message api.InvocResult
}

func (state *SearchStateStruct) ConfirmedMessage() model.MessageConfirmed {
	var item model.MessageConfirmed

	if state.Message.Msg != nil {
		item.Cid = state.Message.MsgCid.String()
		item.Height = int64(state.Tipset.Height())
		item.Version = int(state.Message.Msg.Version)
		item.From = state.Message.Msg.From.String()
		item.To = state.Message.Msg.To.String()
		item.Nonce = state.Message.Msg.Nonce
		item.Value = state.Message.Msg.Value.String()
		item.GasLimit = state.Message.Msg.GasLimit
		item.GasFeeCap = state.Message.Msg.GasFeeCap.String()
		item.GasPremium = state.Message.Msg.GasPremium.String()
		item.Method = uint64(state.Message.Msg.Method)
		item.Params = base64.StdEncoding.EncodeToString(state.Message.Msg.Params)
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
	return item
}

func (state *SearchStateStruct) CreateMessage() model.Message {
	item := model.Message{}

	if state.Message.Msg != nil {
		item.Cid = state.Message.MsgCid.String()
		item.Version = state.Message.Msg.Version
		item.From = state.Message.Msg.From.String()
		item.To = state.Message.Msg.To.String()
		item.Nonce = state.Message.Msg.Nonce
		item.Value = state.Message.Msg.Value.String()
		item.GasLimit = state.Message.Msg.GasLimit
		item.GasFeeCap = state.Message.Msg.GasFeeCap.String()
		item.GasPremium = state.Message.Msg.GasPremium.String()
		item.Method = uint64(state.Message.Msg.Method)
		item.Params = base64.StdEncoding.EncodeToString(state.Message.Msg.Params)
	}

	if state.Tipset != nil {
		item.Height = uint64(state.Tipset.Height())
	}

	return item
}

func (t *Node) Node() *Node {
	return t
}

func (t *Node) Connect(address1 string, token string) (dtypes.NetworkName, error) {
	lotus := GetLotusInstance(&LotusOptions{address: address1})

	name, _ := lotus.api.StateNetworkName(context.Background())
	fmt.Println("network name: ", name)
	if name == "mainnet" {
		address.CurrentNetwork = address.Mainnet
		log.Println("address network : mainnet")
	} else {
		address.CurrentNetwork = address.Testnet
		log.Println("address network : testnet")
	}
	return name, nil
}

func (t *Node) Close() {
	if lotus.closer != nil {
		lotus.closer()
	}
	postgres.GetInstanceDB().Close()
	kvdb.Open().Close()
	t.SyncTimerStop()
}

func (t *Node) GetActor(id string) (*types.Actor, error) {
	addr, err := address.NewFromString(id)

	if err != nil {
		log.Println(err)
		return nil, err
	}

	actor, err := lotus.api.StateGetActor(context.Background(), addr, types.EmptyTSK)
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

	msg, err := lotus.api.ChainGetMessage(context.Background(), c)
	if err != nil {
		log.Fatal(err)
	}

	return msg, nil
}

func (t *Node) StateSearchMsg(id string) (*api.MsgLookup, error) {
	c, err := cid.Decode(id)
	if err != nil {
		log.Print(err)
		return nil, err
	}

	var key = "node/state/search/" + id

	value, found := GetCacheInstance().cache.Get(key)
	if found {
		log.Println(key)
		res := value.(api.MsgLookup)
		return &res, nil
	}

	msg, err := lotus.api.StateSearchMsg(context.Background(), types.EmptyTSK, c, api.LookbackNoLimit, true)
	if err != nil {
		log.Print(err)
		return nil, err
	}

	// set cache
	if msg != nil {
		GetCacheInstance().cache.SetWithTTL(key, *msg, 1, 5*time.Minute)
	}

	return msg, err
}

func (t *Node) MsigGetPending(addr string) ([]*api.MsigTransaction, error) {
	res, err := address.NewFromString(addr)
	if err != nil {
		return nil, err
	}

	return lotus.api.MsigGetPending(context.Background(), res, types.EmptyTSK)
}

// match types take an *types.Message and return a bool value.
type Match func(*api.InvocResult) bool

func (t *Node) SearchState(ctx context.Context, match Match, limit *int, offset *int, height int) ([]*SearchStateStruct, int, error) {
	// look in cache
	var ts *types.TipSet
	var err error

	// get last tipsetkey from cache or default to chainhead
	value, found := GetCacheInstance().cache.Get("node/chainhead/tipsetkey")
	if found {
		tsk := value.(types.TipSetKey)
		ts, err = t.ChainGetTipSet(ctx, tsk)
		log.Println("cache: hit tipsetkey")
		if err != nil {
			ts, _ = lotus.api.ChainHead(ctx)
		}

	} else {
		ts, _ = lotus.api.ChainHead(ctx)
	}

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

	if _offset > len(t1) {
		return res, _count, nil
	}
	if _offset+_limit > len(t1) {
		_limit = len(t1) - _offset
	}

	res = t1[_offset : _offset+_limit]
	return res, _count, nil
}

func (t *Node) ChainGetMessagesInTipset(p0 context.Context, p1 types.TipSetKey, p3 int) ([]*api.InvocResult, error) {
	var key = "node/chain/tipset/messages/" + p1.String()
	cache := GetCacheInstance().cache
	// look in cache
	value, found := cache.Get(key)
	if found {
		res := value.([]*api.InvocResult)
		return res, nil
	}

	// get messages in tipset
	msgs, err := lotus.api.ChainGetMessagesInTipset(p0, p1)
	if err != nil {
		log.Printf("error tipset %s\n", err)
		return nil, err
	}

	// if we are close to chainhead don't run state compute
	if p3 < 1 {
		log.Printf("cache -> ignore t=%d\n", p3)
		var tinvoc []*api.InvocResult
		for _, iter := range msgs {
			tmp := api.InvocResult{
				Msg:    iter.Message,
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

	res, err := lotus.api.StateCompute(p0, api.LookbackNoLimit, tmsg, p1)

	if err != nil {
		return nil, err
	}

	type ExecReturnBytes struct {
		_             struct{} `cbor:",toarray"`
		IDAddress     []byte
		RobustAddress []byte
	}

	for _, iter := range res.Trace {
		if iter.Msg.To.String()[1:] == "01" && iter.Msg.Method == 2 {
			log.Println("found msg to 01....!", iter.MsgRct.ExitCode)
			if iter.MsgRct.ExitCode.IsSuccess() {
				var v ExecReturnBytes
				err := cbor.Unmarshal(iter.MsgRct.Return, &v)
				log.Println(err)
				idaddr, _ := address.NewFromBytes(v.IDAddress)
				roaddr, _ := address.NewFromBytes(v.RobustAddress)
				log.Println(idaddr)
				log.Println(roaddr)
				modelAddr := model.Address{
					ID:     idaddr.String(),
					Robust: roaddr.String(),
				}
				cache.SetWithTTL(addressLookupKey+idaddr.String(), modelAddr, 1, 60*time.Minute)
				cache.SetWithTTL(addressLookupKey+roaddr.String(), modelAddr, 1, 60*time.Minute)
			}
		}
	}

	// add to cache
	//log.Printf("cache -> tipset %d %s\n", p3, "done")
	cache.SetWithTTL(key, res.Trace, 1, 60*time.Minute)

	return res.Trace, err
}

func (t *Node) ChainGetTipSet(p0 context.Context, p1 types.TipSetKey) (*types.TipSet, error) {
	var key = "node/chain/tipset/" + p1.String()
	cache := GetCacheInstance().cache

	// look in cache
	value, found := cache.Get(key)
	if found {
		res := value.(types.TipSet)
		return &res, nil
	}

	// get messages in tipset
	tipset, err := lotus.api.ChainGetTipSet(p0, p1)
	if err != nil {
		return nil, err
	}

	// add to cache
	cache.SetWithTTL(key, *tipset, 1, 60*time.Minute)

	return tipset, err
}

func (t *Node) StateListMessages(ctx context.Context, addr string, lookback int) ([]*api.InvocResult, error) {
	var out []cid.Cid
	var res []cid.Cid

	tipset, err := lotus.api.ChainHead(ctx)
	if err != nil {
		return nil, err
	}

	//lookback := 5

	robust, _ := t.AddressGetRobust(addr)
	id, _ := t.AddressGetID(addr)

	if !id.Empty() {
		res, err = lotus.api.StateListMessages(ctx, &api.MessageMatch{From: id}, types.EmptyTSK, tipset.Height()-abi.ChainEpoch(lookback))
		if err == nil {
			out = append(out, res...)
		}
		res, err = lotus.api.StateListMessages(ctx, &api.MessageMatch{To: id}, types.EmptyTSK, tipset.Height()-abi.ChainEpoch(lookback))
		if err == nil {
			out = append(out, res...)
		}
	}

	if !robust.Empty() {
		res, err = lotus.api.StateListMessages(ctx, &api.MessageMatch{From: robust}, types.EmptyTSK, tipset.Height()-abi.ChainEpoch(lookback))
		if err == nil {
			out = append(out, res...)
		}

		res, err = lotus.api.StateListMessages(ctx, &api.MessageMatch{To: robust}, types.EmptyTSK, tipset.Height()-abi.ChainEpoch(lookback))
		if err == nil {
			out = append(out, res...)
		}
	}

	var invoc []*api.InvocResult
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

func (t *Node) StateReplay(ctx context.Context, p1 types.TipSetKey, p2 cid.Cid) (*api.InvocResult, error) {
	var key = "node/state/replay/" + p2.String()
	cache := GetCacheInstance().cache

	value, found := cache.Get(key)
	if found {
		log.Printf("hit -> state replay cache: %s\n", key)
		res := value.(api.InvocResult)
		return &res, nil
	}

	res, err := lotus.api.StateReplay(ctx, p1, p2)
	if err != nil {
		log.Printf("state replay: %s\n", err)
		return nil, err
	}

	cache.SetWithTTL(key, *res, 1, 5*time.Minute)

	return res, err
}

func (t *Node) ChainHead(ctx context.Context) (*types.TipSet, error) {
	tipset, err := lotus.api.ChainHead(ctx)
	return tipset, err
}

func (t *Node) ChainHeadSub(ctx context.Context) (<-chan []*api.HeadChange, error) {
	headchange, err := lotus.api.ChainNotify(ctx)
	return headchange, err
}

func (t *Node) MpoolSub(ctx context.Context) (<-chan api.MpoolUpdate, error) {
	mpool, err := lotus.api.MpoolSub(ctx)
	return mpool, err
}

func (t *Node) GetPending() ([]*types.SignedMessage, error) {

	tipset, err := lotus.api.ChainHead(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	status, err := lotus.api.MpoolPending(context.Background(), tipset.Key())
	if err != nil {
		log.Fatal(err)
	}

	return status, nil
}

const addressLookupKey = "node/addr/lookup/"

func AddressConvert(id string) (*model.Address, error) {
	result := &model.Address{ID: "", Robust: ""}

	addr, err := address.NewFromString(id)
	if err != nil {
		return nil, err
	}

	switch addr.Protocol() {
	case address.ID:
		result.ID = addr.String()
	default:
		result.Robust = addr.String()
	}
	return result, nil
}

func AddressLookup(id string) (*model.Address, error) {
	var key = addressLookupKey + id
	cache := GetCacheInstance().cache

	value, found := cache.Get(key)
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

	switch addr.Protocol() {
	case address.ID:
		result.ID = addr.String()
		wb := kvdb.Open().Batch()
		defer wb.Cancel()
		robust, err := GetRobustAddress(context.Background(), addr, types.EmptyTSK, wb)
		wb.Flush()
		if err == nil {
			result.Robust = robust.String()
		}
	default:
		result.Robust = addr.String()
		wb := kvdb.Open().Batch()
		defer wb.Cancel()
		id, err := GetIdAddress(context.Background(), addr, types.EmptyTSK, wb)
		wb.Flush()
		if err == nil {
			result.ID = id.String()
		}
	}

	if result.ID != "" && result.Robust != "" {
		cache.SetWithTTL(key, *result, 1, 60*time.Minute)
	}

	return result, nil
}

func (t *Node) AddressGetID(id string) (address.Address, error) {
	addr, err := address.NewFromString(id)
	if err != nil {
		return addr, err
	}
	var rs address.Address
	switch addr.Protocol() {
	case address.ID:
		//protocol = ID
		return addr, nil
	default:
		rs, err = lotus.api.StateLookupID(context.Background(), addr, types.EmptyTSK)
		if err != nil {
			return rs, err
		}
		return rs, nil
	}
}

func (t *Node) AddressGetRobust(id string) (address.Address, error) {
	addr, err := address.NewFromString(id)
	if err != nil {
		return addr, err
	}
	var rs address.Address
	switch addr.Protocol() {
	case address.ID:
		//protocol = ID
		rs, err = lotus.api.StateAccountKey(context.Background(), addr, types.EmptyTSK)
		if err != nil {
			return rs, err
		}
		return rs, nil
	default:
		return addr, nil
	}
}
