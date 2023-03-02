#!/usr/bin/env bash

HOSTOS="linux"
if [[ "$OSTYPE" == "darwin"* ]]; then
  HOSTOS="darwin"
fi
echo "Your OS is $HOSTOS"

EXE="$GOPATH/src/github.com/rigochain/rigo-go/build/${HOSTOS}/rigo"

echo ${1}
NODE_HOMES=()
IDX=0
while [ $IDX -lt ${1} ]
do
  NODE_HOMES[$IDX]="$HOME/rigo_localnet_$IDX"
  echo ${NODE_HOMES[$IDX]}
  IDX=$(($IDX + 1))
done
