#!/bin/bash

trap killgroup SIGINT

killgroup(){
  echo killing...
  kill 0
}

rm -rf api_*

rm -rf log
rm -rf http_log

TENDERMINT_IP=127.0.0.1 \
TENDERMINT_PORT=45000 \
NODE_ID=NDID1 \
ROLE=NDID \
go run server/dev_init/dev_init.go

TENDERMINT_IP=127.0.0.1 \
TENDERMINT_PORT=45001 \
SERVER_PORT=8100 \
NODE_ID=rp1 \
ROLE=rp \
nohup go run server/main.go > api_rp1.log &

TENDERMINT_IP=127.0.0.1 \
TENDERMINT_PORT=45002 \
SERVER_PORT=8101 \
NODE_ID=rp2 \
ROLE=rp \
nohup go run server/main.go > api_rp2.log &

TENDERMINT_IP=127.0.0.1 \
TENDERMINT_PORT=45003 \
SERVER_PORT=8103 \
NODE_ID=rp4 \
ROLE=rp \
nohup go run server/main.go > api_rp4.log &

TENDERMINT_IP=127.0.0.1 \
TENDERMINT_PORT=45000 \
SERVER_PORT=8102 \
NODE_ID=rp3 \
ROLE=rp \
go run server/main.go

wait