package account

import (
	"github.com/rigochain/rigo-go/types"
	"github.com/rigochain/rigo-go/types/xerrors"
	abcitypes "github.com/tendermint/tendermint/abci/types"
	tmjson "github.com/tendermint/tendermint/libs/json"
)

func (ctrler *AcctCtrler) Query(req abcitypes.RequestQuery) ([]byte, xerrors.XError) {
	acct := ctrler.ReadAccount(req.Data)
	if acct == nil {
		return nil, xerrors.ErrNotFoundAccount
	}

	// NOTE
	// `Account::Balance`, which type is *uint256.Int, is marshaled to hex-string.
	// To marshal this value to decimal format...
	_acct := &struct {
		Address types.Address `json:"address"`
		Name    string        `json:"name,omitempty"`
		Nonce   uint64        `json:"nonce,string"`
		Balance string        `json:"balance"`
		Code    []byte        `json:"code,omitempty"`
		DocURL  string        `json:"docURL,omitempty"`
	}{
		Address: acct.Address,
		Name:    acct.Name,
		Nonce:   acct.Nonce,
		Balance: acct.Balance.Dec(),
		Code:    acct.Code,
		DocURL:  acct.DocURL,
	}
	if raw, err := tmjson.Marshal(_acct); err != nil {
		return nil, xerrors.ErrQuery.Wrap(err)
	} else {
		return raw, nil
	}
}
