package state

import (
	"errors"
	"fmt"
	"github.com/kysee/arcanus/cmd/version"
	"github.com/kysee/arcanus/ctrlers/account"
	"github.com/kysee/arcanus/ctrlers/gov"
	"github.com/kysee/arcanus/ctrlers/stake"
	"github.com/kysee/arcanus/genesis"
	"github.com/kysee/arcanus/libs"
	"github.com/kysee/arcanus/libs/crypto"
	"github.com/kysee/arcanus/types"
	"github.com/kysee/arcanus/types/trxs"
	"github.com/kysee/arcanus/types/xerrors"
	abcitypes "github.com/tendermint/tendermint/abci/types"
	tmcrypto "github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/libs/json"
	"github.com/tendermint/tendermint/libs/log"
	tmver "github.com/tendermint/tendermint/version"
	"math/big"
	"sync"
)

var _ abcitypes.Application = (*ChainCtrler)(nil)

type ChainCtrler struct {
	abcitypes.BaseApplication

	currBlockGasInfo *BlockGasInfo
	stateDB          *StateDB
	acctCtrler       *account.AccountCtrler
	stakeCtrler      *stake.StakeCtrler
	govCtrler        *gov.GovCtrler
	trxExecutor      *TrxExecutor

	logger log.Logger
	mtx    sync.Mutex
}

func NewChainCtrler(dbDir string, logger log.Logger) *ChainCtrler {
	stateDB, err := openStateDB("chain_state", dbDir)
	if err != nil {
		panic(err)
	}

	govCtrler, err := gov.NewGovCtrler(dbDir, logger)
	if err != nil {
		panic(err)
	}

	acctCtrler, err := account.NewAccountCtrler(dbDir, logger)
	if err != nil {
		panic(err)
	}

	stakeCtrler, err := stake.NewStakeCtrler(dbDir, logger)
	if err != nil {
		panic(err)
	}

	return &ChainCtrler{
		stateDB:     stateDB,
		acctCtrler:  acctCtrler,
		stakeCtrler: stakeCtrler,
		govCtrler:   govCtrler,
		trxExecutor: NewTrxExecutor(logger, acctCtrler, stakeCtrler),
		logger:      logger,
	}
}

func (ctrler *ChainCtrler) Close() error {
	if err := ctrler.acctCtrler.Close(); err != nil {
		return err
	}
	if err := ctrler.stateDB.Close(); err != nil {
		return err
	}
	return nil
}

func (ctrler *ChainCtrler) Info(info abcitypes.RequestInfo) abcitypes.ResponseInfo {
	ctrler.logger.Info("Info", "version", tmver.ABCIVersion, "AppVersion", version.String())

	lastBlockHeight := ctrler.stateDB.LastBlockHeight()
	if lastBlockHeight > 0 {
		err := ctrler.govCtrler.ImportRules(func() []byte {
			_acct := ctrler.acctCtrler.FindAccount(libs.ZeroBytes(tmcrypto.AddressSize), false)
			if _acct == nil {
				panic(errors.New("the account of governance rules is not found"))
			} else if len(_acct.GetCode()) == 0 {
				panic(errors.New("the account of governance rules has no code"))
			}
			return _acct.GetCode()
		})
		if err != nil {
			panic(err)
		}
	}

	return abcitypes.ResponseInfo{
		Data:             "",
		Version:          tmver.ABCIVersion,
		AppVersion:       version.Uint64(),
		LastBlockHeight:  lastBlockHeight,
		LastBlockAppHash: ctrler.stateDB.LastBlockAppHash(),
	}
}

// InitChain is called only when the ResponseInfo::LastBlockHeight which is returned in Info() is 0.
func (ctrler *ChainCtrler) InitChain(chain abcitypes.RequestInitChain) abcitypes.ResponseInitChain {
	appState := genesis.GenesisAppState{}
	if err := json.Unmarshal(chain.AppStateBytes, &appState); err != nil {
		panic(err)
	}

	// todo: check whether 'appHash' is equal to the original hash of the current blockchain network.
	// but how to get the original hash? official web site????
	appHash, err := appState.Hash()
	if err != nil {
		panic(err)
	}

	if err = ctrler.govCtrler.ImportRules(func() []byte {
		amtPower, _ := new(big.Int).SetString(appState.GovRules.AmountPerPower, 10)
		rwdPower, _ := new(big.Int).SetString(appState.GovRules.RewardPerPower, 10)
		govRules := &gov.GovRules{
			Version:        appState.GovRules.Version,
			AmountPerPower: amtPower,
			RewardPerPower: rwdPower,
		}

		if govRulesCode, err := govRules.Encode(); err != nil {
			panic(err)
		} else {
			// create account for gov rules and save it
			govRulesAddr := libs.ZeroBytes(tmcrypto.AddressSize)
			govRulesAcct := ctrler.acctCtrler.FindOrNewAccount(govRulesAddr, true)
			govRulesAcct.SetCode(govRulesCode) // will be saved at commit
			return govRulesCode
		}
	}); err != nil {
		panic(err)
	}

	for _, validator := range chain.Validators {
		pubKey := validator.GetPubKey()
		pubKeyBytes := pubKey.GetSecp256K1()
		addr, err := crypto.PubBytes2Addr(pubKeyBytes)
		if err != nil {
			panic(err)
		}
		staker := ctrler.stakeCtrler.AddStakerWith(addr, pubKeyBytes)
		stake0 := stake.NewStakeWithPower(staker.Owner, staker.Owner, validator.Power, 0, libs.ZeroBytes(32), ctrler.govCtrler.GetRules())
		staker.AppendStake(stake0)
	}

	for _, holder := range appState.AssetHolders {
		// In acctCtrler.realAccounts, create accounts of the genesis holders.
		// When the ctrler.Commit() is called by the consensus engine,
		// these will be saved and acctCtrler.realAccounts will become acctCtrler.simAccounts.
		acct := ctrler.acctCtrler.FindOrNewAccount(holder.Address, true)
		if acct == nil {
			panic(fmt.Errorf("fail to create an account of the genesis holder(%X)", holder.Address))
		}
		amt, ok := new(big.Int).SetString(holder.Balance, 10)
		if !ok {
			panic(fmt.Errorf("the genesis holder(%X)'s balance(%v) maybe wrong", holder.Address, holder.Balance))
		}
		if err := acct.AddBalance(amt); err != nil {
			panic(fmt.Errorf("failt to set balance(%v) of the account(%x)", amt, acct.GetAddress()))
		}
		ctrler.logger.Info("InitChain", "holder address", acct.GetAddress(), "amount", acct.GetBalance())
	}

	// these values will be saved as state of the consensus engine.
	return abcitypes.ResponseInitChain{
		AppHash: appHash,
	}
}

func (ctrler *ChainCtrler) CheckTx(req abcitypes.RequestCheckTx) abcitypes.ResponseCheckTx {
	response := abcitypes.ResponseCheckTx{Code: abcitypes.CodeTypeOK}

	ctrler.mtx.Lock()
	defer ctrler.mtx.Unlock()

	switch req.Type {
	case abcitypes.CheckTxType_New:
		if txctx, err := trxs.NewTrxContext(req.Tx,
			ctrler.stateDB.LastBlockHeight()+int64(1),
			false,
			ctrler.acctCtrler,
			ctrler.govCtrler.GetRules()); err != nil {

			xerr, ok := err.(xerrors.XError)
			if !ok {
				xerr = xerrors.ErrCheckTx.Wrap(err)
			}
			response.Code = xerr.Code()
			response.Log = xerr.Error()
			//response.Code = xerrors.ErrCodeDeliverTx
			//response.Log = xerrors.ErrCheckTx.With(err).Error()
		} else if err := ctrler.trxExecutor.Execute(txctx); err != nil {
			xerr, ok := err.(xerrors.XError)
			if !ok {
				xerr = xerrors.ErrCheckTx.Wrap(err)
			}
			response.Code = xerr.Code()
			response.Log = xerr.Error()
			//response.Code = xerrors.ErrCodeDeliverTx
			//response.Log = xerrors.ErrCheckTx.With(err).Error()
			response.GasWanted = txctx.Tx.Gas.Int64()
			response.GasUsed = txctx.GasUsed.Int64()
		} else {
			response.GasWanted = txctx.Tx.Gas.Int64()
			response.GasUsed = txctx.GasUsed.Int64()
		}
	case abcitypes.CheckTxType_Recheck:
		// do nothing
	}
	return response
}

func (ctrler *ChainCtrler) BeginBlock(req abcitypes.RequestBeginBlock) abcitypes.ResponseBeginBlock {
	if req.Header.Height != ctrler.stateDB.LastBlockHeight()+1 {
		panic(fmt.Errorf("error block height: expected(%v), actural(%v)", ctrler.stateDB.LastBlockHeight()+1, req.Header.Height))
	}

	// save the block fee info. - it will be used for rewarding
	ctrler.currBlockGasInfo = &BlockGasInfo{
		Height: req.Header.Height,
		Gas:    big.NewInt(0),
		Owner:  req.Header.ProposerAddress,
	}

	// todo: implement processing for the evidences (req.ByzantineValidators)

	return abcitypes.ResponseBeginBlock{}
}

func (ctrler *ChainCtrler) DeliverTx(req abcitypes.RequestDeliverTx) abcitypes.ResponseDeliverTx {
	response := abcitypes.ResponseDeliverTx{}

	if txctx, err := trxs.NewTrxContext(req.Tx,
		ctrler.stateDB.LastBlockHeight()+int64(1),
		true,
		ctrler.acctCtrler,
		ctrler.govCtrler.GetRules()); err != nil {

		xerr, ok := err.(xerrors.XError)
		if !ok {
			xerr = xerrors.ErrDeliverTx.Wrap(err)
		}
		response.Code = xerr.Code()
		response.Log = xerr.Error()
		//response.Code = xerrors.ErrCodeDeliverTx
		//response.Log = xerrors.ErrCheckTx.With(err).Error()
	} else if err := ctrler.trxExecutor.Execute(txctx); err != nil {
		xerr, ok := err.(xerrors.XError)
		if !ok {
			xerr = xerrors.ErrDeliverTx.Wrap(err)
		}
		response.Code = xerr.Code()
		response.Log = xerr.Error()
		//response.Code = xerrors.ErrCodeDeliverTx
		//response.Log = xerrors.ErrCheckTx.With(err).Error()
		response.GasWanted = txctx.Tx.Gas.Int64()
		response.GasUsed = txctx.GasUsed.Int64()
	} else {
		ctrler.currBlockGasInfo.Gas = new(big.Int).Add(ctrler.currBlockGasInfo.Gas, txctx.GasUsed)
		response.GasWanted = txctx.Tx.Gas.Int64()
		response.GasUsed = txctx.GasUsed.Int64()
		response.Events = []abcitypes.Event{
			{
				Type: "tx",
				Attributes: []abcitypes.EventAttribute{
					{Key: []byte(trxs.EVENT_ATTR_TXTYPE), Value: []byte(txctx.Tx.TypeString()), Index: true},
					{Key: []byte(trxs.EVENT_ATTR_TXSENDER), Value: []byte(txctx.Tx.From.String()), Index: true},
					{Key: []byte(trxs.EVENT_ATTR_TXRECVER), Value: []byte(txctx.Tx.To.String()), Index: true},
					{Key: []byte(trxs.EVENT_ATTR_ADDRPAIR), Value: []byte(txctx.Tx.From.String() + txctx.Tx.To.String()), Index: true},
				},
			},
		}
	}

	return response
}

func (ctrler *ChainCtrler) EndBlock(req abcitypes.RequestEndBlock) abcitypes.ResponseEndBlock {
	lastBlockGasInfo := ctrler.stateDB.LastBlockGasInfo()
	if lastBlockGasInfo != nil {
		if req.Height != lastBlockGasInfo.Height+1 {
			panic(fmt.Errorf("error block height: expected(%v), actural(%v)", lastBlockGasInfo.Height+1, req.Height))
		}
		ctrler.stakeCtrler.ApplyReward(lastBlockGasInfo.Owner, lastBlockGasInfo.Gas)
	}

	updatedValidators := ctrler.stakeCtrler.UpdateValidators(int(ctrler.govCtrler.GetRules().GetMaxValidatorCount()))

	return abcitypes.ResponseEndBlock{
		ValidatorUpdates: updatedValidators,
	}
}

func (ctrler *ChainCtrler) Commit() abcitypes.ResponseCommit {

	ctrler.mtx.Lock()
	defer ctrler.mtx.Unlock()

	appHash0, ver0, err := ctrler.acctCtrler.Commit()
	if err != nil {
		panic(err)
	}

	ctrler.logger.Debug("account hash", types.HexBytes(appHash0), "version", ver0)

	appHash1, ver1, err := ctrler.stakeCtrler.Commit()
	if err != nil {
		panic(err)
	}
	ctrler.logger.Debug("stakes hash", types.HexBytes(appHash0), "version", ver1)

	if ver0 != ver1 {
		panic(fmt.Sprintf("Not same versions: account:%v, stake:%v", ver0, ver1))
	}

	appHash := crypto.DefaultHash(append(appHash0, appHash1...))

	ctrler.stateDB.PutLastBlockHeight(ver0)
	ctrler.stateDB.PutLastBlockAppHash(appHash[:])
	ctrler.stateDB.PutLastBlockGasInfo(ctrler.currBlockGasInfo)
	ctrler.currBlockGasInfo = nil

	return abcitypes.ResponseCommit{
		Data: appHash[:],
	}
}
