package evm

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/holiman/uint256"
	ctrlertypes "github.com/rigochain/rigo-go/ctrlers/types"
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

	logger tmlog.Logger
	mtx    sync.RWMutex
}

func NewStateDBWrapper(path string, lastRootHash []byte, acctHandler ctrlertypes.IAccountHandler, logger tmlog.Logger) (*StateDBWrapper, error) {
	//rawDB, err := rawdb.NewLevelDBDatabaseWithFreezer(path, 0, 0, path, "", false)
	rawDB, err := rawdb.NewLevelDBDatabase(path, 0, 0, "", false)
	if err != nil {
		return nil, err
	}

	var hash common.Hash
	copy(hash[:], lastRootHash)

	stateDB, err := state.New(hash, state.NewDatabase(rawDB), nil)
	if err != nil {
		_ = rawDB.Close()
		return nil, err
	}

	return &StateDBWrapper{
		StateDB:    stateDB,
		acctLedger: acctHandler,
		logger:     logger,
	}, nil
}

func (s *StateDBWrapper) ImmutableStateAt(n int64, hash []byte) (*StateDBWrapper, xerrors.XError) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	rootHash := bytes.HexBytes(hash).Array32()
	rawDB, _ := (s.StateDB.Database().TrieDB().DiskDB()).(ethdb.Database)
	stateDB, err := state.New(rootHash, state.NewDatabase(rawDB), nil)
	if err != nil {
		return nil, xerrors.From(err)
	}

	acctLedger, xerr := s.acctLedger.ImmutableAcctCtrlerAt(n)
	if xerr != nil {
		return nil, xerr
	}
	return &StateDBWrapper{
		StateDB:    stateDB,
		acctLedger: acctLedger,
		immutable:  true,
		logger:     s.logger,
	}, nil
}

func (s *StateDBWrapper) Close() error {
	err := s.StateDB.Database().TrieDB().DiskDB().Close()
	s.StateDB = nil
	return err
}

func (s *StateDBWrapper) CreateAccount(addr common.Address) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	_ = s.acctLedger.FindOrNewAccount(addr[:], true)
	s.StateDB.CreateAccount(addr)
	s.logger.Debug("Create account", "address", addr)
}

func (s *StateDBWrapper) SubBalance(addr common.Address, amt *big.Int) {
	s.logger.Debug("SubBalance", "address", addr, "amount", amt)
	if acct := s.acctLedger.FindAccount(addr[:], true); acct != nil {
		if err := acct.SubBalance(uint256.MustFromBig(amt)); err != nil {
			panic(err)
		}
		s.acctLedger.SetAccountCommittable(acct, true)
	}
}

func (s *StateDBWrapper) AddBalance(addr common.Address, amt *big.Int) {
	s.logger.Debug("AddBalance", "address", addr, "amount", amt)
	if acct := s.acctLedger.FindAccount(addr[:], true); acct != nil {
		if err := acct.AddBalance(uint256.MustFromBig(amt)); err != nil {
			panic(err)
		}
		s.acctLedger.SetAccountCommittable(acct, true)
	}
}

func (s *StateDBWrapper) GetBalance(addr common.Address) *big.Int {
	if acct := s.acctLedger.FindAccount(addr[:], true); acct != nil {
		return acct.GetBalance().ToBig()
	}
	return big.NewInt(0)
}

func (s *StateDBWrapper) GetNonce(addr common.Address) uint64 {
	if acct := s.acctLedger.FindAccount(addr[:], true); acct != nil {
		return acct.GetNonce()
	}
	return 0
}

func (s *StateDBWrapper) SetNonce(addr common.Address, n uint64) {
	s.logger.Debug("SetNonce", "address", addr, "nonce", n)
	if acct := s.acctLedger.FindAccount(addr[:], true); acct != nil {
		acct.SetNonce(n)
		s.acctLedger.SetAccountCommittable(acct, true)
	}
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
