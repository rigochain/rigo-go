package types

import "github.com/kysee/arcanus/types/xerrors"

type IEncoder interface {
	Encode() ([]byte, xerrors.XError)
	Decode([]byte) xerrors.XError
}
