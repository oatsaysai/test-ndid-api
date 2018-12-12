package common

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/labstack/echo"
	"github.com/satori/go.uuid"
	"github.com/test-ndid-api/server/config"
	"github.com/test-ndid-api/server/core/crypto"
	"github.com/test-ndid-api/server/tendermint"
)

type Request struct {
	Mode            int           `json:"mode"`
	Namespace       string        `json:"namespace"`
	Identifier      string        `json:"identifier"`
	ReferenceID     string        `json:"reference_id"`
	CallbackURL     string        `json:"callback_url"`
	IdpIDList       []string      `json:"idp_id_list"`
	DataRequestList []DataRequest `json:"data_request_list"`
	RequestMessage  string        `json:"request_message"`
	MinIal          float64       `json:"min_ial"`
	MinAal          float64       `json:"min_aal"`
	MinIdp          int           `json:"min_idp"`
	RequestTimeout  int           `json:"request_timeout"`
}

type Responses struct {
	RequestID string `json:"request_id"`
}

type DataRequest struct {
	ServiceID     string   `json:"service_id"`
	AsIDList      []string `json:"as_id_list"`
	MinAs         int      `json:"min_as"`
	RequestParams struct {
		ID    int `json:"id"`
		Count int `json:"count"`
	} `json:"request_params"`
}

type Response struct {
	Ial              float64 `json:"ial"`
	Aal              float64 `json:"aal"`
	Status           string  `json:"status"`
	Signature        string  `json:"signature"`
	IdentityProof    string  `json:"identity_proof"`
	PrivateProofHash string  `json:"private_proof_hash"`
	IdpID            string  `json:"idp_id"`
	ValidProof       *bool   `json:"valid_proof"`
	ValidIal         *bool   `json:"valid_ial"`
	ValidSignature   *bool   `json:"valid_signature"`
}

type RequestToTm struct {
	RequestID       string        `json:"request_id"`
	MinIdp          int           `json:"min_idp"`
	MinAal          float64       `json:"min_aal"`
	MinIal          float64       `json:"min_ial"`
	Timeout         int           `json:"request_timeout"`
	DataRequestList []DataRequest `json:"data_request_list"`
	MessageHash     string        `json:"request_message_hash"`
	Mode            int           `json:"mode"`
}

func CreateRequest(c echo.Context) error {
	startTime := time.Now()
	cfg := config.LoadConfiguration()
	request := new(Request)
	if err := c.Bind(request); err != nil {
		return err
	}
	id := uuid.Must(uuid.NewV4(), nil)
	var param RequestToTm
	param.RequestID = id.String()

	writeLogTimeWithNonce(string("CreateRequest"), []byte(id.String()), startTime)

	param.MinIdp = request.MinIdp
	param.MinAal = request.MinAal
	param.MinIal = request.MinIal
	param.Timeout = request.RequestTimeout
	param.DataRequestList = request.DataRequestList
	param.MessageHash = "Hash(" + request.RequestMessage + ")"
	param.Mode = request.Mode
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	signature, nonce := crypto.CreateSignatureAndNonce("CreateRequest", string(paramJSON), cfg.NodeID)

	// Event log
	writeEventLog(cfg.NodeID, startTime, "create_request_before_tm_tx", param.RequestID)

	go tendermint.Transact([]byte("CreateRequest"), paramJSON, []byte(nonce), signature, []byte(cfg.NodeID))
	// tendermint.Transact([]byte("CreateRequest"), paramJSON, []byte(nonce), signature, []byte(cfg.NodeID))
	stopTime := time.Now()
	writeLogTimeWithNonce(string("CreateRequest"), []byte(id.String()), stopTime)
	// writeLog(string("CreateRequest"), (stopTime.UnixNano()-startTime.UnixNano())/int64(time.Millisecond))
	return c.JSON(http.StatusCreated, &Responses{id.String()})
}

func writeLog(filename string, time int64) {
	createDirIfNotExist("http_log")
	f, err := os.OpenFile("http_log/"+filename+".log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	_, err = f.WriteString(strconv.FormatInt(time, 10) + "\r\n")
	if err != nil {
		panic(err)
	}
}

func writeLogTimeWithNonce(filename string, nonce []byte, time time.Time) {
	createDirIfNotExist("http_log")
	f, err := os.OpenFile("http_log/"+filename+".log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
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

func createDirIfNotExist(dir string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			panic(err)
		}
	}
}

func writeEventLog(filename string, time time.Time, name string, requestID string) {
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
}
