package util

import (
	"github.com/filecoin-project/go-state-types/big"
	"github.com/glifio/graph/gql/model"
)

func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func AddressCompareFromTo(_address string, _from *model.Address, _to *model.Address) (bool){
	return AddressCompare(_address, _from) || AddressCompare(_address, _to);
}

func AddressCompare(_p1 string, _p2 *model.Address) (bool){
	return (len(_p2.Robust) > 1 && _p2.Robust[1:] == _p1[1:]) ||
	 (len(_p2.ID) > 1 && _p2.ID[1:] == _p1[1:]);
}

const (
	gasOveruseNum   = 11
	gasOveruseDenom = 10
)

// ComputeGasOverestimationBurn computes amount of gas to be refunded and amount of gas to be burned
// Result is (refund, burn)
func ComputeGasOverestimationBurn(gasUsed, gasLimit int64) (int64, int64) {
	if gasUsed == 0 {
		return 0, gasLimit
	}

	// over = gasLimit/gasUsed - 1 - 0.1
	// over = min(over, 1)
	// gasToBurn = (gasLimit - gasUsed) * over

	// so to factor out division from `over`
	// over*gasUsed = min(gasLimit - (11*gasUsed)/10, gasUsed)
	// gasToBurn = ((gasLimit - gasUsed)*over*gasUsed) / gasUsed
	over := gasLimit - (gasOveruseNum*gasUsed)/gasOveruseDenom
	if over < 0 {
		return gasLimit - gasUsed, 0
	}

	// if we want sharper scaling it goes here:
	// over *= 2

	if over > gasUsed {
		over = gasUsed
	}

	// needs bigint, as it overflows in pathological case gasLimit > 2^32 gasUsed = gasLimit / 2
	gasToBurn := big.NewInt(gasLimit - gasUsed)
	gasToBurn = big.Mul(gasToBurn, big.NewInt(over))
	gasToBurn = big.Div(gasToBurn, big.NewInt(gasUsed))

	return gasLimit - gasUsed - gasToBurn.Int64(), gasToBurn.Int64()
}
