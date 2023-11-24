#!/bin/bash

sleep 10
# set service var
service=$(hostname | cut -d "-" -f3 | tr -cd '[:alpha:]')

echo "$service"

case $service in
  validator*)
    curl -f 127.0.0.1:26657/abci_info
    if [ $? == 0 ]; then
      echo "validate test successed"
      echo "return value is $?"
    else
      echo "validate test failed"
      echo "return value is $?"
      exit 1
    fi
    ;;
  sentry)
    curl -f 127.0.0.1:26657/abci_info
    if [ $? == 0 ]; then
      echo "validate test successed"
      echo "return value is $?"
    else
      echo "validate test failed"
      echo "return value is $?"
      exit 1
    fi
    ;;
  rpc)
    curl -f 127.0.0.1:26657/abci_info
    if [ $? == 0 ]; then
      echo "validate test successed"
      echo "return value is $?"
    else
      echo "validate test failed"
      echo "return value is $?"
      exit 1
    fi
    ;;
esac
