package graph

//go:generate go run github.com/99designs/gqlgen

import (
	"github.com/glifio/graph/gql/model"
	"github.com/glifio/graph/pkg/lily"
	"github.com/glifio/graph/pkg/node"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct{
	todos []*model.Todo
	NodeService node.NodeInterface
	MessageService lily.MessageInterface
	MessageConfirmedService lily.MessageConfirmedInterface
	BlockService lily.BlockInterface
}