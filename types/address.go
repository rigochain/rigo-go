package types

import (
	"encoding/hex"
	abytes "github.com/kysee/arcanus/types/bytes"
	"github.com/kysee/arcanus/types/xerrors"
	"github.com/tendermint/tendermint/crypto"
	"strings"
)

const AddrSize = 20
const (
	ACCT_COMMON_TYPE int16 = 1 + iota
)

type Address = abytes.HexBytes

func RandAddress() Address {
	return abytes.RandBytes(AddrSize)
}

func ZeroAddress() Address {
	return abytes.ZeroBytes(AddrSize)
}

func HexToAddress(_hex string) (Address, error) {
	if strings.HasPrefix(_hex, "0x") {
		_hex = _hex[2:]
	}
	bzAddr, err := hex.DecodeString(_hex)
	if err != nil {
		return nil, xerrors.NewFrom(err)
	}
	if len(bzAddr) != crypto.AddressSize {
		return nil, xerrors.New("error of address length: address length should be 20 bytes")
	}
	return bzAddr, nil
}