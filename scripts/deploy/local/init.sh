#!/usr/bin/env bash

. ./env.sh


# init
idx=0
for home in "${NODE_HOMES[@]}"; do
  echo ""
  read -p "initialize arcanus at ${home} ?" a
  echo ""

  rm -rf ${home}
  $EXE init --home "${home}" --chain_id localnet --priv_validator_secret 1
  if [[ "$idx" -gt "0" ]]; then
    echo "copy genesis file... ${NODE_HOMES[0]}/config/genesis.json  to $home/config/genesis.json"
    cp -f ${NODE_HOMES[0]}/config/genesis.json $home/config/genesis.json
  fi
  idx=$((idx+1))
done