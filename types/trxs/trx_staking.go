package trxs

type TrxPayloadStaking struct{}

func (tx *TrxPayloadStaking) GetType() int32 {
	return TRX_STAKING
}

func (tx *TrxPayloadStaking) Decode(bz []byte) error {
	return nil
}

func (tx *TrxPayloadStaking) Encode() ([]byte, error) {
	return nil, nil
}

type TrxPayloadUnstaking struct{}

func (tx *TrxPayloadUnstaking) GetType() int32 {
	return TRX_UNSTAKING
}

func (tx *TrxPayloadUnstaking) Decode(bz []byte) error {
	return nil
}

func (tx *TrxPayloadUnstaking) Encode() ([]byte, error) {
	return nil, nil
}
