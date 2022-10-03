#!/usr/bin/env bash

ARCANUS_DATADIR="$HOME/.arcanus"
ARCANUS_LOG="${ARCANUS_DATADIR}/log"
ARCANUS_BIN="$HOME/bin/arcanus"

if [ "$DEPLOYMENT_GROUP_NAME" == "arcanus-dev-init-dg" ]; then

  if [[ -d "$ARCANUS_DATADIR" ]]; then
    echo "Remove ${ARCANUS_DATADIR} ..."
    rm -rf ${ARCANUS_DATADIR}
  fi

  echo "Initialize ${ARCANUS_DATADIR} ..."
  ${ARCANUS_BIN} init --chain_id demonet --priv_validator_secret '1' # 1>${ARCANUS_LOG} 2>&1

elif [ "$DEPLOYMENT_GROUP_NAME" == "arcanus-dev-reset-dg" ]; then
  echo "Reset data in ${ARCANUS_DATADIR}..."
  ${ARCANUS_BIN} unsafe-reset-all --priv_validator_secret '1' # 1>${ARCANUS_LOG} 2>&1
fi

echo "Start arcanus ..."
nohup ${ARCANUS_BIN} start --rpc.laddr 'tcp://0.0.0.0:26657' --rpc.cors_allowed_origins '*' --priv_validator_secret '1' 1>${ARCANUS_LOG} 2>&1 &