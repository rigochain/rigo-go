package types

import (
	"encoding/binary"
	"github.com/kysee/arcanus/types/xerrors"
)

const (
	QUERY_ACCOUNT int16 = 1 + iota
	QUERY_STAKES
)

type QueryData struct {
	Command int16
	Params  []byte
}

func DecodeQueryData(bz []byte) (*QueryData, error) {
	if len(bz) < 2 {
		return nil, xerrors.ErrInvalidQueryCmd
	}
	cmd := int16(binary.BigEndian.Uint16(bz[:2]))
	params := bz[2:]
	return &QueryData{
		Command: cmd,
		Params:  params,
	}, nil
}

func (q *QueryData) Encode() []byte {
	bzcmd := make([]byte, 2)
	binary.BigEndian.PutUint16(bzcmd, uint16(q.Command))
	return append(bzcmd, q.Params...)
}
