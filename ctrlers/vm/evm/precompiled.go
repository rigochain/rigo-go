package evm

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
	rigo_crypto "github.com/rigochain/rigo-go/types/crypto"
	"math/big"
)

func init() {
	vm.PrecompiledContractsHomestead[common.BytesToAddress([]byte{1})] = &rigo_ecrecover{}
	vm.PrecompiledContractsByzantium[common.BytesToAddress([]byte{1})] = &rigo_ecrecover{}
	vm.PrecompiledContractsIstanbul[common.BytesToAddress([]byte{1})] = &rigo_ecrecover{}
	vm.PrecompiledContractsBerlin[common.BytesToAddress([]byte{1})] = &rigo_ecrecover{}
}

// ECRECOVER implemented as a native contract.
type rigo_ecrecover struct{}

func (c *rigo_ecrecover) RequiredGas(input []byte) uint64 {
	return params.EcrecoverGas
}

func (c *rigo_ecrecover) Run(input []byte) ([]byte, error) {
	const ecRecoverInputLength = 128

	input = common.RightPadBytes(input, ecRecoverInputLength)
	// "input" is (hash, v, r, s), each 32 bytes
	// but for ecrecover we want (r, s, v)

	r := new(big.Int).SetBytes(input[64:96])
	s := new(big.Int).SetBytes(input[96:128])
	v := input[63] // - 27

	// tighter sig s values input homestead only apply to tx sigs
	if !allZero(input[32:63]) || !crypto.ValidateSignatureValues(v, r, s, false) {
		return nil, nil
	}
	// We must make sure not to modify the 'input', so placing the 'v' along with
	// the signature needs to be done on a new allocation
	sig := make([]byte, 65)
	copy(sig, input[64:128])
	sig[64] = v
	// v needs to be at the end for libsecp256k1
	publicKey, err := crypto.SigToPub(input[:32], sig)
	if err != nil {
		return nil, err
	}

	return rigo_crypto.Pub2Addr(publicKey), nil
}

func allZero(b []byte) bool {
	for _, byte := range b {
		if byte != 0 {
			return false
		}
	}
	return true
}
