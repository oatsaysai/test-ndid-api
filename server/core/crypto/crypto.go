package crypto

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"github.com/tendermint/tendermint/libs/common"
	"github.com/test-ndid-api/server/config"
)

func CreateSignatureAndNonce(fnName, param, nodeID string) ([]byte, string) {
	cfg := config.LoadConfiguration()
	var privKey *rsa.PrivateKey
	privKeyFile, err := ioutil.ReadFile("server/dev_key/" + cfg.Role + "/" + cfg.NodeID)
	if err != nil {
		log.Fatal(err)
	}
	privKey = getPrivateKeyFromString(string(privKeyFile))
	nonce := base64.StdEncoding.EncodeToString([]byte(common.RandStr(12)))
	tempPSSmessage := append([]byte(fnName), param...)
	tempPSSmessage = append(tempPSSmessage, []byte(nonce)...)
	PSSmessage := []byte(base64.StdEncoding.EncodeToString(tempPSSmessage))
	newhash := crypto.SHA256
	pssh := newhash.New()
	pssh.Write(PSSmessage)
	hashed := pssh.Sum(nil)
	signature, err := rsa.SignPKCS1v15(rand.Reader, privKey, newhash, hashed)
	if err != nil {
		fmt.Println("error:", err)
	}
	return signature, nonce
}

func getPrivateKeyFromString(privK string) *rsa.PrivateKey {
	privK = strings.Replace(privK, "\t", "", -1)
	block, _ := pem.Decode([]byte(privK))
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		fmt.Println(err.Error())
	}
	return privateKey
}
