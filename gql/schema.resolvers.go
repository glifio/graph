package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"encoding/base64"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/filecoin-project/lily/model/derived"
	"github.com/filecoin-project/specs-actors/actors/builtin"
	"github.com/filecoin-project/specs-actors/actors/builtin/multisig"
	"github.com/filecoin-project/specs-actors/actors/runtime"
	"github.com/glifio/graph/gql/generated"
	"github.com/glifio/graph/gql/model"
	util "github.com/glifio/graph/internal/utils"
	"github.com/glifio/graph/pkg/lily"
	"github.com/google/uuid"
	"github.com/jinzhu/copier"
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
	var item model.Block
	copier.Copy(&item, &block)
	return &item, err
}

func (r *queryResolver) Message(ctx context.Context, cid *string) (*model.MessageConfirmed, error) {
	//msg, err := r.NodeService.GetMessage(*cid)
	msg, err := r.MessageConfirmedService.Get(*cid)
	var item model.MessageConfirmed
	copier.Copy(&item, &msg)
	item.Params = &msg.ParsedMessage.Params
	return &item, err
}

func (r *queryResolver) Messages(ctx context.Context, address *string, limit *int, offset *int) ([]*model.Message, error) {
	//return postgres.GetMessages(), nil
	var items []*model.Message
	var savedItems []lily.MessageItem
	savedItems, err := r.MessageService.List(*limit, *offset)
	if err != nil {
		return nil, err
	}
	for i, savedItem := range savedItems {
		var item model.Message
		savedItem = savedItems[i]
		item.Cid = savedItem.Cid
		item.Height = savedItem.Height
		item.From = savedItem.From
		item.To = savedItem.To
		item.Value = strconv.FormatFloat(savedItem.Value, 'f', -1, 64)
		item.Method = savedItem.Method
		item.Params = savedItem.Params
		items = append(items, &item)
	}
	return items, nil
}

func (r *queryResolver) PendingMessages(ctx context.Context, address *string, limit *int, offset *int) ([]*model.MessagePending, error) {
	var items []*model.MessagePending
	//pending, err := r.NodeService.GetPendingMessages(*address)
	pending, err := r.NodeService.GetPending()
	if err != nil {
		return nil, err
	}
	// for i, item := range pending {
	// 	var msg model.Message
	// 	msg.Cid = item[i].Cid.String()
	// 	m, err := r.NodeService.GetMessage(msg.Cid)
	// 	if err != nil {
	// 		msg.To = m.To.String()
	// 		msg.Method = m.Method.String()
	// 		*msg.GasFeeCap = m.GasFeeCap.String()
	// 		*msg.GasLimit = strconv.FormatInt(m.GasLimit, 64)
	// 	}
	// 	fmt.Println(item[i].Cid)
	// 	items = append(items, &msg)
	// }
	for _, item := range pending {
		var msg model.MessagePending
		msg.Cid = item.Cid().String()
		msg.Version = new(int)
		*msg.Version = int(item.Message.Version)
		msg.Method = item.Message.Method.String()
		msg.GasFeeCap = new(string)
		var gasfeecap = item.Message.GasFeeCap.String()
		msg.GasFeeCap = &gasfeecap

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
	return items, nil
}

func (r *queryResolver) MessagesConfirmed(ctx context.Context, address *string, limit *int, offset *int) ([]*model.MessageConfirmed, error) {
	var items []*model.MessageConfirmed
	var rs []derived.GasOutputs

	addr, err := r.NodeService.AddressLookup(*address)
	if err != nil {
		return nil, err
	}

	rs, err = r.MessageConfirmedService.Search(addr, limit, offset)
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

		if iter.Params != nil {
			// confirm the hashes match
			var rt runtime.Runtime
			rt.ValidateImmediateCallerType(builtin.CallerTypesSignable...)
			txn := &multisig.Transaction{To: iter.To, Value: iter.Value, Method: iter.Method, Params: iter.Params, Approved: iter.Approved}

			calculatedHash, _ := multisig.ComputeProposalHash(txn, rt.HashBlake2b)
			item.ProposalHash = base64.URLEncoding.EncodeToString(calculatedHash)
		}

		item.To = iter.To.String()
		item.Value = iter.Value.String()
		for _, appr := range iter.Approved {
			//a, _ := r.NodeService.AddressLookup(appr.String())
			item.Approved = append(item.Approved, appr.String())
		}
		//copier.Copy(&item, &r)
		items = append(items, &item)
	}
	return items, nil
}

func (r *queryResolver) StateListMessages(ctx context.Context, address string) ([]*model.MessageConfirmed, error) {
	var items []*model.MessageConfirmed

	pending, err := r.NodeService.StateListMessages(ctx, address)
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

		obj, err := r.NodeService.StateDecodeParams(iter.Msg.To, iter.Msg.Method, iter.Msg.Params)

		if err == nil && obj != "" {
			item.Params = &obj
		}

		items = append(items, &item)
	}
	return items, nil
}

func (r *queryResolver) MessageLowConfidence(ctx context.Context, cid string) (*model.MessageConfirmed, error) {
	var item model.MessageConfirmed

	iter, err := r.NodeService.StateReplay(ctx, cid)
	if err != nil {
		return nil, err
	}

	statemsg, err := r.NodeService.StateSearchMsg(cid)
	if err != nil {
		return nil, err
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

	obj, err := r.NodeService.StateDecodeParams(iter.Msg.To, iter.Msg.Method, iter.Msg.Params)

	if err == nil && obj != "" {
		item.Params = &obj
	}

	return &item, nil
}

func (r *subscriptionResolver) Messages(ctx context.Context) (<-chan []*model.Message, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *subscriptionResolver) ChainHead(ctx context.Context) (<-chan *model.ChainHead, error) {
	if r.ChainSubs == nil {
		chain, err := r.NodeService.ChainHeadSub(context.TODO())
		if err != nil {
			return nil, err
		}

		r.ChainSubs = &Sub{
			Headchange: chain,
			Height:     0,
			Observers: map[uuid.UUID]struct {
				HeadChange chan *model.ChainHead
			}{},
		}

		go func() {
			fmt.Printf("create chainhead listener\n")
			var current int64

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
			fmt.Printf("delete listen\n")
		}()
	}

	id := uuid.New()
	events := make(chan *model.ChainHead, 1)

	go func() {
		<-ctx.Done()
		r.mu.Lock()
		delete(r.ChainSubs.Observers, id)
		fmt.Printf("delete observer[%d]: %s\n", len(r.ChainSubs.Observers), id.String())
		r.mu.Unlock()
	}()

	r.mu.Lock()
	r.ChainSubs.Observers[id] = struct {
		HeadChange chan *model.ChainHead
	}{HeadChange: events}
	if r.ChainSubs.Height != 0 {
		events <- &model.ChainHead{Height: r.ChainSubs.Height}
	}
	fmt.Printf("add observer[%d]: %s\n", len(r.ChainSubs.Observers), id.String())
	r.mu.Unlock()

	return events, nil
}

func (r *subscriptionResolver) MpoolUpdate(ctx context.Context, address *string) (<-chan *model.MpoolUpdate, error) {
	if r.MpoolObserver == nil {
		mpoolsub, err := r.NodeService.MpoolSub(context.TODO())
		if err != nil {
			return nil, err
		}

		r.MpoolObserver = &MpoolObserver{
			channel: mpoolsub,
			Observers: map[uuid.UUID]struct {
				address string
				update  chan *model.MpoolUpdate
			}{},
		}

		go func() {
			fmt.Printf("create mpoolupdate listener\n")

			for msg := range r.MpoolObserver.channel {
				var res model.MpoolUpdate

				res.Type = (*int)(&msg.Type)
				res.Message = &model.Message{}
				res.Message.Cid = msg.Message.Cid().String()
				res.Message.Version = &msg.Message.Message.Version
				res.Message.From = msg.Message.Message.From.String()
				res.Message.To = msg.Message.Message.To.String()
				res.Message.Nonce = &msg.Message.Message.Nonce
				res.Message.Value = msg.Message.Message.Value.String()
				res.Message.GasLimit = &msg.Message.Message.GasLimit
				gasfeecap := msg.Message.Message.GasFeeCap.String()
				res.Message.GasFeeCap = &gasfeecap
				gaspremium := msg.Message.Message.GasPremium.String()
				res.Message.GasPremium = &gaspremium
				res.Message.Method = msg.Message.Message.Method.String()
				// if msg.Message.Message.Params != nil {
				// 	params := string(msg.Message.Message.Params)
				// 	res.Message.Params = &params
				// }

				r.mu.Lock()
				for _, observer := range r.MpoolObserver.Observers {
					fmt.Printf("address: %s\n", *address)
					if res.Message.From == observer.address || res.Message.To == observer.address {
						observer.update <- &res
					}
				}
				r.mu.Unlock()
			}
			fmt.Printf("delete listen\n")
		}()
	}

	id := uuid.New()
	events := make(chan *model.MpoolUpdate, 1)

	go func() {
		<-ctx.Done()
		r.mu.Lock()
		delete(r.MpoolObserver.Observers, id)
		fmt.Printf("delete observer[%d]: %s\n", len(r.MpoolObserver.Observers), id.String())
		r.mu.Unlock()
	}()

	r.mu.Lock()
	r.MpoolObserver.Observers[id] = struct {
		address string
		update  chan *model.MpoolUpdate
	}{address: *address, update: events}
	//events <- &model.MpoolUpdate{Height: r.ChainSubs.Height}
	fmt.Printf("add observer[%d]: %s\n", len(r.MpoolObserver.Observers), id.String())
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
