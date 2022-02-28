package lily

import (
	"github.com/filecoin-project/lily/model/derived"
	"github.com/glifio/graph/gql/model"
)

type GasOutputs struct {
	//lint:ignore U1000 tableName is a convention used by go-pg
	tableName          struct{} `pg:"derived_gas_outputs"`
	Height             int64    `pg:",pk,use_zero,notnull"`
	Cid                string   `pg:",pk,notnull"`
	StateRoot          string   `pg:",pk,notnull"`
	From               string   `pg:",notnull"`
	To                 string   `pg:",notnull"`
	Value              string   `pg:"type:numeric,notnull"`
	GasFeeCap          string   `pg:"type:numeric,notnull"`
	GasPremium         string   `pg:"type:numeric,notnull"`
	GasLimit           int64    `pg:",use_zero,notnull"`
	SizeBytes          int      `pg:",use_zero,notnull"`
	Nonce              uint64   `pg:",use_zero,notnull"`
	Method             uint64   `pg:",use_zero,notnull"`
	ActorName          string   `pg:",notnull"`
	ActorFamily        string   `pg:",notnull"`
	ExitCode           int64    `pg:",use_zero,notnull"`
	GasUsed            int64    `pg:",use_zero,notnull"`
	ParentBaseFee      string   `pg:"type:numeric,notnull"`
	BaseFeeBurn        string   `pg:"type:numeric,notnull"`
	OverEstimationBurn string   `pg:"type:numeric,notnull"`
	MinerPenalty       string   `pg:"type:numeric,notnull"`
	MinerTip           string   `pg:"type:numeric,notnull"`
	Refund             string   `pg:"type:numeric,notnull"`
	GasRefund          int64    `pg:",use_zero,notnull"`
	GasBurned          int64    `pg:",use_zero,notnull"`
	ParsedMessage      *ParsedMessage   `pg:"rel:has-one"`
}

type ParsedMessage struct {
	Height int64  `pg:",pk,notnull,use_zero"`
	Cid    string `pg:",pk,notnull"`
	From   string `pg:",notnull"`
	To     string `pg:",notnull"`
	Value  string `pg:"type:numeric,notnull"`
	Method string `pg:",notnull"`
	Params string `pg:",type:jsonb"`
}

type MessageConfirmedInterface interface {
	Get(id string, height *int) (*GasOutputs, error)
	List(address *string, limit *int, offset *int) ([]derived.GasOutputs, error)
	Search(address *model.Address, limit *int, offset *int) ([]derived.GasOutputs, error)
}
