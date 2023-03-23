package types

import (
	"github.com/rigochain/rigo-go/types"
	"github.com/rigochain/rigo-go/types/bytes"
	"github.com/rigochain/rigo-go/types/xerrors"
	tmtypes "github.com/tendermint/tendermint/types"
	"google.golang.org/protobuf/proto"
	"math/big"
	"time"
)

const (
	TRX_TRANSFER int32 = 1 + iota
	TRX_STAKING
	TRX_UNSTAKING
	TRX_PROPOSAL
	TRX_VOTING
	TRX_EXECUTE
)

const (
	EVENT_ATTR_TXTYPE   = "type"
	EVENT_ATTR_TXSENDER = "sender"
	EVENT_ATTR_TXRECVER = "receiver"
	EVENT_ATTR_ADDRPAIR = "addrpair"
)

type ITrxPayload interface {
	Type() int32
	Encode() ([]byte, xerrors.XError)
	Decode([]byte) xerrors.XError
}

type Trx struct {
	Version uint32         `json:"version,omitempty"`
	Time    int64          `json:"time"`
	Nonce   uint64         `json:"nonce"`
	From    types.Address  `json:"from"`
	To      types.Address  `json:"to"`
	Amount  *big.Int       `json:"amount"`
	Gas     *big.Int       `json:"gas"`
	Type    int32          `json:"type"`
	Payload ITrxPayload    `json:"payload,omitempty"`
	Sig     bytes.HexBytes `json:"sig"`
}

func NewTrx(ver uint32, from, to types.Address, nonce uint64, gas, amt *big.Int, payload ITrxPayload) *Trx {
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
	case TRX_EXECUTE:
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
	case TRX_EXECUTE:
		return xerrors.New("not supported payload type")
	default:
		return xerrors.New("unknown payload type")
	}

	tx.Version = txProto.Version
	tx.Time = txProto.Time
	tx.Nonce = txProto.Nonce
	tx.From = txProto.From
	tx.To = txProto.To
	tx.Amount = new(big.Int).SetBytes(txProto.XAmount)
	tx.Gas = new(big.Int).SetBytes(txProto.XGas)
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
