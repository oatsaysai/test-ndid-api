package ndid

import (
	"encoding/json"
	"fmt"

	"github.com/test-ndid-api/server/config"
	"github.com/test-ndid-api/server/core/crypto"
	"github.com/test-ndid-api/server/tendermint"
)

type InitNDIDParam struct {
	NodeID          string `json:"node_id"`
	PublicKey       string `json:"public_key"`
	MasterPublicKey string `json:"master_public_key"`
}

type EndInitParam struct{}

type RegisterNodeParam struct {
	NodeID          string  `json:"node_id"`
	PublicKey       string  `json:"public_key"`
	MasterPublicKey string  `json:"master_public_key"`
	NodeName        string  `json:"node_name"`
	Role            string  `json:"role"`
	MaxIal          float64 `json:"max_ial"`
	MaxAal          float64 `json:"max_aal"`
}

type SetNodeTokenParam struct {
	NodeID string  `json:"node_id"`
	Amount float64 `json:"amount"`
}

type AddServiceParam struct {
	ServiceID   string `json:"service_id"`
	ServiceName string `json:"service_name"`
}

type AddNamespaceParam struct {
	Namespace   string `json:"namespace"`
	Description string `json:"description"`
}

type RegisterServiceDestinationByNDIDParam struct {
	ServiceID string `json:"service_id"`
	NodeID    string `json:"node_id"`
}

type SetValidatorParam struct {
	PublicKey string `json:"public_key"`
	Power     int64  `json:"power"`
}

func InitNDID(param InitNDIDParam) {
	cfg := config.LoadConfiguration()
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	signature, nonce := crypto.CreateSignatureAndNonce("InitNDID", string(paramJSON), cfg.NodeID)
	tendermint.Transact([]byte("InitNDID"), paramJSON, []byte(nonce), signature, []byte(cfg.NodeID))
}

func EndInit(param EndInitParam) {
	cfg := config.LoadConfiguration()
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	signature, nonce := crypto.CreateSignatureAndNonce("EndInit", string(paramJSON), cfg.NodeID)
	tendermint.Transact([]byte("EndInit"), paramJSON, []byte(nonce), signature, []byte(cfg.NodeID))
}

func RegisterNode(param RegisterNodeParam) {
	cfg := config.LoadConfiguration()
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	signature, nonce := crypto.CreateSignatureAndNonce("RegisterNode", string(paramJSON), cfg.NodeID)
	tendermint.Transact([]byte("RegisterNode"), paramJSON, []byte(nonce), signature, []byte(cfg.NodeID))
}

func SetNodeToken(param SetNodeTokenParam) {
	cfg := config.LoadConfiguration()
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	signature, nonce := crypto.CreateSignatureAndNonce("SetNodeToken", string(paramJSON), cfg.NodeID)
	tendermint.Transact([]byte("SetNodeToken"), paramJSON, []byte(nonce), signature, []byte(cfg.NodeID))
}

func AddService(param AddServiceParam) {
	cfg := config.LoadConfiguration()
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	signature, nonce := crypto.CreateSignatureAndNonce("AddService", string(paramJSON), cfg.NodeID)
	tendermint.Transact([]byte("AddService"), paramJSON, []byte(nonce), signature, []byte(cfg.NodeID))
}

func AddNamespace(param AddNamespaceParam) {
	cfg := config.LoadConfiguration()
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	signature, nonce := crypto.CreateSignatureAndNonce("AddNamespace", string(paramJSON), cfg.NodeID)
	tendermint.Transact([]byte("AddNamespace"), paramJSON, []byte(nonce), signature, []byte(cfg.NodeID))
}

func RegisterServiceDestinationByNDID(param RegisterServiceDestinationByNDIDParam) {
	cfg := config.LoadConfiguration()
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	signature, nonce := crypto.CreateSignatureAndNonce("RegisterServiceDestinationByNDID", string(paramJSON), cfg.NodeID)
	tendermint.Transact([]byte("RegisterServiceDestinationByNDID"), paramJSON, []byte(nonce), signature, []byte(cfg.NodeID))
}

func SetValidator(param SetValidatorParam) {
	cfg := config.LoadConfiguration()
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	signature, nonce := crypto.CreateSignatureAndNonce("SetValidator", string(paramJSON), cfg.NodeID)
	tendermint.Transact([]byte("SetValidator"), paramJSON, []byte(nonce), signature, []byte(cfg.NodeID))
}
