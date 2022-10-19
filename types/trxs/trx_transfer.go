package trxs

type TrxPayloadAssetTransfer struct{}

func (tx *TrxPayloadAssetTransfer) Type() int32 {
	return TRX_TRANSFER
}

func (tx *TrxPayloadAssetTransfer) Decode(bz []byte) error {
	//pm := TrxPayloadAssetTransferProto{}
	//if err := proto.Unmarshal(bz, &pm); err != nil {
	//	return xerrors.Wrap(err)
	//}
	//tx.amount = new(big.Int).SetBytes(pm.XAmount)
	//return nil
	return nil
}

func (tx *TrxPayloadAssetTransfer) Encode() ([]byte, error) {
	//pm := TrxPayloadAssetTransferProto{
	//	XAmount: tx.amount.Bytes(),
	//}
	//if bz, err := proto.Marshal(&pm); err != nil {
	//	return nil, xerrors.Wrap(err)
	//} else {
	//	return bz, nil
	//}
	return nil, nil
}
