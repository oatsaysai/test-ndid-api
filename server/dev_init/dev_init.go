package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/test-ndid-api/server/config"
	"github.com/test-ndid-api/server/core/ndid"
)

func main() {
	fmt.Println("========= Initializing keys for development =========")
	initNDID()
	registerNodeAndSetToken()
	addNamespace("cid", "Thai citizen ID")
	addService("bank_statement", "All transactions in the past 3 months")
	addService("customer_info", "Customer infomation")
	registerServiceDestinationByNDID("as1", "bank_statement")
	registerServiceDestinationByNDID("as1", "customer_info")
	registerServiceDestinationByNDID("as2", "bank_statement")
	registerServiceDestinationByNDID("as2", "customer_info")
	registerServiceDestinationByNDID("as3", "bank_statement")
	registerServiceDestinationByNDID("as3", "customer_info")
	setValidator(1, 10)
	setValidator(2, 10)
	setValidator(3, 10)
	setValidator(4, 10)
	setValidator(5, 10)
	setValidator(6, 10)
}

func setValidator(num, power int64) {
	var param ndid.SetValidatorParam
	param.Power = power
	param.PublicKey = getValidatorPubkey(num)
	ndid.SetValidator(param)
}

func initNDID() {
	ndidMasterPubKey, err := ioutil.ReadFile("server/dev_key/ndid/ndid1_master.pub")
	if err != nil {
		log.Fatal(err)
	}
	ndidPubKey, err := ioutil.ReadFile("server/dev_key/ndid/ndid1.pub")
	if err != nil {
		log.Fatal(err)
	}
	cfg := config.LoadConfiguration()
	var param ndid.InitNDIDParam
	param.NodeID = cfg.NodeID
	param.MasterPublicKey = string(ndidMasterPubKey)
	param.PublicKey = string(ndidPubKey)
	ndid.InitNDID(param)

	var endInitParam ndid.EndInitParam
	ndid.EndInit(endInitParam)
}

func registerNodeAndSetToken() {
	roles := [...]string{"rp", "idp", "as"}
	for _, role := range roles {
		for i := 1; i <= 4; i++ {
			masterPubKey, err := ioutil.ReadFile("server/dev_key/" + role + "/" + role + strconv.Itoa(i) + "_master.pub")
			if err != nil {
				log.Fatal(err)
			}
			pubKey, err := ioutil.ReadFile("server/dev_key/" + role + "/" + role + strconv.Itoa(i) + ".pub")
			if err != nil {
				log.Fatal(err)
			}
			var param ndid.RegisterNodeParam
			param.MasterPublicKey = string(masterPubKey)
			param.PublicKey = string(pubKey)
			param.Role = strings.ToUpper(role)
			if role == "idp" {
				param.Role = "IdP"
				param.MaxIal = 3
				param.MaxAal = 3
			}
			param.NodeID = role + strconv.Itoa(i)
			param.NodeName = "" //all anonymous
			ndid.RegisterNode(param)

			var setTokenParam ndid.SetNodeTokenParam
			setTokenParam.NodeID = param.NodeID
			setTokenParam.Amount = 100000
			ndid.SetNodeToken(setTokenParam)
		}
	}
}

func addNamespace(namespace, description string) {
	var param ndid.AddNamespaceParam
	param.Namespace = namespace
	param.Description = description
	ndid.AddNamespace(param)
}

func addService(serviceID, serviceName string) {
	var param ndid.AddServiceParam
	param.ServiceID = serviceID
	param.ServiceName = serviceName
	ndid.AddService(param)
}

func registerServiceDestinationByNDID(nodeID, serviceID string) {
	var param ndid.RegisterServiceDestinationByNDIDParam
	param.NodeID = nodeID
	param.ServiceID = serviceID
	ndid.RegisterServiceDestinationByNDID(param)
}

var tendermintAddr = "http://localhost:45000"
var tendermintAddr2 = "http://localhost:45001"
var tendermintAddr3 = "http://localhost:45002"
var tendermintAddr4 = "http://localhost:45003"
var tendermintAddr5 = "http://localhost:45004"
var tendermintAddr6 = "http://localhost:45005"

// var tendermintAddr = "http://192.168.3.182:45000"
// var tendermintAddr2 = "http://192.168.3.182:45001"
// var tendermintAddr3 = "http://192.168.3.182:45002"

func getValidatorPubkey(num int64) string {
	var URL *url.URL
	var err error
	if num == 1 {
		URL, err = url.Parse(tendermintAddr)
	} else if num == 2 {
		URL, err = url.Parse(tendermintAddr2)
	} else if num == 3 {
		URL, err = url.Parse(tendermintAddr3)
	} else if num == 4 {
		URL, err = url.Parse(tendermintAddr4)
	} else if num == 5 {
		URL, err = url.Parse(tendermintAddr5)
	} else if num == 6 {
		URL, err = url.Parse(tendermintAddr6)
	}
	if err != nil {
		panic("boom")
	}
	URL.Path += "/status"
	parameters := url.Values{}
	URL.RawQuery = parameters.Encode()
	encodedURL := URL.String()
	req, err := http.NewRequest("GET", encodedURL, nil)
	if err != nil {
		fmt.Println(err.Error())
		return ""
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	var body ResponseStatus
	json.NewDecoder(resp.Body).Decode(&body)
	return body.Result.ValidatorInfo.PubKey.Value
}

type ResponseStatus struct {
	Jsonrpc string `json:"jsonrpc"`
	ID      string `json:"id"`
	Result  struct {
		NodeInfo struct {
			ID         string   `json:"id"`
			ListenAddr string   `json:"listen_addr"`
			Network    string   `json:"network"`
			Version    string   `json:"version"`
			Channels   string   `json:"channels"`
			Moniker    string   `json:"moniker"`
			Other      []string `json:"other"`
		} `json:"node_info"`
		SyncInfo struct {
			LatestBlockHash   string    `json:"latest_block_hash"`
			LatestAppHash     string    `json:"latest_app_hash"`
			LatestBlockHeight string    `json:"latest_block_height"`
			LatestBlockTime   time.Time `json:"latest_block_time"`
			CatchingUp        bool      `json:"catching_up"`
		} `json:"sync_info"`
		ValidatorInfo struct {
			Address string `json:"address"`
			PubKey  struct {
				Type  string `json:"type"`
				Value string `json:"value"`
			} `json:"pub_key"`
			VotingPower string `json:"voting_power"`
		} `json:"validator_info"`
	} `json:"result"`
}
