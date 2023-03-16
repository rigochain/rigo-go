package client

import (
	"github.com/gorilla/websocket"
	"github.com/rigochain/rigo-go/libs/client/rpc"
	"github.com/tendermint/tendermint/libs/json"
)

type Subscriber struct {
	conn *websocket.Conn
	done chan interface{}
}

func NewSubscriber(url string) (*Subscriber, error) {
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, err
	}

	return &Subscriber{
		conn: conn,
		done: make(chan interface{}),
	}, nil
}

func (sub *Subscriber) Watch(query string, callback func(*Subscriber, []byte)) error {
	req, err := rpc.NewRequest("subscribe", query)
	if err != nil {
		return err
	}

	bz, err := json.Marshal(req)
	if err != nil {
		return err
	}

	err = sub.conn.WriteMessage(websocket.TextMessage, bz)
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case <-sub.done:
				return
			default:
				ty, msg, err := sub.conn.ReadMessage()
				if err != nil {
					return
				}

				if ty == 1 {
					resp := &rpc.JSONRpcResp{}
					if err := json.Unmarshal(msg, resp); err != nil {
						panic(err)
					}

					if resp.Error != nil {
						panic(string(resp.Error))
					}

					if len(resp.Result) > 2 {
						callback(sub, resp.Result)
					}
				}
			}

		}
	}()

	return nil
}

func (sub *Subscriber) Stop() {
	close(sub.done)
	_ = sub.conn.Close()
}
