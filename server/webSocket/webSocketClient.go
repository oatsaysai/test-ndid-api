package webSocket

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
	"github.com/test-ndid-api/server/config"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var (
	wsClientIndex = 0
)

type WebSocketClient struct {
	Connected bool
	IsAlive   bool
	Conns     *websocket.Conn
	queue     [][]byte
}

func NewWebSocketClient() *WebSocketClient {
	fmt.Println("NewWebSocketClient")
	return &WebSocketClient{
		Connected: false,
		IsAlive:   false,
		Conns:     nil,
		queue:     make([][]byte, 0),
	}
}

type WebSocketPool struct {
	rpcID       int
	Connections int
	conns       []*websocket.Conn
	connsBroken []bool
	txs         []chan []byte
}

func NewWebSocketPool(connections int) *WebSocketPool {
	return &WebSocketPool{
		rpcID:       1,
		Connections: connections,
		conns:       make([]*websocket.Conn, connections),
		connsBroken: make([]bool, connections),
		txs:         make([]chan []byte, connections),
	}
}

func connect(host string) (*websocket.Conn, *http.Response, error) {
	u := url.URL{Scheme: "ws", Host: host, Path: "/websocket"}
	return websocket.DefaultDialer.Dial(u.String(), nil)
}

func (w *WebSocketPool) InitializeAndConnect() error {
	rand.Seed(time.Now().Unix())
	cfg := config.LoadConfiguration()
	addr := cfg.TendermintIP + ":" + cfg.TendermintPort
	for i := 0; i < w.Connections; i++ {
		c, _, err := connect(addr)
		if err != nil {
			return err
		}
		w.conns[i] = c
	}
	for i := 0; i < w.Connections; i++ {
		go w.read(i)
		go w.write(i)
	}
	return nil
}

func (w *WebSocketPool) read(connIndex int) {
	defer func() {
		w.conns[connIndex].Close()
	}()
	w.conns[connIndex].SetReadLimit(maxMessageSize)
	w.conns[connIndex].SetReadDeadline(time.Now().Add(pongWait))
	w.conns[connIndex].SetPongHandler(func(string) error { w.conns[connIndex].SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, _, err := w.conns[connIndex].ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		// fmt.Printf("%s\n", message)
	}
}

func (w *WebSocketPool) write(connIndex int) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		w.conns[connIndex].Close()
	}()
	for {
		select {
		case message, ok := <-w.txs[connIndex]:
			w.conns[connIndex].SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				w.conns[connIndex].WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			ws, err := w.conns[connIndex].NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			ws.Write(message)
			n := len(w.txs[connIndex])
			for i := 0; i < n; i++ {
				ws.Write(<-w.txs[connIndex])
			}
			if err := ws.Close(); err != nil {
				return
			}
		case <-ticker.C:
			w.conns[connIndex].SetWriteDeadline(time.Now().Add(writeWait))
			if err := w.conns[connIndex].WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (w *WebSocketPool) BroadcastTxSync(tx []byte) interface{} {
	wsClientIndex++
	if wsClientIndex >= len(w.conns) {
		wsClientIndex = 0
	}
	w.rpcID++
	base64Tx := base64.StdEncoding.EncodeToString(tx)
	var query JsonRPCQuery
	query.Jsonrpc = "2.0"
	query.Method = "broadcast_tx_sync"
	query.ID = strconv.Itoa(w.rpcID)
	query.Params.Tx = base64Tx
	str, _ := json.Marshal(query)
	if w.txs[wsClientIndex] == nil {
		w.txs[wsClientIndex] = make(chan []byte, 0)
	}
	w.txs[wsClientIndex] <- str
	return nil
}

func (w *WebSocketClient) CreateNewIfNotExist() *websocket.Conn {
	if w.Conns == nil {
		cfg := config.LoadConfiguration()
		addr := cfg.TendermintIP + ":" + cfg.TendermintPort
		u := url.URL{Scheme: "ws", Host: addr, Path: "/websocket"}
		var err error
		w.Conns, _, err = websocket.DefaultDialer.Dial(u.String(), nil)
		if err != nil {
			log.Fatal("dial:", err)
		}
	}
	return w.Conns
}

type JsonRPCQuery struct {
	Jsonrpc string `json:"jsonrpc"`
	Method  string `json:"method"`
	ID      string `json:"id"`
	Params  struct {
		Query string `json:"query"`
		Tx    string `json:"tx"`
	} `json:"params"`
}

type TxCommitResult struct {
	Jsonrpc string `json:"jsonrpc"`
	ID      string `json:"id"`
	Result  struct {
		CheckTx struct {
		} `json:"check_tx"`
		DeliverTx struct {
			Data string `json:"data"`
			Log  string `json:"log"`
			Tags []struct {
				Key   string `json:"key"`
				Value string `json:"value"`
			} `json:"tags"`
		} `json:"deliver_tx"`
		Hash   string `json:"hash"`
		Height string `json:"height"`
	} `json:"result"`
}
