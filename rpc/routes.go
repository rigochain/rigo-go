package rpc

import (
	tmrpccore "github.com/tendermint/tendermint/rpc/core"
	tmrpccore_server "github.com/tendermint/tendermint/rpc/jsonrpc/server"
)

func AddRoutes() {
	tmrpccore.Routes["account"] = tmrpccore_server.NewRPCFunc(QueryAccount, "addr")
	tmrpccore.Routes["delegatee"] = tmrpccore_server.NewRPCFunc(QueryDelegatee, "addr")
	tmrpccore.Routes["stakes"] = tmrpccore_server.NewRPCFunc(QueryStakes, "addr")
	tmrpccore.Routes["proposals"] = tmrpccore_server.NewRPCFunc(QueryProposals, "txhash")
	tmrpccore.Routes["rule"] = tmrpccore_server.NewRPCFunc(QueryRule, "")
	tmrpccore.Routes["vm_call"] = tmrpccore_server.NewRPCFunc(QueryVM, "addr,to,height,data")
	tmrpccore.Routes["subscribe"] = tmrpccore_server.NewRPCFunc(Subscribe, "query")
	tmrpccore.Routes["unsubscribe"] = tmrpccore_server.NewRPCFunc(Unsubscribe, "query")
	tmrpccore.Routes["tx_search"] = tmrpccore_server.NewRPCFunc(TxSearch, "query,prove,page,per_page,order_by")
	tmrpccore.Routes["validators"] = tmrpccore_server.NewRPCFunc(Validators, "height,page,per_page")
}
