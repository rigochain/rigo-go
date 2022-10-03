package types

import (
	tmbytes "github.com/tendermint/tendermint/libs/bytes"
)

type HexBytes = tmbytes.HexBytes
type Address = HexBytes
type AcctKey = [AddrSize]byte
