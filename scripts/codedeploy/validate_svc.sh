#!/bin/bash

# set service var
service=$(hostname | cut -d "-" -f3 | tr -cd '[:alpha:]')

echo "$service"

case $service in
  validator*)
    rs=$(curl -f 127.0.0.1:26657/abci_info)
    if [ "$rs" -eq 0 ]; then
      echo "$rs test success"
    else
      exit 1
    fi
    ;;
  sentry)
    rs=$(curl -f 127.0.0.1:26657/abci_info)
    if [ "$rs" -eq 0 ]; then
      echo "$rs test success"
    else
      exit 1
    fi
    ;;
  rpc)
    rs=$(curl -f 127.0.0.1:26657/abci_info)
    if [ "$rs" -eq 0 ]; then
      echo "$rs test success"
    else
      exit 1
    fi
    ;;
esac
