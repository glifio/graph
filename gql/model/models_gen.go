// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package model

import (
	"fmt"
	"io"
	"strconv"
)

type Actor struct {
	ID        string `json:"id"`
	Code      string `json:"Code"`
	Head      string `json:"Head"`
	Nonce     string `json:"Nonce"`
	Balance   string `json:"Balance"`
	StateRoot string `json:"StateRoot"`
	Height    int64  `json:"Height"`
}

type Address struct {
	ID     string `json:"id"`
	Robust string `json:"robust"`
}

type Block struct {
	Cid             string   `json:"cid"`
	Height          int64    `json:"height"`
	Miner           string   `json:"miner"`
	Parents         []string `json:"parents"`
	ParentWeight    string   `json:"parentWeight"`
	ParentBaseFee   string   `json:"parentBaseFee"`
	ParentStateRoot string   `json:"parentStateRoot"`
	WinCount        *int64   `json:"winCount"`
	Messages        string   `json:"messages"`
	Timestamp       uint64   `json:"timestamp"`
	ForkSignaling   *uint64  `json:"forkSignaling"`
}

type ChainHead struct {
	Height int64 `json:"height"`
}

type ExecutionTrace struct {
	ExecutionTrace string `json:"executionTrace"`
}

type GasCost struct {
	GasUsed            int64  `json:"gasUsed"`
	BaseFeeBurn        string `json:"baseFeeBurn"`
	OverEstimationBurn string `json:"overEstimationBurn"`
	MinerPenalty       string `json:"minerPenalty"`
	MinerTip           string `json:"minerTip"`
	Refund             string `json:"refund"`
	TotalCost          string `json:"totalCost"`
}

type InvocResult struct {
	GasCost        *GasCost        `json:"gasCost"`
	Receipt        *MessageReceipt `json:"receipt"`
	ExecutionTrace *ExecutionTrace `json:"executionTrace"`
}

type MessageReceipt struct {
	ExitCode int64  `json:"exitCode"`
	Return   string `json:"return"`
	GasUsed  int64  `json:"gasUsed"`
}

type MpoolUpdate struct {
	Type    int             `json:"type"`
	Message *MessagePending `json:"message"`
}

type MsigTransaction struct {
	ID           int64      `json:"id"`
	To           *Address   `json:"to"`
	Value        string     `json:"value"`
	Method       uint64     `json:"method"`
	Params       string     `json:"params"`
	Approved     []*Address `json:"approved"`
	ProposalHash string     `json:"proposalHash"`
}

type QueryMessage struct {
	Messages []*Message `json:"messages"`
}

type Status struct {
	Height   uint64 `json:"height"`
	Estimate int64  `json:"estimate"`
}

type TipSet struct {
	Cids         []string `json:"cids"`
	Blks         []*Block `json:"blks"`
	Height       uint64   `json:"height"`
	Key          string   `json:"key"`
	MinTimestamp uint64   `json:"minTimestamp"`
}

type FilUnit string

const (
	FilUnitFil      FilUnit = "Fil"
	FilUnitAttoFil  FilUnit = "AttoFil"
	FilUnitFemtoFil FilUnit = "FemtoFil"
	FilUnitPicoFil  FilUnit = "PicoFil"
	FilUnitNanoFil  FilUnit = "NanoFil"
)

var AllFilUnit = []FilUnit{
	FilUnitFil,
	FilUnitAttoFil,
	FilUnitFemtoFil,
	FilUnitPicoFil,
	FilUnitNanoFil,
}

func (e FilUnit) IsValid() bool {
	switch e {
	case FilUnitFil, FilUnitAttoFil, FilUnitFemtoFil, FilUnitPicoFil, FilUnitNanoFil:
		return true
	}
	return false
}

func (e FilUnit) String() string {
	return string(e)
}

func (e *FilUnit) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = FilUnit(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid FilUnit", str)
	}
	return nil
}

func (e FilUnit) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}
