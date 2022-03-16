package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/filecoin-project/lily/model/derived"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/specs-actors/actors/builtin"
	"github.com/filecoin-project/specs-actors/actors/builtin/multisig"
	"github.com/glifio/graph/gql/generated"
	"github.com/glifio/graph/gql/model"
	util "github.com/glifio/graph/internal/utils"
	"github.com/google/uuid"
	"github.com/ipfs/go-cid"
	"github.com/jinzhu/copier"
	"golang.org/x/crypto/blake2b"
)

func (r *messageConfirmedResolver) From(ctx context.Context, obj *model.MessageConfirmed) (*model.Address, error) {
	addr, err := r.NodeService.AddressLookup(obj.From)
	return addr, err
}

func (r *messageConfirmedResolver) To(ctx context.Context, obj *model.MessageConfirmed) (*model.Address, error) {
	addr, err := r.NodeService.AddressLookup(obj.To)
	return addr, err
}

func (r *messageConfirmedResolver) MethodName(ctx context.Context, obj *model.MessageConfirmed) (string, error) {
	if obj.ActorName == "" {
		return "", nil
	}
	switch strings.Split(obj.ActorName, "/")[2] {
	case "account":
		return reflect.ValueOf(&builtin.MethodsAccount).Elem().Type().Field(int(obj.Method)).Name, nil
	case "init":
		return reflect.ValueOf(&builtin.MethodsInit).Elem().Type().Field(int(obj.Method)).Name, nil
	case "reward":
		return reflect.ValueOf(&builtin.MethodsReward).Elem().Type().Field(int(obj.Method)).Name, nil
	case "multisig":
		return reflect.ValueOf(&builtin.MethodsMultisig).Elem().Type().Field(int(obj.Method)).Name, nil
	case "paymentchannel":
		return reflect.ValueOf(&builtin.MethodsPaych).Elem().Type().Field(int(obj.Method)).Name, nil
	case "storagemarket":
		return reflect.ValueOf(&builtin.MethodsMarket).Elem().Type().Field(int(obj.Method)).Name, nil
	case "storageminer":
		return reflect.ValueOf(&builtin.MethodsMiner).Elem().Type().Field(int(obj.Method)).Name, nil
	case "storagepower":
		return reflect.ValueOf(&builtin.MethodsPower).Elem().Type().Field(int(obj.Method)).Name, nil
	default:
		return "???", nil
	}
}

func (r *messageConfirmedResolver) Block(ctx context.Context, obj *model.MessageConfirmed) (*model.Block, error) {
	block, err := r.BlockService.GetByMessage(obj.Height, obj.Cid)
	var item model.Block
	copier.Copy(&item, &block)
	return &item, err
}

func (r *queryResolver) Block(ctx context.Context, address string, height int64) (*model.Block, error) {
	block, err := r.BlockService.GetByMessage(height, address)
	if err != nil {
		return nil, err
	}
	var item model.Block
	copier.Copy(&item, &block)
	return &item, err
}

func (r *queryResolver) Message(ctx context.Context, p1 string, height *int) (*model.MessageConfirmed, error) {
	limit := 1
	offset := 0

	msgCID, _ := cid.Decode(p1)
	maxheight, _ := r.MessageConfirmedService.GetMaxHeight()

	// Look in State
	matchFunc := func(msg *types.Message) bool {		
		return msg.Cid().Equals(msgCID)
	}

	r1, count, err := r.NodeService.SearchState(ctx, matchFunc, &limit, &offset, maxheight)

	if err == nil && count == 1 {
		iter := r1[0]
		var item model.MessageConfirmed
		item.Cid = iter.Message.Cid.String()
		item.Height = int64(iter.Tipset.Height())
		item.Value = iter.Message.Message.Value.String()
		item.From = iter.Message.Message.From.String()
		item.To = iter.Message.Message.To.String()
		item.Nonce = iter.Message.Message.Nonce
		item.Version = int(iter.Message.Message.Version)
		item.GasFeeCap = iter.Message.Message.GasFeeCap.String()
		item.GasLimit = iter.Message.Message.GasLimit
		item.GasPremium = iter.Message.Message.GasPremium.String()
		item.Method = uint64(iter.Message.Message.Method)
		obj, err := r.NodeService.StateDecodeParams(iter.Message.Message.To, iter.Message.Message.Method, iter.Message.Message.Params)
		if err == nil && obj != "" {
			item.Params = &obj
		}

		replay, err := r.NodeService.StateReplay(ctx, iter.Message.Cid.String())
		if err == nil {
			item.MinerTip = replay.GasCost.MinerTip.String()
			item.BaseFeeBurn = replay.GasCost.BaseFeeBurn.String()
			item.OverEstimationBurn = replay.GasCost.OverEstimationBurn.String()		
		}
		log.Printf("message: found in state: %d\n", item.Nonce)
		return &item, nil
	}

	// only look in state
	if(*height == -1){
		return nil, nil
	}

	// Look in Lily
	msg, parsed, err := r.MessageConfirmedService.Get(p1, height)
	if err != nil {
		return nil, err
	}

	if msg == nil {
		return nil, nil
	}

	var item model.MessageConfirmed
	copier.Copy(&item, &msg)
	if parsed != nil {
		item.Params = &parsed.Params
	}
	log.Printf("message: found in lily: %d\n", item.Nonce)
	return &item, err
}

func (r *queryResolver) Messages(ctx context.Context, address string, limit *int, offset *int) ([]*model.MessageConfirmed, error) {
	var items []*model.MessageConfirmed
	//	var stateMsgs []*model.MessageConfirmed

	maxheight, _ := r.MessageConfirmedService.GetMaxHeight()

	addr, err := r.NodeService.AddressLookup(address)
	if err != nil {
		return nil, err
	}
	
	matchFunc := func(msg *types.Message) bool {		
		if len(addr.ID)>1 && (addr.ID[1:] == msg.From.String()[1:] || addr.ID[1:] == msg.To.String()[1:]) {
			return true
		}
		if len(addr.Robust)>1 && (addr.Robust[1:] == msg.From.String()[1:] || addr.Robust[1:] == msg.To.String()[1:]) {
			return true
		}
		return false
	}

	r1, count, err := r.NodeService.SearchState(ctx, matchFunc, limit, offset, maxheight)
	if err == nil {
		for _, iter := range r1 {
			var item model.MessageConfirmed
			item.Cid = iter.Message.Cid.String()
			item.Height = int64(iter.Tipset.Height())
			item.Value = iter.Message.Message.Value.String()
			item.From = iter.Message.Message.From.String()
			item.To = iter.Message.Message.To.String()
			item.Nonce = iter.Message.Message.Nonce
			item.Version = int(iter.Message.Message.Version)
			item.GasFeeCap = iter.Message.Message.GasFeeCap.String()
			item.GasLimit = iter.Message.Message.GasLimit
			item.GasPremium = iter.Message.Message.GasPremium.String()
			item.Method = uint64(iter.Message.Message.Method)
			obj, err := r.NodeService.StateDecodeParams(iter.Message.Message.To, iter.Message.Message.Method, iter.Message.Message.Params)
			if err == nil && obj != "" {
				item.Params = &obj
			}

			// replay, err := t.api.StateReplay(ctx, iter.ts.Key(), iter.message.Cid)
			// if err == nil {
			// 	item.MinerTip = replay.GasCost.MinerTip.String()
			// 	item.BaseFeeBurn = replay.GasCost.BaseFeeBurn.String()
			// 	item.OverEstimationBurn = replay.GasCost.OverEstimationBurn.String()		
			// }
			items = append(items, &item)
		}
	}

	log.Printf("messages: found in state: %d\n", len(r1))

	if len(r1) >= *limit {
		return items, nil
	}

	var lily_offset int
	lily_limit := *limit - len(r1)

	if len(r1) > 0 {
		// if we found part of the messages in state then offset in lily should be zero
		lily_offset = 0
	} else {
		// if we found no part of the messages in state then offset in lily should be less the count in state
		lily_offset = *offset - count
	}

	r2, err := r.MessageConfirmedService.Search(addr, lily_limit, lily_offset)
	if err != nil {
		return nil, err
	}
	for _, r := range r2 {
		var item model.MessageConfirmed
		copier.Copy(&item, &r)
		items = append(items, &item)
	}

	log.Printf("messages: found in lily: %d\n", len(r2))
	return items, nil
}

func (r *queryResolver) PendingMessage(ctx context.Context, cid string) (*model.MessagePending, error) {
	pending, err := r.NodeService.GetPending()

	if err != nil {
		return nil, err
	}

	for _, item := range pending {
		if item.Cid().String() == cid {
			var msg model.MessagePending
			msg.Cid = item.Cid().String()
			msg.Version = strconv.FormatUint(item.Message.Version, 10)
			msg.Method = item.Message.Method.String()
			msg.GasFeeCap = new(string)
			var gasfeecap = item.Message.GasFeeCap.String()
			msg.GasFeeCap = &gasfeecap

			msg.Value = item.Message.Value.String()

			fromaddr, _ := r.NodeService.AddressLookup(item.Message.From.String())
			msg.From = fromaddr
			toaddr, _ := r.NodeService.AddressLookup(item.Message.To.String())
			msg.To = toaddr

			msg.GasPremium = new(string)
			var gasPremium = item.Message.GasPremium.String()
			msg.GasPremium = &gasPremium

			var gaslimit = strconv.FormatInt(item.Message.GasLimit, 10)
			msg.GasLimit = &gaslimit

			obj, err := r.NodeService.StateDecodeParams(item.Message.To, item.Message.Method, item.Message.Params)

			if err == nil && obj != "" {
				msg.Params = &obj
			}

			return &msg, nil
		}
	}
	return nil, nil
}

func (r *queryResolver) PendingMessages(ctx context.Context, address *string, limit *int, offset *int) ([]*model.MessagePending, error) {
	var items []*model.MessagePending

	pending, err := r.NodeService.GetPending()

	if err != nil {
		return nil, err
	}

	queryAddress, _ := r.NodeService.AddressLookup(*address)

	for _, item := range pending {
		var msg model.MessagePending

		if queryAddress.Robust == item.Message.From.String() || queryAddress.Robust == item.Message.To.String() ||
			queryAddress.ID == item.Message.From.String() || queryAddress.ID == item.Message.To.String() {

			msg.Cid = item.Cid().String()
			msg.Version = strconv.FormatUint(item.Message.Version, 10)
			msg.Method = item.Message.Method.String()
			msg.GasFeeCap = new(string)
			var gasfeecap = item.Message.GasFeeCap.String()
			msg.GasFeeCap = &gasfeecap

			// todo optimize
			fromaddr, _ := r.NodeService.AddressLookup(item.Message.From.String())
			msg.From = fromaddr
			toaddr, _ := r.NodeService.AddressLookup(item.Message.To.String())
			msg.To = toaddr

			msg.Value = item.Message.Value.String()

			msg.GasPremium = new(string)
			var gasPremium = item.Message.GasPremium.String()
			msg.GasPremium = &gasPremium

			var gaslimit = strconv.FormatInt(item.Message.GasLimit, 10)
			msg.GasLimit = &gaslimit

			obj, err := r.NodeService.StateDecodeParams(item.Message.To, item.Message.Method, item.Message.Params)
			if err == nil && obj != "" {
				msg.Params = &obj
			}

			items = append(items, &msg)
		}
	}
	return items, nil
}

func (r *queryResolver) MpoolPending(ctx context.Context, address *string) ([]*model.MpoolUpdate, error) {
	var pool []*model.MpoolUpdate
	return pool, nil
}

func (r *queryResolver) MessagesConfirmed(ctx context.Context, address *string, limit *int, offset *int) ([]*model.MessageConfirmed, error) {
	var items []*model.MessageConfirmed
	var rs []derived.GasOutputs

	addr, err := r.NodeService.AddressLookup(*address)
	if err != nil {
		return nil, err
	}

	rs, err = r.MessageConfirmedService.Search(addr, *limit, *offset)
	if err != nil {
		return nil, err
	}
	for _, r := range rs {
		var item model.MessageConfirmed
		copier.Copy(&item, &r)
		items = append(items, &item)
	}
	return items, nil
}

func (r *queryResolver) Address(ctx context.Context, str string) (*model.Address, error) {
	addr, err := r.NodeService.AddressLookup(str)
	return addr, err
}

func (r *queryResolver) Actor(ctx context.Context, address string) (*model.Actor, error) {
	// TODO get this data from lily instead of the node
	item, err := r.NodeService.GetActor(address)
	if err != nil {
		return nil, err
	} else {
		return &model.Actor{
			ID:      address,
			Code:    item.Code.String(),
			Head:    item.Head.String(),
			Nonce:   strconv.FormatUint(item.Nonce, 10),
			Balance: item.Balance.String(),
			// StateRoot: item.StateRoot,
			// Height:    item.Height,
		}, nil
	}
}

func (r *queryResolver) Actors(ctx context.Context) ([]*model.Actor, error) {
	panic(fmt.Errorf("not implemented"))

	// var items []*model.Actor
	// var savedItems []lily.ActorItem
	// savedItems, err := r.ActorService.List()
	// if err != nil {
	// 	return nil, err
	// }
	// for i, savedItem := range savedItems {
	// 	var item model.Actor
	// 	savedItem = savedItems[i]
	// 	item.ID = savedItem.ID
	// 	item.Code = savedItem.Code
	// 	item.Head = savedItem.Head
	// 	item.StateRoot = savedItem.StateRoot
	// 	item.Nonce = savedItem.Nonce
	// 	item.Balance = savedItem.Balance
	// 	item.Height = savedItem.Height
	// 	items = append(items, &item)
	// }
	// return items, nil
}

func (r *queryResolver) MsigPending(ctx context.Context, address *string, limit *int, offset *int) ([]*model.MsigTransaction, error) {
	var items []*model.MsigTransaction
	//var rs []derived.GasOutputs
	pending, err := r.NodeService.MsigGetPending(*address)
	if err != nil {
		return nil, err
	}

	if len(pending) < *offset {
		return nil, nil
	}

	for _, iter := range pending[*offset:util.Min(*offset+*limit, len(pending))] {
		var item model.MsigTransaction
		item.ID = iter.ID
		item.Method = uint64(iter.Method)
		obj, err := r.NodeService.StateDecodeParams(iter.To, iter.Method, iter.Params)

		if err == nil && obj != "" {
			item.Params = &obj
		}

		txn := &multisig.Transaction{To: iter.To, Value: iter.Value, Method: iter.Method, Params: iter.Params, Approved: iter.Approved}
		calculatedHash, _ := multisig.ComputeProposalHash(txn, blake2b.Sum256)
		item.ProposalHash = base64.StdEncoding.EncodeToString(calculatedHash)

		toaddr, _ := r.NodeService.AddressLookup(iter.To.String())
		item.To = toaddr
		item.Value = iter.Value.String()
		for _, appr := range iter.Approved {
			approvedaddr, err := r.NodeService.AddressLookup(appr.String())
			if err == nil {
				item.Approved = append(item.Approved, approvedaddr)
			}
		}
		items = append(items, &item)
	}
	return items, nil
}

func (r *queryResolver) StateListMessages(ctx context.Context, address string, lookback *int) ([]*model.MessageConfirmed, error) {
	var items []*model.MessageConfirmed

	pending, err := r.NodeService.StateListMessages(ctx, address, *lookback)
	if err != nil {
		return nil, err
	}

	for _, iter := range pending {
		var item model.MessageConfirmed
		// res, _ := r.NodeService.GetMessage(iter.String())
		statemsg, err2 := r.NodeService.StateSearchMsg(iter.MsgCid.String())
		if err2 != nil {
			fmt.Println(err2)
		} else {
			item.Height = int64(statemsg.Height)
		}
		item.Cid = iter.MsgCid.String()
		item.Version = int(iter.Msg.Version)
		item.From = iter.Msg.From.String()
		item.To = iter.Msg.To.String()
		item.Nonce = iter.Msg.Nonce
		item.Value = iter.Msg.Value.String()
		item.GasLimit = iter.Msg.GasLimit
		gasfeecap := iter.Msg.GasFeeCap.String()
		item.GasFeeCap = gasfeecap
		gaspremium := iter.Msg.GasPremium.String()
		item.GasPremium = gaspremium
		item.Method = uint64(iter.Msg.Method)
		item.MinerTip = iter.GasCost.MinerTip.String()
		item.BaseFeeBurn = iter.GasCost.BaseFeeBurn.String()
		item.OverEstimationBurn = iter.GasCost.OverEstimationBurn.String()
		item.Refund = iter.GasCost.Refund.String()
		item.MinerPenalty = iter.GasCost.MinerPenalty.String()
		item.MinerTip = iter.GasCost.MinerTip.String()

		obj, err := r.NodeService.StateDecodeParams(iter.Msg.To, iter.Msg.Method, iter.Msg.Params)

		if err == nil && obj != "" {
			item.Params = &obj
		}

		items = append(items, &item)
	}

	// sort the result by height
	sort.Slice(items, func(i, j int) bool {
		return items[i].Height < items[j].Height
	})

	return items, nil
}

func (r *queryResolver) MessageLowConfidence(ctx context.Context, cid string) (*model.MessageConfirmed, error) {
	var item model.MessageConfirmed

	statemsg, err := r.NodeService.StateSearchMsg(cid)
	if err != nil {
		return nil, err
	}

	if statemsg == nil {
		return nil, fmt.Errorf("not found")
	}

	item.Height = int64(statemsg.Height)
	item.Cid = statemsg.Message.String()

	iter, err := r.NodeService.StateReplay(ctx, cid)
	if err == nil {
		item.Version =  int(iter.Msg.Version)
		item.From =  iter.Msg.From.String()
		item.To = iter.Msg.To.String()
		item.Nonce = iter.Msg.Nonce
		item.Value = iter.Msg.Value.String()
		item.GasLimit = iter.Msg.GasLimit
		gasfeecap := iter.Msg.GasFeeCap.String()
		item.GasFeeCap = gasfeecap
		gaspremium := iter.Msg.GasPremium.String()
		item.GasPremium = gaspremium
		item.Method = uint64(iter.Msg.Method)
		item.GasUsed = iter.GasCost.GasUsed.Int64()
		_, item.GasBurned = util.ComputeGasOverestimationBurn(iter.GasCost.GasUsed.Int64(), iter.Msg.GasLimit)
		item.MinerTip = iter.GasCost.MinerTip.String()
		item.BaseFeeBurn = iter.GasCost.BaseFeeBurn.String()
		item.OverEstimationBurn = iter.GasCost.OverEstimationBurn.String()
		item.Refund = iter.GasCost.Refund.String()
		item.MinerPenalty = iter.GasCost.MinerPenalty.String()
		item.MinerTip = iter.GasCost.MinerTip.String()
	
		obj, err := r.NodeService.StateDecodeParams(iter.Msg.To, iter.Msg.Method, iter.Msg.Params)
	
		if err == nil && obj != "" {
			item.Params = &obj
		}
	}

	return &item, nil
}

func (r *subscriptionResolver) Messages(ctx context.Context) (<-chan []*model.Message, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *subscriptionResolver) ChainHead(ctx context.Context) (<-chan *model.ChainHead, error) {
	if r.ChainSubs == nil {

		r.ChainSubs = &Sub{
			Height: 0,
			Observers: map[uuid.UUID]struct {
				HeadChange chan *model.ChainHead
			}{},
		}

		go func() {
			var current int64

			for {
				for {
					log.Printf("subscribe to chainhead\n")
					chain, err := r.NodeService.ChainHeadSub(context.TODO())

					if err == nil {
						r.ChainSubs.Headchange = chain
						log.Printf("subscription success\n")
						break
					}

					log.Printf("...subscription failed: %s\n", err)
					time.Sleep(15 * time.Second)
				}

				for headchanges := range r.ChainSubs.Headchange {
					var res *model.ChainHead
					for _, elem := range headchanges {
						current = int64(elem.Val.Height())
						res = &model.ChainHead{Height: int64(elem.Val.Height())}
					}
					if current > r.ChainSubs.Height {
						r.ChainSubs.Height = current
						r.mu.Lock()
						for _, observer := range r.ChainSubs.Observers {
							observer.HeadChange <- res
						}
						r.mu.Unlock()
					}
				}
				log.Printf("subscription stalled\n")
			}
		}()
	}

	id := uuid.New()
	events := make(chan *model.ChainHead, 1)

	go func() {
		<-ctx.Done()
		r.mu.Lock()
		delete(r.ChainSubs.Observers, id)
		log.Printf("delete observer[%d]: %s\n", len(r.ChainSubs.Observers), id.String())
		r.mu.Unlock()
	}()

	r.mu.Lock()
	r.ChainSubs.Observers[id] = struct {
		HeadChange chan *model.ChainHead
	}{HeadChange: events}
	if r.ChainSubs.Height != 0 {
		events <- &model.ChainHead{Height: r.ChainSubs.Height}
	}
	log.Printf("add observer[%d]: %s\n", len(r.ChainSubs.Observers), id.String())
	r.mu.Unlock()

	return events, nil
}

func (r *subscriptionResolver) MpoolUpdate(ctx context.Context, address *string) (<-chan *model.MpoolUpdate, error) {
	if r.MpoolObserver == nil {

		r.MpoolObserver = &MpoolObserver{
			Observers: map[uuid.UUID]struct {
				address string
				update  chan *model.MpoolUpdate
			}{},
		}

		go func() {
			for {
				for {
					log.Printf("subscribe to mpoolsub\n")
					mpoolsub, err := r.NodeService.MpoolSub(context.TODO())

					if err == nil {
						r.MpoolObserver.channel = mpoolsub
						log.Printf("mpoolsub subscription success\n")
						break
					}

					log.Printf("...mpoolsub subscription failed: %s\n", err)
					time.Sleep(15 * time.Second)
				}

				for msg := range r.MpoolObserver.channel {
					var res model.MpoolUpdate

					res.Type = int(msg.Type)
					res.Message = &model.MessagePending{}
					res.Message.Cid = msg.Message.Cid().String()
					res.Message.Version = strconv.FormatUint(msg.Message.Message.Version, 10)
					fromaddr, _ := r.NodeService.AddressLookup(msg.Message.Message.From.String())
					res.Message.From = fromaddr
					toaddr, _ := r.NodeService.AddressLookup(msg.Message.Message.To.String())
					res.Message.To = toaddr
					nonce := strconv.FormatUint(msg.Message.Message.Nonce, 10)
					res.Message.Nonce = &nonce
					res.Message.Value = msg.Message.Message.Value.String()
					gaslimit := strconv.FormatInt(msg.Message.Message.GasLimit, 10)
					res.Message.GasLimit = &gaslimit
					gasfeecap := msg.Message.Message.GasFeeCap.String()
					res.Message.GasFeeCap = &gasfeecap
					gaspremium := msg.Message.Message.GasPremium.String()
					res.Message.GasPremium = &gaspremium
					res.Message.Method = msg.Message.Message.Method.String()

					obj, err := r.NodeService.StateDecodeParams(msg.Message.Message.To, msg.Message.Message.Method, msg.Message.Message.Params)

					if err == nil && obj != "" {
						res.Message.Params = &obj
					}

					r.mu.Lock()
					for _, observer := range r.MpoolObserver.Observers {
						if res.Message.From.Robust == observer.address || res.Message.To.Robust == observer.address ||
							res.Message.From.ID == observer.address || res.Message.To.ID == observer.address {
							//fmt.Printf("update: %s cid: %s\n", observer.address, res.Message.Cid)
							observer.update <- &res
						}
					}
					r.mu.Unlock()
				}
				log.Printf("mpoolsub subscription stalled\n")
			}
		}()
	}

	id := uuid.New()
	events := make(chan *model.MpoolUpdate, 1)

	go func() {
		<-ctx.Done()
		r.mu.Lock()
		delete(r.MpoolObserver.Observers, id)
		log.Printf("delete mpoolsub observer[%d]: %s\n", len(r.MpoolObserver.Observers), id.String())
		r.mu.Unlock()
	}()

	r.mu.Lock()
	r.MpoolObserver.Observers[id] = struct {
		address string
		update  chan *model.MpoolUpdate
	}{address: *address, update: events}
	//events <- &model.MpoolUpdate{Height: r.ChainSubs.Height}
	log.Printf("add mpoolsub observer[%d]: %s\n", len(r.MpoolObserver.Observers), id.String())
	r.mu.Unlock()

	return events, nil
}

// MessageConfirmed returns generated.MessageConfirmedResolver implementation.
func (r *Resolver) MessageConfirmed() generated.MessageConfirmedResolver {
	return &messageConfirmedResolver{r}
}

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

// Subscription returns generated.SubscriptionResolver implementation.
func (r *Resolver) Subscription() generated.SubscriptionResolver { return &subscriptionResolver{r} }

type messageConfirmedResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
type subscriptionResolver struct{ *Resolver }
