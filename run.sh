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
TENDERMINT_PORT=45000 \
SERVER_PORT=8102 \
NODE_ID=rp3 \
ROLE=rp \
go run server/main.go



# go run dev_init/dev_init.go

# SERVER_PORT=8080 \
# NODE_ID=ndid1 \
# ROLE=ndid \
# nohup go run server.go > api_ndid1.log &

# SERVER_PORT=8101 \
# NODE_ID=idp2 \
# ROLE=idp \
# nohup go run server.go > api_idp2.log &

# SERVER_PORT=8200 \
# NODE_ID=rp1 \
# ROLE=rp \
# nohup go run server.go > api_rp1.log &

# SERVER_PORT=8300 \
# NODE_ID=as1 \
# ROLE=as \
# nohup go run server.go > api_as1.log &

# SERVER_PORT=8301 \
# NODE_ID=as2 \
# ROLE=as \
# nohup go run server.go > api_as2.log &

# SERVER_PORT=8100 \
# NODE_ID=idp1 \
# ROLE=idp \
# go run server.go

wait