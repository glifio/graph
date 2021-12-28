package model

type Message struct {
	Cid        string  `json:"cid"`
	Version    *int    `json:"version"`
	From   	   string  `json:"from"`
	To         string  `json:"to"`
	Nonce      *string `json:"Nonce"`
	Value      float64 `json:"value"`
	GasLimit   *string `json:"GasLimit"`
	GasFeeCap  *string `json:"GasFeeCap"`
	GasPremium *string `json:"GasPremium"`
	Method     string  `json:"method"`
	Height     float64 `json:"height"`
	Params     *string `json:"params"`
}

type MessageConfirmed struct {
	Cid                string   `json:"cid"`
	Height             int64    `json:"height"`
	StateRoot          string   `json:"stateRoot"`
	Version            int      `json:"version"`
	From               string   `json:"from"`
	To                 string   `json:"to"`
	Value              string   `json:"value"`
	GasFeeCap          string   `json:"gasFeeCap"`
	GasPremium         string   `json:"gasPremium"`
	GasLimit           int64    `json:"gasLimit"`
	SizeBytes          int      `json:"sizeBytes"`
	Nonce              uint64   `json:"nonce"`
	Method             uint64   `json:"method"`
	MethodName         string   `json:"methodName"`
	ActorName          string   `json:"actorName"`
	ActorFamily        string   `json:"actorFamily"`
	ExitCode           int64    `json:"exitCode"`
	GasUsed            int64    `json:"gasUsed"`
	ParentBaseFee      string   `json:"parentBaseFee"`
	BaseFeeBurn        string   `json:"baseFeeBurn"`
	OverEstimationBurn string   `json:"overEstimationBurn"`
	MinerPenalty       string   `json:"minerPenalty"`
	MinerTip           string   `json:"minerTip"`
	Refund             string   `json:"refund"`
	GasRefund          int64    `json:"gasRefund"`
	GasBurned          int64    `json:"gasBurned"`
}