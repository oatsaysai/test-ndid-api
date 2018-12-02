package tendermint

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/gorilla/websocket"
	"github.com/tendermint/tendermint/libs/common"
	protoTm "github.com/test-ndid-api/protos/tendermint"
	"github.com/test-ndid-api/server/config"
)

var (
	event = Event{nil}
	ws    = WebSocket{nil}
)

type txResult struct {
	Jsonrpc string `json:"jsonrpc"`
	ID      string `json:"id"`
	Result  struct {
		Query string `json:"query"`
		Data  struct {
			Type  string `json:"type"`
			Value struct {
				TxResult struct {
					Height string `json:"height"`
					Index  int    `json:"index"`
					Tx     string `json:"tx"`
					Result struct {
						Log string `json:"log"`
						Fee struct {
						} `json:"fee"`
					} `json:"result"`
				} `json:"TxResult"`
			} `json:"value"`
		} `json:"data"`
	} `json:"result"`
}

type ResponseTx struct {
	Result struct {
		Height  int `json:"height"`
		CheckTx struct {
			Code int      `json:"code"`
			Log  string   `json:"log"`
			Fee  struct{} `json:"fee"`
		} `json:"check_tx"`
		DeliverTx struct {
			Log  string   `json:"log"`
			Fee  struct{} `json:"fee"`
			Tags []common.KVPair
		} `json:"deliver_tx"`
		Hash string `json:"hash"`
	} `json:"result"`
	Jsonrpc string `json:"jsonrpc"`
	ID      string `json:"id"`
}

type ResponseTxSync struct {
	Jsonrpc string `json:"jsonrpc"`
	ID      string `json:"id"`
	Result  struct {
		Code int    `json:"code"`
		Data string `json:"data"`
		Log  string `json:"log"`
		Hash string `json:"hash"`
	} `json:"result"`
}

type JsonRPCQuery struct {
	Jsonrpc string `json:"jsonrpc"`
	Method  string `json:"method"`
	ID      string `json:"id"`
	Params  struct {
		Query string `json:"query"`
	} `json:"params"`
}

// var tendermintAddr = getEnv("TENDERMINT_ADDRESS", "http://localhost:45000")
// var tendermintAddr = "http://localhost:45000"

func Transact(fnName []byte, param []byte, nonce []byte, signature []byte, nodeID []byte) {
	cfg := config.LoadConfiguration()
	if cfg.Role == "NDID" {
		broadcastTxCommit(fnName, param, nonce, signature, nodeID)
	} else {
		// broadcastTxCommit(fnName, param, nonce, signature, nodeID)
		broadcastTxSync(fnName, param, nonce, signature, nodeID)
	}
}

func broadcastTxSync(fnName []byte, param []byte, nonce []byte, signature []byte, nodeID []byte) (interface{}, error) {

	startTime := time.Now()
	writeLogTimeWithNonce(string(fnName), nonce, startTime)
	cfg := config.LoadConfiguration()

	var tx protoTm.Tx
	tx.Method = string(fnName)
	tx.Params = string(param)
	tx.Nonce = nonce
	tx.Signature = signature
	tx.NodeId = string(nodeID)

	txByte, err := proto.Marshal(&tx)
	if err != nil {
		log.Printf("err: %s", err.Error())
	}

	txEncoded := hex.EncodeToString(txByte)

	tmAddr := "http://" + cfg.TendermintIP + ":" + cfg.TendermintPort
	var URL *url.URL
	URL, err = url.Parse(tmAddr)
	if err != nil {
		panic("boom")
	}
	URL.Path += "/broadcast_tx_sync"
	parameters := url.Values{}
	parameters.Add("tx", `0x`+txEncoded)
	URL.RawQuery = parameters.Encode()
	encodedURL := URL.String()
	req, err := http.NewRequest("GET", encodedURL, nil)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		// panic(err)
		return nil, err
	}
	defer resp.Body.Close()

	var body ResponseTx
	json.NewDecoder(resp.Body).Decode(&body)
	return body, nil

	// // fmt.Println("broadcastTxSync")
	// cfg := config.LoadConfiguration()
	// startTime := time.Now()
	// writeLogTimeWithNonce(string(fnName), nonce, startTime)
	// var path string
	// path += string(fnName)
	// path += "|"
	// path += base64.StdEncoding.EncodeToString(param)
	// path += "|"
	// path += string(nonce)
	// path += "|"
	// path += base64.StdEncoding.EncodeToString(signature)
	// path += "|"
	// path += base64.StdEncoding.EncodeToString(nodeID)
	// var URL *url.URL
	// tmAddr := "http://" + cfg.TendermintIP + ":" + cfg.TendermintPort
	// URL, err := url.Parse(tmAddr)
	// if err != nil {
	// 	panic("boom")
	// }
	// URL.Path += "/broadcast_tx_sync"
	// parameters := url.Values{}
	// parameters.Add("tx", `"`+path+`"`)
	// URL.RawQuery = parameters.Encode()
	// encodedURL := URL.String()
	// req, err := http.NewRequest("GET", encodedURL, nil)
	// if err != nil {
	// 	fmt.Println(err.Error())
	// 	return nil, err
	// }
	// req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	// client := &http.Client{
	// 	CheckRedirect: func(req *http.Request, via []*http.Request) error {
	// 		return http.ErrUseLastResponse
	// 	},
	// }
	// resp, err := client.Do(req)
	// if err != nil {
	// 	fmt.Println(err.Error())
	// 	return nil, err
	// }
	// defer resp.Body.Close()
	// var body ResponseTxSync
	// json.NewDecoder(resp.Body).Decode(&body)

	// // sub(string(nonce), body.Result.Hash, string(fnName))

	// // txEncoded := base64.StdEncoding.EncodeToString(signature)
	// // ch := make(chan string)
	// // event.AddListener(txEncoded, ch)
	// // var wg sync.WaitGroup
	// // wg.Add(1)
	// // var log string
	// // go func(wg *sync.WaitGroup) (string, error) {
	// // 	for {
	// // 		msg := <-ch
	// // 		log = msg
	// // 		event.RemoveListener(txEncoded, ch)
	// // 		wg.Done()
	// // 	}
	// // }(&wg)
	// // go sub(txEncoded, body.Result.Hash)
	// // wg.Wait()
	// // stopTime := time.Now()
	// // writeLogTimeWithNonce(string(fnName), nonce, stopTime)
	// return body, nil
}

func broadcastTxCommit(fnName []byte, param []byte, nonce []byte, signature []byte, nodeID []byte) (interface{}, error) {
	// fmt.Println("broadcastTxCommit")
	cfg := config.LoadConfiguration()
	startTime := time.Now()
	writeLogTimeWithNonce(string(fnName), nonce, startTime)

	// var path string
	// path += string(fnName)
	// path += "|"
	// path += base64.StdEncoding.EncodeToString(param)
	// path += "|"
	// path += string(nonce)
	// path += "|"
	// path += base64.StdEncoding.EncodeToString(signature)
	// path += "|"
	// path += base64.StdEncoding.EncodeToString(nodeID)

	var tx protoTm.Tx
	tx.Method = string(fnName)
	tx.Params = string(param)
	tx.Nonce = nonce
	tx.Signature = signature
	tx.NodeId = string(nodeID)

	txByte, err := proto.Marshal(&tx)
	if err != nil {
		log.Printf("err: %s", err.Error())
	}

	txEncoded := hex.EncodeToString(txByte)

	tmAddr := "http://" + cfg.TendermintIP + ":" + cfg.TendermintPort
	var URL *url.URL
	URL, err = url.Parse(tmAddr)
	if err != nil {
		panic("boom")
	}
	URL.Path += "/broadcast_tx_commit"
	parameters := url.Values{}
	parameters.Add("tx", `0x`+txEncoded)

	URL.RawQuery = parameters.Encode()
	encodedURL := URL.String()
	req, err := http.NewRequest("GET", encodedURL, nil)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	defer resp.Body.Close()
	var body ResponseTx
	json.NewDecoder(resp.Body).Decode(&body)
	// stopTime := time.Now()
	// writeLogTimeWithNonce(string(fnName), nonce, stopTime)
	return body, nil
}

func writeLogTimeWithNonce(filename string, nonce []byte, time time.Time) {
	createDirIfNotExist("log")
	f, err := os.OpenFile("log/"+filename+".log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	cfg := config.LoadConfiguration()
	_, err = f.WriteString(string(nonce) + "|" + strconv.FormatInt(time.UnixNano(), 10) + "|" + cfg.NodeID + "\r\n")
	if err != nil {
		panic(err)
	}
}

func writeLog(filename string, time int64) {
	createDirIfNotExist("log")
	f, err := os.OpenFile("log/"+filename+".log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	_, err = f.WriteString(strconv.FormatInt(time, 10) + "\r\n")
	if err != nil {
		panic(err)
	}
}

func createDirIfNotExist(dir string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			panic(err)
		}
	}
}

func sub(txEncoded, hash, fnName string) {
	c := ws.CreateNewIfNotExist()
	// defer c.Close()

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read err:", err)
				return
			}
			// log.Printf("recv: %s", message)
			var res txResult
			err = json.Unmarshal(message, &res)
			if err != nil {
				log.Println("err:", err)
				return
			}
			if res.Result.Data.Value.TxResult.Result.Log != "" {
				stopTime := time.Now()
				writeLogTimeWithNonce(string(fnName), []byte(txEncoded), stopTime)
				// event.Emit(txEncoded, res.Result.Data.Value.TxResult.Result.Log)
				return
			}
		}
	}()

	var query JsonRPCQuery
	query.Jsonrpc = "2.0"
	query.Method = "subscribe"
	query.ID = txEncoded
	query.Params.Query = `tm.event = 'NewBlock'`
	// query.Params.Query = `tm.event = 'Tx' AND tx.hash = '` + hash + `'`

	str, err := json.Marshal(query)
	if err != nil {
		log.Println("err:", err)
	}

	err = c.WriteMessage(websocket.TextMessage, str)
	if err != nil {
		log.Println("write err:", err)
		return
	}

	for {
		select {
		case <-done:
			// log.Println("done")
			return
		}
	}

	// flag.Parse()
	// log.SetFlags(0)

	// interrupt := make(chan os.Signal, 1)
	// signal.Notify(interrupt, os.Interrupt)

	// cfg := config.LoadConfiguration()
	// addr := cfg.TendermintIP + ":" + cfg.TendermintPort

	// u := url.URL{Scheme: "ws", Host: addr, Path: "/websocket"}
	// // log.Printf("connecting to %s", u.String())

	// c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	// if err != nil {
	// 	log.Fatal("dial:", err)
	// }
	// defer c.Close()

	// done := make(chan struct{})

	// go func() {
	// 	defer close(done)
	// 	for {
	// 		_, message, err := c.ReadMessage()
	// 		if err != nil {
	// 			log.Println("read err:", err)
	// 			return
	// 		}
	// 		// log.Printf("recv: %s", message)
	// 		var res txResult
	// 		err = json.Unmarshal(message, &res)
	// 		if err != nil {
	// 			log.Println("err:", err)
	// 			return
	// 		}
	// 		if res.Result.Data.Value.TxResult.Result.Log != "" {
	// 			stopTime := time.Now()
	// 			writeLogTimeWithNonce(string(fnName), []byte(txEncoded), stopTime)
	// 			event.Emit(txEncoded, res.Result.Data.Value.TxResult.Result.Log)
	// 			return
	// 		}
	// 	}
	// }()

	// var query JsonRPCQuery
	// query.Jsonrpc = "2.0"
	// query.Method = "subscribe"
	// query.ID = txEncoded
	// // query.Params.Query = `tm.event = 'NewBlock'`
	// query.Params.Query = `tm.event = 'Tx' AND tx.hash = '` + hash + `'`

	// str, err := json.Marshal(query)
	// if err != nil {
	// 	log.Println("err:", err)
	// }

	// err = c.WriteMessage(websocket.TextMessage, str)
	// if err != nil {
	// 	log.Println("write err:", err)
	// 	return
	// }

	// for {
	// 	select {
	// 	case <-done:
	// 		// log.Println("done")
	// 		return
	// 	}
	// }
}

type WebSocket struct {
	con *websocket.Conn
}

func (w *WebSocket) CreateNewIfNotExist() *websocket.Conn {
	if w.con == nil {
		cfg := config.LoadConfiguration()
		addr := cfg.TendermintIP + ":" + cfg.TendermintPort
		u := url.URL{Scheme: "ws", Host: addr, Path: "/websocket"}
		var err error
		w.con, _, err = websocket.DefaultDialer.Dial(u.String(), nil)
		if err != nil {
			log.Fatal("dial:", err)
		}
		// defer w.con.Close()
	}
	return w.con
}

type Event struct {
	listeners map[string][]chan string
}

func (b *Event) AddListener(e string, ch chan string) {
	if b.listeners == nil {
		b.listeners = make(map[string][]chan string)
	}
	if _, ok := b.listeners[e]; ok {
		b.listeners[e] = append(b.listeners[e], ch)
	} else {
		b.listeners[e] = []chan string{ch}
	}
}

func (b *Event) RemoveListener(e string, ch chan string) {
	if _, ok := b.listeners[e]; ok {
		for i := range b.listeners[e] {
			if b.listeners[e][i] == ch {
				b.listeners[e] = append(b.listeners[e][:i], b.listeners[e][i+1:]...)
				break
			}
		}
	}
}

func (b *Event) Emit(e string, response string) {
	if _, ok := b.listeners[e]; ok {
		for _, handler := range b.listeners[e] {
			go func(handler chan string) {
				handler <- response
			}(handler)
		}
	}
}
