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
