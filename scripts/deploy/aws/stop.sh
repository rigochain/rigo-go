#!/usr/bin/env bash

PID=`ps -ef | grep 'arcanus' | grep -v grep | awk '{print $2}'`
if [[ -n "$PID" ]]; then
  echo "Stopping kms-go (PID:${PID}) ..."
  kill -15 ${PID}
fi