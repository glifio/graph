package graph

//go:generate go run github.com/99designs/gqlgen

import (
	"sync"

	"github.com/filecoin-project/lotus/api"
	"github.com/glifio/graph/gql/model"
	"github.com/glifio/graph/pkg/lily"
	"github.com/glifio/graph/pkg/node"
	"github.com/google/uuid"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Sub struct {
	Headchange <-chan []*api.HeadChange
	Height int64
	Observers map[uuid.UUID]struct {
		HeadChange  chan *model.ChainHead
	}
}

type Resolver struct{
	todos []*model.Todo
	NodeService node.NodeInterface
	MessageService lily.MessageInterface
	MessageConfirmedService lily.MessageConfirmedInterface
	BlockService lily.BlockInterface

	ChainSubs *Sub
	mu    sync.Mutex // nolint: structcheck
}