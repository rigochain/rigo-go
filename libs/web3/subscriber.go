package web3

import (
	"github.com/gorilla/websocket"
	"github.com/rigochain/rigo-go/libs/web3/types"
	"github.com/tendermint/tendermint/libs/json"
)

type Subscriber struct {
	url  string
	conn *websocket.Conn

	query string
}

func NewSubscriber(url string) (*Subscriber, error) {
	return &Subscriber{
		url: url,
	}, nil
}

func (sub *Subscriber) Start(query string, callback func(*Subscriber, []byte)) error {
	conn, _, err := websocket.DefaultDialer.Dial(sub.url, nil)
	if err != nil {
		return err
	}

	req, err := types.NewRequest(0, "subscribe", query)
	if err != nil {
		return err
	}

	bz, err := json.Marshal(req)
	if err != nil {
		return err
	}

	err = conn.WriteMessage(websocket.TextMessage, bz)
	if err != nil {
		return err
	}

	go func() {
		for {
			ty, msg, err := conn.ReadMessage()
			if err != nil {
				break
			}

			if ty == websocket.TextMessage {
				resp := &types.JSONRpcResp{}
				if err := json.Unmarshal(msg, resp); err != nil {
					panic(err)
				}

				if resp.Error != nil {
					panic(string(resp.Error))
				}

				if len(resp.Result) > 2 {
					callback(sub, resp.Result)
				}
			} else {
				//fmt.Println("ReadMessage", "other type", ty, msg)
			}
		}
	}()

	sub.conn = conn
	sub.query = query

	return nil
}

func (sub *Subscriber) Stop() {
	req, err := types.NewRequest(1, "unsubscribe", sub.query)
	if err != nil {
		panic(err)
	}

	bz, err := json.Marshal(req)
	if err != nil {
		panic(err)
	}

	err = sub.conn.WriteMessage(websocket.TextMessage, bz)
	if err != nil {
		panic(err)
	}
	sub.query = ""

	_ = sub.conn.Close()
	sub.conn = nil
}
