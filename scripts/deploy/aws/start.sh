#!/usr/bin/env bash

ARCANUS_DATADIR="$HOME/.arcanus"
ARCANUS_LOG="${ARCANUS_DATADIR}/log"
ARCANUS_BINFILE="$HOME/arcanus-deploy/arcanus"
ARCANUS_INITDIR="$HOME/arcanus-deploy/.arcanus"


if [ "$DEPLOYMENT_GROUP_NAME" == "arcanus-dev-init-dg" ]; then
  #
  # Remove ~/.arcanus and copy ~//arcanus-deploy/.arcanus to ~/.arcanus
  #
  if [[ -d "$ARCANUS_DATADIR" ]]; then
    echo "Remove ${ARCANUS_DATADIR} ..."
    rm -rf ${ARCANUS_DATADIR}
  fi

  echo "Copy '${ARCANUS_INITFILES}' to '${ARCANUS_DATADIR}' ..."

  # copy ./.arcanus to ~/.arcanus
  cp -rf ${ARCANUS_INITFILES} ${ARCANUS_DATADIR}

elif [ "$DEPLOYMENT_GROUP_NAME" == "arcanus-dev-reset-dg" ]; then
  #
  # Reset ~/.arcanus. Remove blockchain data.
  #
  echo "Reset data in ${ARCANUS_DATADIR}..."
  ${ARCANUS_BINFILE} unsafe-reset-all --priv_validator_secret '1' # 1>${ARCANUS_LOG} 2>&1
fi

echo "Start arcanus ..."
nohup ${ARCANUS_BINFILE} start --rpc.laddr 'tcp://0.0.0.0:26657' --rpc.cors_allowed_origins '*' --priv_validator_secret '1' 1>${ARCANUS_LOG} 2>&1 &