package types

import "github.com/rigochain/rigo-go/types/xerrors"

type IEncoder interface {
	Encode() ([]byte, xerrors.XError)
	Decode([]byte) xerrors.XError
}
