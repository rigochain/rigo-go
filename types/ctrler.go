package types

type ILedgerCtrler interface {
	Commit() ([]byte, int64, error)
	Close() error
}
