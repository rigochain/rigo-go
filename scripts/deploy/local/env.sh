#!/usr/bin/env bash

HOSTOS="linux"
if [[ "$OSTYPE" == "darwin"* ]]; then
  HOSTOS="darwin"
fi
echo "Your OS is $HOSTOS"

EXE="../../../build/${HOSTOS}/arcanus"

echo ${1}
NODE_HOMES=()
IDX=0
while [ $IDX -lt ${1} ]
do
  NODE_HOMES[$IDX]="$HOME/arcanus_localnet_$IDX"
  echo ${NODE_HOMES[$IDX]}
  IDX=$(($IDX + 1))
done
