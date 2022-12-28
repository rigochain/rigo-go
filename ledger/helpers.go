package ledger

import "github.com/kysee/arcanus/types/xerrors"

func iterateItems[T ILedgerItem](m map[LedgerKey]T, cb func(T) xerrors.XError) xerrors.XError {
	for _, v := range m {
		if xerr := cb(v); xerr != nil {
			return xerr
		}
	}
	return nil
}
