package node

import (
	"context"
	"log"
	"net/http"
	"sync"

	"github.com/filecoin-project/go-jsonrpc"
	"github.com/filecoin-project/lotus/api"
)

var once sync.Once

type LotusNode struct {
	api    api.FullNodeStruct
	closer jsonrpc.ClientCloser
}

type LotusOptions struct {
	address string
}

// variabel Global
var lotus *LotusNode

func GetLotusInstance(opts *LotusOptions) *LotusNode {

	once.Do(func() {
		log.Printf("new lotus client: %s/n", opts.address)
		lotus = &LotusNode{}

		head := http.Header{}
		token := ""

		if token != "" {
			head.Set("Authorization", "Bearer "+token)
		}

		var err error
		closer, err := jsonrpc.NewMergeClient(context.Background(),
			opts.address,
			"Filecoin",
			api.GetInternalStructs(&lotus.api),
			head)
		if err != nil {
			log.Fatalf("connecting with lotus failed: %s", err)
		}

		lotus.closer = closer
	})

	return lotus
}

func (node *LotusNode) Close() {
	if node.closer != nil {
		node.closer()
	}
}
