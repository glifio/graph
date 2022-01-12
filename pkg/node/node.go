package node

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-jsonrpc"
	"github.com/filecoin-project/lotus/api"
	lotusapi "github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/api/v1api"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/glifio/graph/gql/model"
	"github.com/ipfs/go-cid"
)

type NodeInterface interface {
	GetActor(id string) (*types.Actor, error)
	GetPendingMessages(id string) ([][]lotusapi.MessageCheckStatus, error)
	GetPending() ([]*types.SignedMessage, error)
	GetMessage(cidcc string) (*types.Message, error)
	AddressLookup(id string) (*model.Address, error)
	MsigGetPending(addr string) ([]*lotusapi.MsigTransaction, error)
	ChainHeadSub(ctx context.Context) (<-chan []*lotusapi.HeadChange, error)
	MpoolSub(ctx context.Context) (<-chan lotusapi.MpoolUpdate, error)
}

type Node struct {
	//api1 lotusapi.FullNodeStruct
	closer jsonrpc.ClientCloser
	api v1api.FullNodeStruct
}

func (t *Node) Connect(address string, token string){
	head := http.Header{}

	if token != "" {
		head.Set("Authorization","Bearer " + token)
	}

	var err error
	t.closer, err = jsonrpc.NewMergeClient(context.Background(), 
		address, 
		"Filecoin", 
		api.GetInternalStructs(&t.api), 
		head)
	if err != nil {
		log.Fatalf("connecting with lotus failed: %s", err)
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
		log.Fatal(err)
	}
	
	msg, err := t.api.StateSearchMsg(context.Background(), types.EmptyTSK, c, 0, true )
	if err != nil {
		log.Fatal(err)
	}

	return msg, nil
}

func (t *Node) GetPendingMessages(id string) ([][]lotusapi.MessageCheckStatus, error) {
	addr, err := address.NewFromString(id)
	if err != nil {
		log.Fatal(err)
	}
	
	status, err := t.api.MpoolCheckPendingMessages(context.Background(), addr)
	//status, err := t.api.MpoolCheckMessages(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	return status, nil
}

func (t *Node) MsigGetPending(addr string) ([]*lotusapi.MsigTransaction, error) {
	res, err := address.NewFromString(addr)
	if err != nil {
		return nil, err
	}
	
	pending, err := t.api.MsigGetPending(context.Background(), res, types.EmptyTSK)
	if err != nil {
		fmt.Println("get")
	}
	return pending, err
}

func (t *Node) ChainHead(ctx context.Context) (*types.TipSet, error) {
	tipset, err := t.api.ChainHead(ctx)
	return tipset, err
}

func (t *Node) ChainHeadSub(ctx context.Context) (<-chan []*lotusapi.HeadChange, error) {
	headchange, err := t.api.ChainNotify(ctx)
	if err != nil {
		log.Fatal(err)
	}
	return headchange, err
}

func (t *Node) MpoolSub(ctx context.Context) (<-chan lotusapi.MpoolUpdate, error) {
	mpool, err := t.api.MpoolSub(ctx)
	if err != nil {
		log.Fatal(err)
	}
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
	result := &model.Address{}
	addr, err := address.NewFromString(id)
	if err != nil {
		log.Fatal(err)
	}
	var rs address.Address
	switch(addr.Protocol()){
		case address.ID:
			//protocol = ID
			result.ID = addr.String()
			rs, err = t.api.StateAccountKey(context.Background(), addr, types.EmptyTSK)
			if(err==nil){
				result.Robust = rs.String()
			}
		default:
			result.Robust = addr.String()
			rs, err = t.api.StateLookupID(context.Background(), addr, types.EmptyTSK)
			if(err==nil){
				result.ID = rs.String()
			}
	}
	return result, nil
}