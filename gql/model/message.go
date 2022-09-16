package model

import (
	"strconv"

	"github.com/filecoin-project/lotus/chain/types"
)

type Message struct {
	Cid        string `json:"cid"`
	Version    uint64 `json:"version"`
	To         string `json:"to"`
	From       string `json:"from"`
	Nonce      uint64 `json:"nonce"`
	Value      string `json:"value"`
	GasLimit   int64  `json:"gasLimit"`
	GasFeeCap  string `json:"gasFeeCap"`
	GasPremium string `json:"gasPremium"`
	Method     uint64 `json:"method"`
	Height     uint64 `json:"height"`
	Params     string `json:"params"`
	// GasCost    *GasCost `json:"gasCost"`
}

// type GasCost struct {
// 	GasUsed            int64    `json:"gasUsed"`
// 	ParentBaseFee      string   `json:"parentBaseFee"`
// 	BaseFeeBurn        string   `json:"baseFeeBurn"`
// 	OverEstimationBurn string   `json:"overEstimationBurn"`
// 	MinerPenalty       string   `json:"minerPenalty"`
// 	MinerTip           string   `json:"minerTip"`
// 	Refund             string   `json:"refund"`
// 	GasRefund          int64    `json:"gasRefund"`
// 	GasBurned          int64    `json:"gasBurned"`
// }

type MessageConfirmed struct {
	Cid        string `json:"cid"`
	Height     int64  `json:"height"`
	StateRoot  string `json:"stateRoot"`
	Version    int    `json:"version"`
	From       string `json:"from"`
	To         string `json:"to"`
	Value      string `json:"value"`
	GasFeeCap  string `json:"gasFeeCap"`
	GasPremium string `json:"gasPremium"`
	GasLimit   int64  `json:"gasLimit"`
	SizeBytes  int    `json:"sizeBytes"`
	Nonce      uint64 `json:"nonce"`
	Method     uint64 `json:"method"`
	//	MethodName         string   `json:"methodName"`
	ActorName          string `json:"actorName"`
	ActorFamily        string `json:"actorFamily"`
	ExitCode           int64  `json:"exitCode"`
	GasUsed            int64  `json:"gasUsed"`
	ParentBaseFee      string `json:"parentBaseFee"`
	BaseFeeBurn        string `json:"baseFeeBurn"`
	OverEstimationBurn string `json:"overEstimationBurn"`
	MinerPenalty       string `json:"minerPenalty"`
	MinerTip           string `json:"minerTip"`
	Refund             string `json:"refund"`
	GasRefund          int64  `json:"gasRefund"`
	GasBurned          int64  `json:"gasBurned"`
	Params             string `json:"params"`
}

type MessagePending struct {
	Cid     string `json:"cid"`
	Version string `json:"version"`
	From    string `json:"from"`
	To      string `json:"to"`
	// To         *Address `json:"to"`
	// From       *Address `json:"from"`
	Nonce      *string `json:"nonce"`
	Value      string  `json:"value"`
	GasLimit   *string `json:"gasLimit"`
	GasFeeCap  *string `json:"gasFeeCap"`
	GasPremium *string `json:"gasPremium"`
	Method     uint64  `json:"method"`
	Height     string  `json:"height"`
	Params     string  `json:"params"`
}

func StrPtr(x string) *string {
	return &x
}

func CreatePendingMessage(item *types.Message) *MessagePending {
	message := &MessagePending{
		Cid:        item.Cid().String(),
		Version:    strconv.FormatUint(item.Version, 10),
		From:       item.From.String(),
		To:         item.To.String(),
		Nonce:      StrPtr(strconv.FormatUint(item.Nonce, 10)),
		Value:      item.Value.String(),
		GasLimit:   StrPtr(strconv.FormatInt(item.GasLimit, 10)),
		GasFeeCap:  StrPtr(item.GasFeeCap.String()),
		GasPremium: StrPtr(item.GasPremium.String()),
		Method:     uint64(item.Method),
	}

	return message
}

func CreateMessage(item *types.Message) *Message {
	message := &Message{
		Cid:        item.Cid().String(), // unsigned cid
		Version:    item.Version,
		From:       item.From.String(),
		To:         item.To.String(),
		Nonce:      item.Nonce,
		Value:      item.Value.String(),
		GasLimit:   item.GasLimit,
		GasFeeCap:  item.GasFeeCap.String(),
		GasPremium: item.GasPremium.String(),
		Method:     uint64(item.Method),
	}

	return message
}
