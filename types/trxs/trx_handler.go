package trxs

type ITrxHandler interface {
	Validate(*TrxContext) error
	Apply(*TrxContext) error
}
