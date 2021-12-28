package lily

import "github.com/filecoin-project/lily/model/derived"

type MessageConfirmedItem struct {
	Cid    string  `json:"cid"`
	Height float64 `json:"height"`
	From   string  `json:"from"`
	To     string  `json:"to"`
	Value  float64 `json:"value"`
	Method string  `json:"method"`
	Params *string `json:"params"`
	GasFeeCap float64
	GasPremium float64
	GasLimit   string    
	SizeBytes      string
	Nonce          string
	StateRoot        string
	ExitCode       string
	GasUsed          string
	ParentBaseFee  float64
	BaseFeeBurn    float64
	OverEstimationBurn float64
	MinerPenalty  float64
	MinerTip     float64
	Refund       float64
	GasRefund        string 
	GasBurned         string
	ActorName     string
	ActorFamily       string
}

type MessageConfirmedInterface interface {
	Get(id string) (*derived.GasOutputs, error)
	List(address *string, limit *int, offset *int) ([]derived.GasOutputs, error)
}
