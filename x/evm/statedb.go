package evm

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	ctrlertypes "github.com/rigochain/rigo-go/ctrlers/types"
	"math/big"
)

var (
	lastRoot common.Hash
)

type StateDBWrapper struct {
	*state.StateDB
	acctHelper ctrlertypes.IAccountHelper
}

func NewStateDBWrapper(path string, acctHelper ctrlertypes.IAccountHelper) (*StateDBWrapper, error) {
	//rawDB, err := rawdb.NewLevelDBDatabaseWithFreezer(path, 0, 0, path, "", false)
	rawDB, err := rawdb.NewLevelDBDatabase(path, 0, 0, "", false)
	if err != nil {
		return nil, err
	}

	stateDB, err := state.New(lastRoot, state.NewDatabase(rawDB), nil)
	if err != nil {
		_ = rawDB.Close()
		return nil, err
	}

	return &StateDBWrapper{stateDB, acctHelper}, nil
}

func (s *StateDBWrapper) CreateAccount(addr common.Address) {
	s.StateDB.CreateAccount(addr)
}

func (s *StateDBWrapper) SubBalance(addr common.Address, amt *big.Int) {
	s.StateDB.SubBalance(addr, amt)
}

func (s *StateDBWrapper) AddBalance(addr common.Address, amt *big.Int) {
	s.StateDB.AddBalance(addr, amt)
}

func (s *StateDBWrapper) GetBalance(addr common.Address) *big.Int {
	return s.StateDB.GetBalance(addr)
}

func (s *StateDBWrapper) GetNonce(addr common.Address) uint64 {
	return s.StateDB.GetNonce(addr)
}

func (s *StateDBWrapper) SetNonce(addr common.Address, n uint64) {
	s.StateDB.SetNonce(addr, n)
}

func (s *StateDBWrapper) GetCodeHash(addr common.Address) common.Hash {
	return s.StateDB.GetCodeHash(addr)
}

func (s *StateDBWrapper) GetCode(addr common.Address) []byte {
	return s.StateDB.GetCode(addr)
}

func (s *StateDBWrapper) SetCode(addr common.Address, code []byte) {
	s.StateDB.SetCode(addr, code)
}

func (s *StateDBWrapper) GetCodeSize(addr common.Address) int {
	return s.StateDB.GetCodeSize(addr)
}

func (s *StateDBWrapper) AddRefund(gas uint64) {
	s.StateDB.AddRefund(gas)
}

func (s *StateDBWrapper) SubRefund(gas uint64) {
	s.StateDB.SubRefund(gas)
}

func (s *StateDBWrapper) GetRefund() uint64 {
	return s.StateDB.GetRefund()
}

func (s *StateDBWrapper) GetCommittedState(addr common.Address, hash common.Hash) common.Hash {
	return s.StateDB.GetCommittedState(addr, hash)
}

func (s *StateDBWrapper) GetState(addr common.Address, hash common.Hash) common.Hash {
	return s.StateDB.GetState(addr, hash)
}

func (s *StateDBWrapper) SetState(addr common.Address, key, value common.Hash) {
	s.StateDB.SetState(addr, key, value)
}

func (s *StateDBWrapper) Suicide(addr common.Address) bool {
	return s.StateDB.Suicide(addr)
}

func (s *StateDBWrapper) HasSuicided(addr common.Address) bool {
	return s.StateDB.HasSuicided(addr)
}

func (s *StateDBWrapper) Exist(addr common.Address) bool {
	return s.StateDB.Exist(addr)
}

func (s *StateDBWrapper) Empty(addr common.Address) bool {
	return s.StateDB.Empty(addr)
}

func (s *StateDBWrapper) PrepareAccessList(addr common.Address, dest *common.Address, precompiles []common.Address, txAccesses types.AccessList) {
	s.StateDB.PrepareAccessList(addr, dest, precompiles, txAccesses)
}

func (s *StateDBWrapper) AddressInAccessList(addr common.Address) bool {
	return s.StateDB.AddressInAccessList(addr)
}

func (s *StateDBWrapper) SlotInAccessList(addr common.Address, slot common.Hash) (bool, bool) {
	return s.StateDB.SlotInAccessList(addr, slot)
}

func (s *StateDBWrapper) AddAddressToAccessList(addr common.Address) {
	s.StateDB.AddAddressToAccessList(addr)
}

func (s *StateDBWrapper) AddSlotToAccessList(addr common.Address, slot common.Hash) {
	s.StateDB.AddSlotToAccessList(addr, slot)
}

func (s *StateDBWrapper) RevertToSnapshot(revid int) {
	s.StateDB.RevertToSnapshot(revid)
}

func (s *StateDBWrapper) Snapshot() int {
	return s.StateDB.Snapshot()
}

func (s *StateDBWrapper) AddLog(log *types.Log) {
	s.StateDB.AddLog(log)
}

func (s *StateDBWrapper) AddPreimage(hash common.Hash, preimage []byte) {
	s.StateDB.AddPreimage(hash, preimage)
}

func (s *StateDBWrapper) ForEachStorage(addr common.Address, cb func(common.Hash, common.Hash) bool) error {
	return s.StateDB.ForEachStorage(addr, cb)
}

var _ vm.StateDB = (*StateDBWrapper)(nil)
