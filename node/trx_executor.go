package node

import (
	"fmt"
	"github.com/holiman/uint256"
	ctrlertypes "github.com/rigochain/rigo-go/ctrlers/types"
	rtypes "github.com/rigochain/rigo-go/types"
	"github.com/rigochain/rigo-go/types/xerrors"
	"github.com/tendermint/tendermint/libs/log"
	"math"
	"runtime"
	"strconv"
	"strings"
)

type TrxExecutor struct {
	txCtxChs []chan *ctrlertypes.TrxContext
	logger   log.Logger
}

func NewTrxExecutor(n int, logger log.Logger) *TrxExecutor {
	txCtxChs := make([]chan *ctrlertypes.TrxContext, n)
	for i := 0; i < n; i++ {
		txCtxChs[i] = make(chan *ctrlertypes.TrxContext, 5000)
	}
	return &TrxExecutor{
		txCtxChs: txCtxChs,
		logger:   logger,
	}
}

func (txe *TrxExecutor) Start() {
	for i, ch := range txe.txCtxChs {
		go executionRoutine(fmt.Sprintf("executionRoutine-%d", i), ch, txe.logger)
	}
}

func (txe *TrxExecutor) Stop() {
	for _, ch := range txe.txCtxChs {
		close(ch)
	}
	txe.txCtxChs = nil
}

func (txe *TrxExecutor) ExecuteSync(ctx *ctrlertypes.TrxContext) xerrors.XError {
	xerr := validateTrx(ctx)
	if xerr != nil {
		return xerr
	}
	xerr = runTrx(ctx)
	if xerr != nil {
		return xerr
	}
	return nil
}

func (txe *TrxExecutor) ExecuteAsync(ctx *ctrlertypes.TrxContext) xerrors.XError {
	n := len(txe.txCtxChs)
	i := int(ctx.Tx.From[0]) % n

	if txe.txCtxChs[i] == nil {
		return xerrors.NewOrdinary("transaction execution channel is not available")
	}
	//if ctx.Exec {
	//	txe.logger.Info("[DEBUG] TrxExecutor::ExecuteAsync", "index", i, "txhash", ctx.TxHash)
	//}
	txe.txCtxChs[i] <- ctx

	return nil
}

// for test
func goid() int {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	idField := strings.Fields(strings.TrimPrefix(string(buf[:n]), "goroutine "))[0]
	id, err := strconv.Atoi(idField)
	if err != nil {
		panic(fmt.Sprintf("cannot get goroutine id: %v", err))
	}
	return id
}

func executionRoutine(name string, ch chan *ctrlertypes.TrxContext, logger log.Logger) {
	logger.Info("Start transaction execution routine", "goid", goid(), "name", name)

	for ctx := range ch {
		//if ctx.Exec {
		//	logger.Info("[DEBUG] Begin of executionRoutine", "txhash", ctx.TxHash, "goid", goid(), "name", name)
		//}
		var xerr xerrors.XError

		if xerr = validateTrx(ctx); xerr == nil {
			xerr = runTrx(ctx)
		}

		//if ctx.Exec {
		//	logger.Info("[DEBUG] End of executionRoutine", "txhash", ctx.TxHash, "goid", goid(), "name", name)
		//}

		ctx.Callback(ctx, xerr)
	}
}

func commonValidation0(ctx *ctrlertypes.TrxContext) xerrors.XError {
	//
	// the following CAN be parellely done
	//
	tx := ctx.Tx

	if len(tx.From) != rtypes.AddrSize {
		return xerrors.ErrInvalidAddress
	}
	if len(tx.To) != rtypes.AddrSize {
		return xerrors.ErrInvalidAddress
	}
	if tx.Amount.Sign() < 0 {
		return xerrors.ErrInvalidAmount
	}
	if tx.Gas < 0 || tx.Gas > math.MaxInt64 {
		return xerrors.ErrInvalidGas
	}
	if tx.GasPrice.Sign() < 0 || tx.GasPrice.Cmp(ctx.GovHandler.GasPrice()) != 0 {
		return xerrors.ErrInvalidGasPrice
	}

	feeAmt := new(uint256.Int).Mul(tx.GasPrice, uint256.NewInt(tx.Gas))
	if feeAmt.Cmp(ctx.GovHandler.MinTrxFee()) < 0 {
		return xerrors.ErrInvalidGas.Wrapf("too small gas(fee)")
	}

	if ctx.Exec {
		_, pubKeyBytes, xerr := ctrlertypes.VerifyTrxRLP(tx, ctx.ChainID)
		if xerr != nil {
			return xerr
		}
		ctx.SenderPubKey = pubKeyBytes
	}
	return nil
}

func commonValidation1(ctx *ctrlertypes.TrxContext) xerrors.XError {

	//
	// this validation MUST be serially done
	//
	tx := ctx.Tx

	feeAmt := new(uint256.Int).Mul(tx.GasPrice, uint256.NewInt(tx.Gas))
	needAmt := new(uint256.Int).Add(feeAmt, tx.Amount)
	if xerr := ctx.Sender.CheckBalance(needAmt); xerr != nil {
		return xerr
	}
	if xerr := ctx.Sender.CheckNonce(tx.Nonce); xerr != nil {
		return xerr.Wrap(fmt.Errorf("invalid nonce - ledger: %v, tx:%v, address: %v, txhash: %X", ctx.Sender.GetNonce(), tx.Nonce, ctx.Sender.Address, ctx.TxHash))
	}
	return nil
}

func validateTrx(ctx *ctrlertypes.TrxContext) xerrors.XError {

	//
	// tx validation
	if xerr := commonValidation0(ctx); xerr != nil {
		return xerr
	}
	if xerr := commonValidation1(ctx); xerr != nil {
		return xerr
	}

	switch ctx.Tx.GetType() {
	case ctrlertypes.TRX_PROPOSAL, ctrlertypes.TRX_VOTING:
		if xerr := ctx.TrxGovHandler.ValidateTrx(ctx); xerr != nil {
			return xerr
		}
	case ctrlertypes.TRX_TRANSFER, ctrlertypes.TRX_SETDOC:
		if xerr := ctx.TrxAcctHandler.ValidateTrx(ctx); xerr != nil {
			return xerr
		}
	case ctrlertypes.TRX_STAKING, ctrlertypes.TRX_UNSTAKING, ctrlertypes.TRX_WITHDRAW:
		if xerr := ctx.TrxStakeHandler.ValidateTrx(ctx); xerr != nil {
			return xerr
		}
	case ctrlertypes.TRX_CONTRACT:
		if xerr := ctx.TrxEVMHandler.ValidateTrx(ctx); xerr != nil {
			return xerr
		}
	default:
		return xerrors.ErrUnknownTrxType
	}

	return nil
}

func runTrx(ctx *ctrlertypes.TrxContext) xerrors.XError {

	//
	// tx execution
	switch ctx.Tx.GetType() {
	case ctrlertypes.TRX_CONTRACT:
		if xerr := ctx.TrxEVMHandler.ExecuteTrx(ctx); xerr != nil && xerr != xerrors.ErrUnknownTrxType {
			return xerr
		}
	case ctrlertypes.TRX_PROPOSAL, ctrlertypes.TRX_VOTING:
		if xerr := ctx.TrxGovHandler.ExecuteTrx(ctx); xerr != nil {
			return xerr
		}
	case ctrlertypes.TRX_TRANSFER, ctrlertypes.TRX_SETDOC:
		if xerr := ctx.TrxAcctHandler.ExecuteTrx(ctx); xerr != nil {
			// todo: rollback changes in TrxGovHandler.ExecuteTrx
			return xerr
		}
	case ctrlertypes.TRX_STAKING, ctrlertypes.TRX_UNSTAKING, ctrlertypes.TRX_WITHDRAW:
		if xerr := ctx.TrxStakeHandler.ExecuteTrx(ctx); xerr != nil {
			// todo: rollback changes in TrxGovHandler.ExecuteTrx and TrxAcctHandler.ExecuteTrx
			return xerr
		}
	default:
		return xerrors.ErrUnknownTrxType
	}

	if xerr := postRunTrx(ctx); xerr != nil {
		return xerr
	}

	return nil
}

func postRunTrx(ctx *ctrlertypes.TrxContext) xerrors.XError {

	if ctx.Exec &&
		ctx.Tx.GetType() == ctrlertypes.TRX_CONTRACT &&
		ctx.Tx.To.Compare(rtypes.ZeroAddress()) == 0 {
		// this tx is to deploy contract
		var contractAddr rtypes.Address
		for _, evt := range ctx.Events {
			if evt.GetType() == "evm" {
				for _, attr := range evt.Attributes {
					if string(attr.Key) == "contractAddress" {
						if caddr, err := rtypes.HexToAddress(string(attr.Value)); err != nil {
							return xerrors.From(err)
						} else {
							contractAddr = caddr
							break
						}
					}
				}
				if contractAddr != nil {
					break
				}
			}
		}
		if contractAddr == nil {
			return xerrors.NewOrdinary("there is no contract address")
		}

		acct := ctx.AcctHandler.FindAccount(contractAddr, ctx.Exec)
		if acct == nil {
			return xerrors.ErrNotFoundAccount.Wrapf("contract address: %v", contractAddr)
		}

		// mark the new account as contract account
		acct.SetCode([]byte("contract"))

		if xerr := ctx.AcctHandler.SetAccountCommittable(acct, ctx.Exec); xerr != nil {
			return xerr
		}
	}

	if ctx.Tx.GetType() != ctrlertypes.TRX_CONTRACT {
		//
		// The gas & nonce is already processed in `EVMCtrler` if the tx type is `TRX_CONTRACT`.

		// processing fee = gas * gasPrice
		fee := new(uint256.Int).Mul(ctx.Tx.GasPrice, uint256.NewInt(uint64(ctx.Tx.Gas)))
		if xerr := ctx.Sender.SubBalance(fee); xerr != nil {
			return xerr
		}

		// processing nonce
		ctx.Sender.AddNonce()

		if xerr := ctx.AcctHandler.SetAccountCommittable(ctx.Sender, ctx.Exec); xerr != nil {
			return xerr
		}

		// set used gas
		ctx.GasUsed = ctx.Tx.Gas
	}
	return nil
}
