#!/usr/bin/env bash

ARCANUS_DATADIR="$HOME/.rigo"
ARCANUS_LOG="${ARCANUS_DATADIR}/log"
ARCANUS_BINFILE="$HOME/rigo-deploy/rigo"
ARCANUS_INITDIR="$HOME/rigo-deploy/.rigo"


if [ "$DEPLOYMENT_GROUP_NAME" == "rigo-dev-init-dg" ]; then
  #
  # Remove ~/.rigo and copy ~//rigo-deploy/.rigo to ~/.rigo
  #
  if [[ -d "$ARCANUS_DATADIR" ]]; then
    echo "Remove ${ARCANUS_DATADIR} ..."
    rm -rf ${ARCANUS_DATADIR}
  fi

  echo "Copy '${ARCANUS_INITFILES}' to '${ARCANUS_DATADIR}' ..."

  # copy ./.rigo to ~/.rigo
  cp -rf ${ARCANUS_INITFILES} ${ARCANUS_DATADIR}

elif [ "$DEPLOYMENT_GROUP_NAME" == "rigo-dev-reset-dg" ]; then
  #
  # Reset ~/.rigo. Remove blockchain data.
  #
  echo "Reset data in ${ARCANUS_DATADIR}..."
  ${ARCANUS_BINFILE} unsafe-reset-all --priv_validator_secret '1' # 1>${ARCANUS_LOG} 2>&1
fi

echo "Start rigo ..."
nohup ${ARCANUS_BINFILE} start --rpc.laddr 'tcp://0.0.0.0:26657' --rpc.cors_allowed_origins '*' --priv_validator_secret '1' 1>${ARCANUS_LOG} 2>&1 &