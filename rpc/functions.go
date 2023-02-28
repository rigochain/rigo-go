package rpc

import (
	abytes "github.com/rigochain/rigo-go/types/bytes"
	"github.com/rigochain/rigo-go/types/xerrors"
	tmbytes "github.com/tendermint/tendermint/libs/bytes"
	tmrpccore "github.com/tendermint/tendermint/rpc/core"
	tmrpccoretypes "github.com/tendermint/tendermint/rpc/core/types"
	tmrpctypes "github.com/tendermint/tendermint/rpc/jsonrpc/types"
	"regexp"
	"strings"
)

var hexReg = regexp.MustCompile(`(?i)[a-f0-9]{40,}`)

func QueryAccount(ctx *tmrpctypes.Context, addr abytes.HexBytes) (*QueryResult, error) {
	if resp, err := tmrpccore.ABCIQuery(ctx, "account", tmbytes.HexBytes(addr), 0, false); err != nil {
		return nil, err
	} else {
		return &QueryResult{resp.Response}, nil
	}
}

func QueryDelegatee(ctx *tmrpctypes.Context, addr abytes.HexBytes) (*QueryResult, error) {
	if resp, err := tmrpccore.ABCIQuery(ctx, "delegatee", tmbytes.HexBytes(addr), 0, false); err != nil {
		return nil, err
	} else {
		return &QueryResult{resp.Response}, nil
	}
}

func QueryStakes(ctx *tmrpctypes.Context, addr abytes.HexBytes) (*QueryResult, error) {
	if resp, err := tmrpccore.ABCIQuery(ctx, "stakes", tmbytes.HexBytes(addr), 0, false); err != nil {
		return nil, err
	} else {
		return &QueryResult{resp.Response}, nil
	}
}

func QueryProposals(ctx *tmrpctypes.Context, txhash abytes.HexBytes) (*QueryResult, error) {
	if resp, err := tmrpccore.ABCIQuery(ctx, "proposals", tmbytes.HexBytes(txhash), 0, false); err != nil {
		return nil, err
	} else {
		return &QueryResult{resp.Response}, nil
	}
}

func QueryRule(ctx *tmrpctypes.Context) (*QueryResult, error) {

	if resp, err := tmrpccore.ABCIQuery(ctx, "rule", nil, 0, false); err != nil {
		return nil, err
	} else {
		return &QueryResult{resp.Response}, nil
	}
}

func Subscribe(ctx *tmrpctypes.Context, query string) (*tmrpccoretypes.ResultSubscribe, error) {
	// return error when the event subscription request is received over http session.
	// related to: #103
	if ctx.WSConn == nil || ctx.JSONReq == nil {
		return nil, xerrors.New("error connection type: no websocket connection")
	}
	// make hex string like address or hash be uppercase
	//  address's size is 20bytes(40characters)
	//  hash's size is 32bytes(64characters)
	return tmrpccore.Subscribe(ctx, hexReg.ReplaceAllStringFunc(query, strings.ToUpper))
}

func Unsubscribe(ctx *tmrpctypes.Context, query string) (*tmrpccoretypes.ResultUnsubscribe, error) {
	return tmrpccore.Unsubscribe(ctx, hexReg.ReplaceAllStringFunc(query, strings.ToUpper))
}

func TxSearch(
	ctx *tmrpctypes.Context,
	query string,
	prove bool,
	pagePtr, perPagePtr *int,
	orderBy string,
) (*tmrpccoretypes.ResultTxSearch, error) {
	return tmrpccore.TxSearch(ctx, hexReg.ReplaceAllStringFunc(query, strings.ToUpper), prove, pagePtr, perPagePtr, orderBy)
}
