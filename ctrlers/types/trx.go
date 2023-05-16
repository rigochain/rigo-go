package types

import (
	"fmt"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/holiman/uint256"
	"github.com/rigochain/rigo-go/types"
	"github.com/rigochain/rigo-go/types/bytes"
	"github.com/rigochain/rigo-go/types/crypto"
	"github.com/rigochain/rigo-go/types/xerrors"
	tmtypes "github.com/tendermint/tendermint/types"
	"google.golang.org/protobuf/proto"
	"io"
	"time"
)

const (
	TRX_TRANSFER int32 = 1 + iota
	TRX_STAKING
	TRX_UNSTAKING
	TRX_PROPOSAL
	TRX_VOTING
	TRX_CONTRACT
)

const (
	EVENT_ATTR_TXTYPE   = "type"
	EVENT_ATTR_TXSENDER = "sender"
	EVENT_ATTR_TXRECVER = "receiver"
	EVENT_ATTR_ADDRPAIR = "addrpair"
)

type trxRPL struct {
	Version uint32
	Time    uint64
	Nonce   uint64
	From    types.Address
	To      types.Address
	Amount  string
	Gas     string
	Type    uint32
	Payload bytes.HexBytes
	Sig     bytes.HexBytes
}

type ITrxPayload interface {
	Type() int32
	Equal(ITrxPayload) bool
	Encode() ([]byte, xerrors.XError)
	Decode([]byte) xerrors.XError
	RLPEncode() ([]byte, error)
	RLPDecode([]byte) error
}

type Trx struct {
	Version uint32         `json:"version,omitempty"`
	Time    int64          `json:"time"`
	Nonce   uint64         `json:"nonce"`
	From    types.Address  `json:"from"`
	To      types.Address  `json:"to"`
	Amount  *uint256.Int   `json:"amount"`
	Gas     *uint256.Int   `json:"gas"`
	Type    int32          `json:"type"`
	Payload ITrxPayload    `json:"payload,omitempty"`
	Sig     bytes.HexBytes `json:"sig"`
}

func (tx *Trx) Equal(_tx *Trx) bool {
	if tx.Version != _tx.Version {
		return false
	}
	if tx.Time != _tx.Time {
		return false
	}
	if tx.Nonce != _tx.Nonce {
		return false
	}
	if tx.From.Compare(_tx.From) != 0 {
		return false
	}
	if tx.To.Compare(_tx.To) != 0 {
		return false
	}
	if tx.Amount.Cmp(_tx.Amount) != 0 {
		return false
	}
	if tx.Gas.Cmp(_tx.Gas) != 0 {
		return false
	}
	if tx.Type != _tx.Type {
		return false
	}
	if tx.Version != _tx.Version {
		return false
	}
	if tx.Version != _tx.Version {
		return false
	}
	if bytes.Compare(tx.Sig, _tx.Sig) != 0 {
		return false
	}
	if tx.Payload != nil {
		return tx.Payload.Equal(_tx.Payload)
	} else if _tx.Payload != nil {
		return false
	}
	return true
}

func (tx *Trx) EncodeRLP(w io.Writer) error {
	var payload bytes.HexBytes
	if tx.Payload != nil {
		_tmp, err := tx.Payload.RLPEncode()
		if err != nil {
			return err
		}
		payload = _tmp
	}
	tmpTx := &trxRPL{
		Version: tx.Version,
		Time:    uint64(tx.Time),
		Nonce:   tx.Nonce,
		From:    tx.From,
		To:      tx.To,
		Amount:  tx.Amount.Hex(),
		Gas:     tx.Gas.Hex(),
		Type:    uint32(tx.Type),
		Payload: payload,
		Sig:     tx.Sig,
	}
	return rlp.Encode(w, tmpTx)
}

func (tx *Trx) DecodeRLP(s *rlp.Stream) error {
	rtx := &trxRPL{}
	if err := s.Decode(rtx); err != nil {
		return err
	}

	var payload ITrxPayload
	if rtx.Payload != nil && len(rtx.Payload) > 0 {
		switch int32(rtx.Type) {
		case TRX_TRANSFER:
			payload = &TrxPayloadAssetTransfer{}
		case TRX_STAKING:
			payload = &TrxPayloadStaking{}
		case TRX_UNSTAKING:
			payload = &TrxPayloadUnstaking{}
		case TRX_PROPOSAL:
			payload = &TrxPayloadProposal{}
		case TRX_VOTING:
			payload = &TrxPayloadVoting{}
		case TRX_CONTRACT:
			payload = &TrxPayloadContract{}
		default:
			return xerrors.ErrInvalidTrxPayloadType
		}
		if err := payload.RLPDecode(rtx.Payload); err != nil {
			return err
		}
	}

	tx.Version = rtx.Version
	tx.Time = int64(rtx.Time)
	tx.Nonce = rtx.Nonce
	tx.From = rtx.From
	tx.To = rtx.To
	tx.Amount = uint256.MustFromHex(rtx.Amount)
	tx.Gas = uint256.MustFromHex(rtx.Gas)
	tx.Type = int32(rtx.Type)
	tx.Payload = payload
	tx.Sig = rtx.Sig
	return nil
}

var _ rlp.Encoder = (*Trx)(nil)
var _ rlp.Decoder = (*Trx)(nil)

func NewTrx(ver uint32, from, to types.Address, nonce uint64, gas, amt *uint256.Int, payload ITrxPayload) *Trx {
	return &Trx{
		Version: ver,
		Time:    time.Now().Round(0).UTC().UnixNano(),
		Nonce:   nonce,
		From:    from,
		To:      to,
		Amount:  amt,
		Gas:     gas,
		Type:    payload.Type(),
		Payload: payload,
	}
}

func (tx *Trx) GetType() int32 {
	return tx.Type
}

func (tx *Trx) TypeString() string {
	switch tx.GetType() {
	case TRX_TRANSFER:
		return "transfer"
	case TRX_STAKING:
		return "staking"
	case TRX_UNSTAKING:
		return "unstaking"
	case TRX_PROPOSAL:
		return "proposal"
	case TRX_VOTING:
		return "voting"
	case TRX_CONTRACT:
		return "execution"
	}
	return ""
}

func (tx *Trx) Decode(bz []byte) xerrors.XError {
	pm := TrxProto{}
	if err := proto.Unmarshal(bz, &pm); err != nil {
		return xerrors.From(err)
	} else if err := tx.fromProto(&pm); err != nil {
		return err
	}
	return nil
}

func (tx *Trx) Encode() ([]byte, xerrors.XError) {
	if pm, err := tx.toProto(); err != nil {
		return nil, xerrors.From(err)
	} else if bz, err := proto.Marshal(pm); err != nil {
		return nil, xerrors.From(err)
	} else {
		return bz, nil
	}
}

func (tx *Trx) fromProto(txProto *TrxProto) xerrors.XError {
	var payload ITrxPayload
	switch txProto.Type {
	case TRX_TRANSFER, TRX_STAKING:
		// there is no payload!!!
	case TRX_UNSTAKING:
		payload = &TrxPayloadUnstaking{}
		if err := payload.Decode(txProto.XPayload); err != nil {
			return err
		}
	case TRX_PROPOSAL:
		payload = &TrxPayloadProposal{}
		if err := payload.Decode(txProto.XPayload); err != nil {
			return err
		}
	case TRX_VOTING:
		payload = &TrxPayloadVoting{}
		if err := payload.Decode(txProto.XPayload); err != nil {
			return err
		}
	case TRX_CONTRACT:
		payload = &TrxPayloadContract{}
		if err := payload.Decode(txProto.XPayload); err != nil {
			return err
		}
	default:
		return xerrors.ErrInvalidTrxPayloadType
	}

	tx.Version = txProto.Version
	tx.Time = txProto.Time
	tx.Nonce = txProto.Nonce
	tx.From = txProto.From
	tx.To = txProto.To
	tx.Amount = new(uint256.Int).SetBytes(txProto.XAmount)
	tx.Gas = new(uint256.Int).SetBytes(txProto.XGas)
	tx.Type = txProto.Type
	tx.Payload = payload
	tx.Sig = txProto.Sig
	return nil
}

func (tx *Trx) toProto() (*TrxProto, xerrors.XError) {
	var payload []byte
	if tx.Payload != nil {
		if bz, err := tx.Payload.Encode(); err != nil {
			return nil, err
		} else {
			payload = bz
		}
	}
	return &TrxProto{
		Version:  tx.Version,
		Time:     tx.Time,
		Nonce:    tx.Nonce,
		From:     tx.From,
		To:       tx.To,
		XAmount:  tx.Amount.Bytes(),
		XGas:     tx.Gas.Bytes(),
		Type:     tx.Type,
		XPayload: payload,
		Sig:      tx.Sig,
	}, nil
}

func (tx *Trx) Hash() ([]byte, error) {
	if tx.Sig != nil {
		oriSig := tx.Sig
		tx.Sig = nil
		defer func() { tx.Sig = oriSig }()
	}
	bz, err := tx.Encode()
	if err != nil {
		return nil, err
	}

	return tmtypes.Tx(bz).Hash(), nil
}

func VerifyTrx(tx *Trx) xerrors.XError {
	sig := tx.Sig
	tx.Sig = nil
	_txbz, xerr := tx.Encode()
	tx.Sig = sig
	if xerr != nil {
		return xerr
	}

	fromAddr, _, xerr := crypto.Sig2Addr(_txbz, sig)
	if xerr != nil {
		return xerr
	}
	if bytes.Compare(fromAddr, tx.From) != 0 {
		return xerrors.ErrInvalidTrxSig.Wrap(fmt.Errorf("wrong address or sig - expected: %v, actual: %v", tx.From, fromAddr))
	}
	return nil
}

func VerifyTrxRLP(tx *Trx) xerrors.XError {
	sig := tx.Sig
	tx.Sig = nil
	_txbz, err := rlp.EncodeToBytes(tx)
	tx.Sig = sig
	if err != nil {
		return xerrors.From(err)
	}
	fromAddr, _, xerr := crypto.Sig2Addr(_txbz, sig)
	if xerr != nil {
		return xerr
	}
	if bytes.Compare(fromAddr, tx.From) != 0 {
		return xerrors.ErrInvalidTrxSig.Wrap(fmt.Errorf("wrong address or sig - expected: %v, actual: %v", tx.From, fromAddr))
	}
	return nil
}
