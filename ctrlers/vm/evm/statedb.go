package evm

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/holiman/uint256"
	ctrlertypes "github.com/rigochain/rigo-go/ctrlers/types"
	types2 "github.com/rigochain/rigo-go/types"
	"github.com/rigochain/rigo-go/types/bytes"
	"github.com/rigochain/rigo-go/types/xerrors"
	tmlog "github.com/tendermint/tendermint/libs/log"
	"math/big"
	"sync"
)

type StateDBWrapper struct {
	*state.StateDB
	acctLedger ctrlertypes.IAccountHandler
	immutable  bool

	//txctx            *ctrlertypes.TrxContext
	accessedObjAddrs map[common.Address]int
	exec             bool

	logger tmlog.Logger
	mtx    sync.RWMutex
}

func NewStateDBWrapper(db ethdb.Database, lastRootHash []byte, acctHandler ctrlertypes.IAccountHandler, logger tmlog.Logger) (*StateDBWrapper, error) {
	var hash common.Hash
	copy(hash[:], lastRootHash)

	stateDB, err := state.New(hash, state.NewDatabase(db), nil)
	if err != nil {
		return nil, err
	}

	return &StateDBWrapper{
		StateDB:          stateDB,
		acctLedger:       acctHandler,
		accessedObjAddrs: make(map[common.Address]int),
		logger:           logger,
	}, nil
}

func (s *StateDBWrapper) Prepare(txhash bytes.HexBytes, txidx int, from, to types2.Address, exec bool) xerrors.XError {
	s.exec = exec
	s.StateDB.Prepare(txhash.Array32(), txidx)

	s.AddAddressToAccessList(from.Array20())
	if !types2.IsZeroAddress(to) {
		s.AddAddressToAccessList(to.Array20())
	}

	return nil
}

func (s *StateDBWrapper) ApplyTo() xerrors.XError {
	for addr, _ := range s.accessedObjAddrs {
		if s.Exist(addr) {
			amt := uint256.MustFromBig(s.StateDB.GetBalance(addr))
			nonce := s.StateDB.GetNonce(addr)

			acct := s.acctLedger.FindOrNewAccount(addr[:], s.exec)
			acct.SetBalance(amt)
			acct.SetNonce(nonce)

			_ = s.acctLedger.SetAccountCommittable(acct, s.exec)

			//s.logger.Debug("ApplyTo", "address", acct.Address, "nonce", acct.Nonce, "balance", acct.Balance.Dec())
		}
	}

	return nil
}

func (s *StateDBWrapper) Close() error {
	err := s.StateDB.Database().TrieDB().DiskDB().Close()
	s.StateDB = nil
	return err
}

func (s *StateDBWrapper) CreateAccount(addr common.Address) {
	//_ = s.acctLedger.FindOrNewAccount(addr[:], s.txctx.Exec)
	s.StateDB.CreateAccount(addr)
	s.logger.Debug("Create account", "address", addr)
}

func (s *StateDBWrapper) SubBalance(addr common.Address, amt *big.Int) {
	s.StateDB.SubBalance(addr, amt)
	//if acct := s.acctLedger.FindAccount(addr[:], s.exec); acct != nil {
	//	if err := acct.SubBalance(uint256.MustFromBig(amt)); err != nil {
	//		panic(err)
	//	}
	//	s.logger.Debug("SubBalance", "address", addr, "sub.amount", amt, "balance", acct.Balance.Dec(), "exec", s.exec)
	//	s.acctLedger.SetAccountCommittable(acct, s.exec)
	//}
}

func (s *StateDBWrapper) AddBalance(addr common.Address, amt *big.Int) {
	s.StateDB.AddBalance(addr, amt)
	//if acct := s.acctLedger.FindAccount(addr[:], s.exec); acct != nil {
	//	if err := acct.AddBalance(uint256.MustFromBig(amt)); err != nil {
	//		panic(err)
	//	}
	//	s.logger.Debug("AddBalance", "address", addr, "add.amount", amt, "balance", acct.Balance.Dec(), "exec", s.exec)
	//	s.acctLedger.SetAccountCommittable(acct, s.exec)
	//}
}

func (s *StateDBWrapper) GetBalance(addr common.Address) *big.Int {
	return s.StateDB.GetBalance(addr)
	//if acct := s.acctLedger.FindAccount(addr[:], s.exec); acct != nil {
	//	return acct.GetBalance().ToBig()
	//}
	//return big.NewInt(0)
}

func (s *StateDBWrapper) GetNonce(addr common.Address) uint64 {
	return s.StateDB.GetNonce(addr)
	//if acct := s.acctLedger.FindAccount(addr[:], s.exec); acct != nil {
	//	return acct.GetNonce()
	//}
	//return 0
}

func (s *StateDBWrapper) SetNonce(addr common.Address, n uint64) {
	s.StateDB.SetNonce(addr, n)
	//s.logger.Debug("SetNonce", "address", addr, "nonce", n)
	//if acct := s.acctLedger.FindAccount(addr[:], s.exec); acct != nil {
	//	acct.SetNonce(n)
	//	s.acctLedger.SetAccountCommittable(acct, s.exec)
	//}
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
	//ret := s.StateDB.Suicide(addr)
	//if ret {
	//	if acct := s.acctLedger.FindAccount(addr[:], s.exec); acct != nil {
	//		acct.Balance = uint256.NewInt(0)
	//		s.logger.Debug("Suicide", "address", addr, "balance", acct.Balance.Dec(), "exec", s.exec)
	//		s.acctLedger.SetAccountCommittable(acct, s.exec)
	//	}
	//}
	//return ret
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
	s.addAddressToAccessList(addr)
	if dest != nil {
		s.addAddressToAccessList(*dest)
	}
	for _, preaddr := range precompiles {
		s.addAddressToAccessList(preaddr)
	}
	for _, el := range txAccesses {
		s.addAddressToAccessList(el.Address)
		for _, key := range el.StorageKeys {
			s.AddSlotToAccessList(el.Address, key)
		}
	}

	s.StateDB.PrepareAccessList(addr, dest, precompiles, txAccesses)
}

func (s *StateDBWrapper) AddressInAccessList(addr common.Address) bool {
	return s.StateDB.AddressInAccessList(addr)
}

func (s *StateDBWrapper) SlotInAccessList(addr common.Address, slot common.Hash) (bool, bool) {
	return s.StateDB.SlotInAccessList(addr, slot)
}

func (s *StateDBWrapper) AddAddressToAccessList(addr common.Address) {
	s.addAddressToAccessList(addr)
	s.StateDB.AddAddressToAccessList(addr)
}

func (s *StateDBWrapper) addAddressToAccessList(addr common.Address) {
	if _, ok := s.accessedObjAddrs[addr]; !ok {
		stateObject := s.GetOrNewStateObject(addr)
		if stateObject != nil {
			rigoAcct := s.acctLedger.FindOrNewAccount(addr[:], s.exec)
			stateObject.SetNonce(rigoAcct.Nonce)
			stateObject.SetBalance(rigoAcct.Balance.ToBig())
			s.accessedObjAddrs[addr] = 1
		}
	}
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
