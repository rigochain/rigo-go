package types

import (
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/rigochain/rigo-go/types/xerrors"
	"io"
)

type TrxPayloadAssetTransfer struct{}

func (tx *TrxPayloadAssetTransfer) Type() int32 {
	return TRX_TRANSFER
}

func (tx *TrxPayloadAssetTransfer) Equal(_tx ITrxPayload) bool {
	return true
}

func (tx *TrxPayloadAssetTransfer) Decode(bz []byte) xerrors.XError {
	return nil
}

func (tx *TrxPayloadAssetTransfer) Encode() ([]byte, xerrors.XError) {
	return nil, nil
}

func (tx *TrxPayloadAssetTransfer) RLPEncode() ([]byte, error) {
	return nil, nil //rlp.EncodeToBytes(tx)
}
func (tx *TrxPayloadAssetTransfer) RLPDecode(bz []byte) error {
	return nil //rlp.DecodeBytes(bz, tx)
}

func (tx *TrxPayloadAssetTransfer) EncodeRLP(w io.Writer) error {
	return nil
}

func (tx *TrxPayloadAssetTransfer) DecodeRLP(s *rlp.Stream) error {
	return nil
}

var _ ITrxPayload = (*TrxPayloadAssetTransfer)(nil)
