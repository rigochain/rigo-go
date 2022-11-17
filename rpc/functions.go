package rpc

import (
	"encoding/hex"
	"encoding/json"
	"github.com/kysee/arcanus/types"
	"github.com/kysee/arcanus/types/xerrors"
	tmrpccore "github.com/tendermint/tendermint/rpc/core"
	tmrpccoretypes "github.com/tendermint/tendermint/rpc/core/types"
	tmrpctypes "github.com/tendermint/tendermint/rpc/jsonrpc/types"
	"regexp"
	"strings"
)

var hexReg = regexp.MustCompile(`(?i)[a-f0-9]{40,}`)

func QueryAccount(ctx *tmrpctypes.Context, addr string) (json.RawMessage, error) {
	bzAddr, err := types.AddressFromHex(addr)
	if err != nil {
		return nil, xerrors.NewFrom(err)
	}

	qd := &types.QueryData{
		Command: types.QUERY_ACCOUNT,
		Params:  bzAddr,
	}
	resp, err := tmrpccore.ABCIQuery(ctx, ctx.HTTPReq.RequestURI, qd.Encode(), 0, false)
	if err != nil {
		// not reachable
		// ABCIQuery never returns error
		return nil, err
	}

	if resp.Response.Code == xerrors.ErrCodeSuccess {
		return resp.Response.Value, nil
	} else {
		return nil, xerrors.NewWith(resp.Response.Code, resp.Response.Log)
	}
}

func QueryStakes(ctx *tmrpctypes.Context, addr string) (json.RawMessage, error) {
	bzAddr, err := types.AddressFromHex(addr)
	if err != nil {
		return nil, xerrors.NewFrom(err)
	}

	qd := &types.QueryData{
		Command: types.QUERY_STAKES,
		Params:  bzAddr,
	}

	resp, err := tmrpccore.ABCIQuery(ctx, ctx.HTTPReq.RequestURI, qd.Encode(), 0, false)
	if err != nil {
		// not reachable
		// ABCIQuery never returns error
		return nil, err
	}

	if resp.Response.Code == xerrors.ErrCodeSuccess {
		return resp.Response.Value, nil
	} else {
		return nil, xerrors.NewWith(resp.Response.Code, resp.Response.Log)
	}
}

func QueryProposal(ctx *tmrpctypes.Context, txhash string) (json.RawMessage, error) {
	bzTxHash, err := hex.DecodeString(txhash)
	if err != nil {
		return nil, xerrors.NewFrom(err)
	}

	qd := &types.QueryData{
		Command: types.QUERY_PROPOSALS,
		Params:  bzTxHash,
	}

	resp, err := tmrpccore.ABCIQuery(ctx, ctx.HTTPReq.RequestURI, qd.Encode(), 0, false)
	if err != nil {
		// not reachable
		// ABCIQuery never returns error
		return nil, err
	}

	if resp.Response.Code == xerrors.ErrCodeSuccess {
		return resp.Response.Value, nil
	} else {
		return nil, xerrors.NewWith(resp.Response.Code, resp.Response.Log)
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
