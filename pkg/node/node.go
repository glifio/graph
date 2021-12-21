package node

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-jsonrpc"
	lotusapi "github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/ipfs/go-cid"
)

type NodeInterface interface {
	GetActor(id string) (*types.Actor, error)
	GetPendingMessages(id string) ([][]lotusapi.MessageCheckStatus, error)
	GetMessage(cidcc string) (*types.Message, error)
}

type Node struct {
	api lotusapi.FullNodeStruct
	closer jsonrpc.ClientCloser
}

func (t *Node) Connect(address string){
	authToken := "<value found in ~/.lotus/token>"
	headers := http.Header{"Authorization": []string{"Bearer " + authToken}}

	var err error
	t.closer, err = jsonrpc.NewMergeClient(context.Background(), address + "/rpc/v0", "Filecoin", []interface{}{&t.api.Internal, &t.api.CommonStruct.Internal}, headers)
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

	fmt.Println("msg: ", msg.GasFeeCap)

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