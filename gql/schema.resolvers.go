package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"
	"math/rand"
	"reflect"
	"strconv"
	"strings"

	"github.com/filecoin-project/lily/model/derived"
	"github.com/filecoin-project/specs-actors/actors/builtin"
	"github.com/glifio/graph/gql/generated"
	"github.com/glifio/graph/gql/model"
	"github.com/glifio/graph/pkg/lily"
	"github.com/jinzhu/copier"
)

func (r *messageResolver) To(ctx context.Context, obj *model.Message) (*model.Actor, error) {
	address := obj.To
	item, err := r.NodeService.GetActor(address)
	if err != nil {
		return nil, err
	} else {
		return &model.Actor{
			ID:   address,
			Code: item.Code.String(),
			Head: item.Head.String(),
			// StateRoot: item.StateRoot,
			// Nonce:     item.Nonce,
			// Height:    item.Height,
			Balance: item.Balance.String(),
		}, nil
	}

	//return &model.Actor{ID: obj.To, Code: "user1 " + obj.Cid}, nil
}

func (r *messageResolver) From(ctx context.Context, obj *model.Message) (*model.Actor, error) {
	address := obj.From
	item, err := r.NodeService.GetActor(address)
	if err != nil {
		return nil, err
	} else {
		return &model.Actor{
			ID:   address,
			Code: item.Code.String(),
			Head: item.Head.String(),
			// StateRoot: item.StateRoot,
			// Nonce:     item.Nonce,
			// Height:    item.Height,
			Balance: item.Balance.String(),
		}, nil
	}
}

func (r *messageConfirmedResolver) From(ctx context.Context, obj *model.MessageConfirmed) (*model.Address, error) {
	addr, err := r.NodeService.AddressLookup(obj.From)
	return addr, err
}

func (r *messageConfirmedResolver) To(ctx context.Context, obj *model.MessageConfirmed) (*model.Address, error) {
	addr, err := r.NodeService.AddressLookup(obj.To)
	return addr, err
}

func (r *messageConfirmedResolver) MethodName(ctx context.Context, obj *model.MessageConfirmed) (string, error) {
	switch(strings.Split(obj.ActorName, "/")[2]){
		case "account":
			return reflect.ValueOf(&builtin.MethodsAccount).Elem().Type().Field(int(obj.Method)).Name, nil;
		case "init":
			return reflect.ValueOf(&builtin.MethodsInit).Elem().Type().Field(int(obj.Method)).Name, nil;
		case "reward":
			return reflect.ValueOf(&builtin.MethodsReward).Elem().Type().Field(int(obj.Method)).Name, nil;
		case "multisig":
			return reflect.ValueOf(&builtin.MethodsMultisig).Elem().Type().Field(int(obj.Method)).Name, nil;
		case "paymentchannel":
			return reflect.ValueOf(&builtin.MethodsPaych).Elem().Type().Field(int(obj.Method)).Name, nil;
		case "storagemarket":
			return reflect.ValueOf(&builtin.MethodsMarket).Elem().Type().Field(int(obj.Method)).Name, nil;
		case "storageminer":
			return reflect.ValueOf(&builtin.MethodsMiner).Elem().Type().Field(int(obj.Method)).Name, nil;
		case "storagepower":
			return reflect.ValueOf(&builtin.MethodsPower).Elem().Type().Field(int(obj.Method)).Name, nil;
		default:
			return "???", nil;
	}
}

func (r *messageConfirmedResolver) Block(ctx context.Context, obj *model.MessageConfirmed) (*model.Block, error) {
	block, err := r.BlockService.GetByMessage(obj.Height, obj.Cid)
	var item model.Block
	copier.Copy(&item, &block)
	return &item, err
}

func (r *mutationResolver) CreateTodo(ctx context.Context, input model.NewTodo) (*model.Todo, error) {
	todo := &model.Todo{
		Text:   input.Text,
		ID:     fmt.Sprintf("T%d", rand.Int()),
		UserID: input.UserID,
	}
	r.todos = append(r.todos, todo)
	return todo, nil
}

func (r *queryResolver) Todos(ctx context.Context) ([]*model.Todo, error) {
	panic(fmt.Errorf("not implemented"))
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
		item.Value = savedItem.Value
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
		// var gaslimit = strconv.FormatInt(item.Message.GasLimit, 64)
		// msg.GasLimit = &gaslimit
		items = append(items, &msg)
	}
	return items, nil
}

func (r *queryResolver) MessagesConfirmed(ctx context.Context, address *string, limit *int, offset *int) ([]*model.MessageConfirmed, error) {
	var items []*model.MessageConfirmed
	var rs []derived.GasOutputs
	rs, err := r.MessageConfirmedService.List(address, limit, offset)
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

func (r *subscriptionResolver) Messages(ctx context.Context) (<-chan []*model.Message, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *todoResolver) User(ctx context.Context, obj *model.Todo) (*model.User, error) {
	return &model.User{ID: obj.UserID, Name: "user " + obj.UserID}, nil
}

func (r *todoResolver) Actor(ctx context.Context, obj *model.Todo) (*model.Actor, error) {
	panic(fmt.Errorf("not implemented"))
}

// Message returns generated.MessageResolver implementation.
func (r *Resolver) Message() generated.MessageResolver { return &messageResolver{r} }

// MessageConfirmed returns generated.MessageConfirmedResolver implementation.
func (r *Resolver) MessageConfirmed() generated.MessageConfirmedResolver {
	return &messageConfirmedResolver{r}
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

// Subscription returns generated.SubscriptionResolver implementation.
func (r *Resolver) Subscription() generated.SubscriptionResolver { return &subscriptionResolver{r} }

// Todo returns generated.TodoResolver implementation.
func (r *Resolver) Todo() generated.TodoResolver { return &todoResolver{r} }

type messageResolver struct{ *Resolver }
type messageConfirmedResolver struct{ *Resolver }
type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
type subscriptionResolver struct{ *Resolver }
type todoResolver struct{ *Resolver }
