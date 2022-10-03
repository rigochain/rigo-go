package rpc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

var (
	rpcIndex = int64(0)
)

type JSONRpcClient struct {
	url        string
	httpClient *http.Client
}

var defaultRpcClient = &JSONRpcClient{
	url: "http://localhost:26657",
	httpClient: &http.Client{
		//Timeout: time.Second * time.Duration(10), // for [connect ~ request ~ response] time
		Transport: &http.Transport{
			DisableKeepAlives: false,
			IdleConnTimeout:   time.Minute,
			MaxConnsPerHost:   100,
		},
	},
}

func DefaultRpcClient() *JSONRpcClient {
	return defaultRpcClient
}

func NewRpcClient(url string, opts ...func(*JSONRpcClient)) *JSONRpcClient {
	ret := &JSONRpcClient{
		url: url,
		httpClient: &http.Client{
			//Timeout: time.Second * time.Duration(10), // for [connect ~ request ~ response] time
			Transport: &http.Transport{
				DisableKeepAlives: false,
				IdleConnTimeout:   time.Minute,
				MaxConnsPerHost:   100,
			},
		},
	}

	for _, cb := range opts {
		cb(ret)
	}
	return ret
}

func (client *JSONRpcClient) call(req *JSONRpcReq) (*JSONRpcResp, error) {
	reqbz, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpBody := bytes.NewBuffer(reqbz)
	httpResp, err := client.httpClient.Post(client.url, "application/json", httpBody)
	if err != nil {
		return nil, err
	}

	defer func() {
		httpResp.Body.Close()
	}()

	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Bad HTTP Response: %v", httpResp.Status)
	}

	respBody, err := ioutil.ReadAll(httpResp.Body)
	if err != nil {
		return nil, err
	}

	res := &JSONRpcResp{}
	if err = json.Unmarshal(respBody, res); err != nil {
		return nil, err
	}
	return res, nil
}

type JSONRpcReq struct {
	Version string          `json:"jsonrpc"`
	Id      int64           `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
}

type JSONRpcResp struct {
	Version string          `json:"version"`
	Id      int64           `json:"id"`
	Result  json.RawMessage `json:"result"`
	Error   json.RawMessage `json:"error"`
}

func NewRequest(method string, args ...interface{}) (*JSONRpcReq, error) {
	params, err := json.Marshal(args)
	if err != nil {
		return nil, err
	}
	rpcIndex++
	ret := &JSONRpcReq{
		Id:     rpcIndex,
		Method: method,
		Params: params,
	}

	//ret.Params = "["
	//for _, a := range args {
	//	ret.Params += a
	//}
	//ret.Params += "]"

	return ret, nil
}
