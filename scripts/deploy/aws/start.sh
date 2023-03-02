#!/usr/bin/env bash

RIGO_DATADIR="$HOME/.rigo"
RIGO_LOG="${RIGO_DATADIR}/log"
RIGO_BINFILE="$HOME/rigo-deploy/rigo"
RIGO_INITDIR="$HOME/rigo-deploy/.rigo"


if [ "$DEPLOYMENT_GROUP_NAME" == "rigo-dev-init-dg" ]; then
  #
  # Remove ~/.rigo and copy ~//rigo-deploy/.rigo to ~/.rigo
  #
  if [[ -d "$RIGO_DATADIR" ]]; then
    echo "Remove ${RIGO_DATADIR} ..."
    rm -rf ${RIGO_DATADIR}
  fi

  echo "Copy '${RIGO_INITFILES}' to '${RIGO_DATADIR}' ..."

  # copy ./.rigo to ~/.rigo
  cp -rf ${RIGO_INITFILES} ${RIGO_DATADIR}

elif [ "$DEPLOYMENT_GROUP_NAME" == "rigo-dev-reset-dg" ]; then
  #
  # Reset ~/.rigo. Remove blockchain data.
  #
  echo "Reset data in ${RIGO_DATADIR}..."
  ${RIGO_BINFILE} unsafe-reset-all --priv_validator_secret '1' # 1>${RIGO_LOG} 2>&1
fi

echo "Start rigo ..."
nohup ${RIGO_BINFILE} start --rpc.laddr 'tcp://0.0.0.0:26657' --rpc.cors_allowed_origins '*' --priv_validator_secret '1' 1>${RIGO_LOG} 2>&1 &