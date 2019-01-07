package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	protoTm "github.com/test-ndid-api/protos/tendermint"
	"github.com/test-ndid-api/server/config"
	"github.com/test-ndid-api/server/core/common"
	"github.com/test-ndid-api/server/tendermint"
	wsClent "github.com/test-ndid-api/server/webSocket"
)

var (
	ws = wsClent.NewWebSocketClient()
)

var port string

func main() {
	cfg := config.LoadConfiguration()
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// CORS
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{echo.GET, echo.HEAD, echo.PUT, echo.PATCH, echo.POST, echo.DELETE},
	}))

	// common
	e.POST("/v2/rp/requests/:namespace/:identifier", common.CreateRequest)
	e.POST("/v2/setToken", common.SetNodeToken)

	go sub()
	// go subTx()

	// ws.InitConnection()
	// go ws.Connect()

	// Server
	e.Logger.Fatal(e.Start(":" + cfg.ServerListenPort))
}

type JsonRPCQuery struct {
	Jsonrpc string `json:"jsonrpc"`
	Method  string `json:"method"`
	ID      string `json:"id"`
	Params  struct {
		Query string `json:"query"`
	} `json:"params"`
}

func sub() {
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
			var res NewBlockResult
			err = json.Unmarshal(message, &res)
			if err != nil {
				log.Println("err:", err)
				return
			}
			writeLogBlockResult("block", res.Result.Data.Value.Block.Header.Height, res.Result.Data.Value.Block.Header.NumTxs)

			// get block result
			tendermint.GetBlockResult(res.Result.Data.Value.Block.Header.Height)

			go func() {
				for _, tx := range res.Result.Data.Value.Block.Data.Txs {

					decodedTx, err := base64.StdEncoding.DecodeString(tx)
					if err != nil {
						log.Println("err:", err)
					}

					var txObj protoTm.Tx
					err = proto.Unmarshal(decodedTx, &txObj)
					if err != nil {
						log.Println("err:", err)
					}

					method := txObj.Method
					param := txObj.Params
					nonce := txObj.Nonce

					if method == "CreateRequest" {
						cfg := config.LoadConfiguration()
						var paramObj common.RequestToTm
						err = json.Unmarshal([]byte(param), &paramObj)
						stopTime := time.Now()
						// Event log
						writeEventLog(cfg.NodeID, stopTime, "create_request_after_tm_tx", paramObj.RequestID, res.Result.Data.Value.Block.Header.Height)
					}

					stopTime := time.Now()
					writeLogTimeWithNonce(string(method), []byte(nonce), stopTime)
				}
			}()
		}
	}()

	var query JsonRPCQuery
	query.Jsonrpc = "2.0"
	query.Method = "subscribe"
	query.ID = "0"
	query.Params.Query = `tm.event = 'NewBlock'`
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
}

func subTx() {
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
			message = message
			// log.Printf("recv: %s", message)
			// var res NewBlockResult
			// err = json.Unmarshal(message, &res)
			// if err != nil {
			// 	log.Println("err:", err)
			// 	return
			// }
			// writeLogBlockResult("block", res.Result.Data.Value.Block.Header.Height, res.Result.Data.Value.Block.Header.NumTxs)

			// // get block result
			// tendermint.GetBlockResult(res.Result.Data.Value.Block.Header.Height)

			// go func() {
			// 	for _, tx := range res.Result.Data.Value.Block.Data.Txs {

			// 		decodedTx, err := base64.StdEncoding.DecodeString(tx)
			// 		if err != nil {
			// 			log.Println("err:", err)
			// 		}

			// 		var txObj protoTm.Tx
			// 		err = proto.Unmarshal(decodedTx, &txObj)
			// 		if err != nil {
			// 			log.Println("err:", err)
			// 		}

			// 		method := txObj.Method
			// 		param := txObj.Params
			// 		nonce := txObj.Nonce

			// 		if method == "CreateRequest" {
			// 			cfg := config.LoadConfiguration()
			// 			var paramObj common.RequestToTm
			// 			err = json.Unmarshal([]byte(param), &paramObj)
			// 			stopTime := time.Now()
			// 			// Event log
			// 			writeEventLog(cfg.NodeID, stopTime, "create_request_after_tm_tx", paramObj.RequestID, res.Result.Data.Value.Block.Header.Height)
			// 		}

			// 		stopTime := time.Now()
			// 		writeLogTimeWithNonce(string(method), []byte(nonce), stopTime)
			// 	}
			// }()
		}
	}()

	var query JsonRPCQuery
	query.Jsonrpc = "2.0"
	query.Method = "subscribe"
	query.ID = "0"
	query.Params.Query = `tm.event = 'Tx'`
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
}

// type WebSocket struct {
// 	con *websocket.Conn
// }

// func (w *WebSocket) CreateNewIfNotExist() *websocket.Conn {
// 	if w.con == nil {
// 		cfg := config.LoadConfiguration()
// 		addr := cfg.TendermintIP + ":" + cfg.TendermintPort
// 		u := url.URL{Scheme: "ws", Host: addr, Path: "/websocket"}
// 		var err error
// 		w.con, _, err = websocket.DefaultDialer.Dial(u.String(), nil)
// 		if err != nil {
// 			log.Fatal("dial:", err)
// 		}
// 		// defer w.con.Close()
// 	}
// 	return w.con
// }

func writeLogBlockResult(filename string, blockNumber string, numTxs string) {
	createDirIfNotExist("log")
	f, err := os.OpenFile("log/"+filename+".log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	_, err = f.WriteString(blockNumber + "|" + numTxs + "\r\n")
	if err != nil {
		panic(err)
	}
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

func writeEventLog(filename string, time time.Time, name string, requestID string, block string) {
	createDirIfNotExist("log")
	f, err := os.OpenFile("log/"+filename+".log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	cfg := config.LoadConfiguration()
	var eventLog EventLog
	eventLog.Datetime = time.UnixNano() / 1000000
	eventLog.Name = name
	eventLog.RequestID = requestID
	eventLog.NodeID = cfg.NodeID
	eventLog.Block = block
	eventLogJSON, err := json.Marshal(eventLog)
	if err != nil {
		fmt.Println("error:", err)
	}
	_, err = f.WriteString(string(eventLogJSON) + "\r\n")
	if err != nil {
		panic(err)
	}
}

type EventLog struct {
	Datetime  int64  `json:"datetime"`
	Name      string `json:"name"`
	RequestID string `json:"requestId"`
	NodeID    string `json:"nodeId"`
	Block     string `json:"block"`
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

type NewBlockResult struct {
	Jsonrpc string `json:"jsonrpc"`
	ID      string `json:"id"`
	Result  struct {
		Query string `json:"query"`
		Data  struct {
			Type  string `json:"type"`
			Value struct {
				Block struct {
					Header struct {
						Version struct {
							Block string `json:"block"`
							App   string `json:"app"`
						} `json:"version"`
						ChainID     string    `json:"chain_id"`
						Height      string    `json:"height"`
						Time        time.Time `json:"time"`
						NumTxs      string    `json:"num_txs"`
						TotalTxs    string    `json:"total_txs"`
						LastBlockID struct {
							Hash  string `json:"hash"`
							Parts struct {
								Total string `json:"total"`
								Hash  string `json:"hash"`
							} `json:"parts"`
						} `json:"last_block_id"`
						LastCommitHash     string `json:"last_commit_hash"`
						DataHash           string `json:"data_hash"`
						ValidatorsHash     string `json:"validators_hash"`
						NextValidatorsHash string `json:"next_validators_hash"`
						ConsensusHash      string `json:"consensus_hash"`
						AppHash            string `json:"app_hash"`
						LastResultsHash    string `json:"last_results_hash"`
						EvidenceHash       string `json:"evidence_hash"`
						ProposerAddress    string `json:"proposer_address"`
					} `json:"header"`
					Data struct {
						Txs []string `json:"txs"`
					} `json:"data"`
					Evidence struct {
						Evidence interface{} `json:"evidence"`
					} `json:"evidence"`
					LastCommit struct {
						BlockID struct {
							Hash  string `json:"hash"`
							Parts struct {
								Total string `json:"total"`
								Hash  string `json:"hash"`
							} `json:"parts"`
						} `json:"block_id"`
						Precommits []struct {
							Type      int       `json:"type"`
							Height    string    `json:"height"`
							Round     string    `json:"round"`
							Timestamp time.Time `json:"timestamp"`
							BlockID   struct {
								Hash  string `json:"hash"`
								Parts struct {
									Total string `json:"total"`
									Hash  string `json:"hash"`
								} `json:"parts"`
							} `json:"block_id"`
							ValidatorAddress string `json:"validator_address"`
							ValidatorIndex   string `json:"validator_index"`
							Signature        string `json:"signature"`
						} `json:"precommits"`
					} `json:"last_commit"`
				} `json:"block"`
			} `json:"value"`
		} `json:"data"`
	} `json:"result"`
}
