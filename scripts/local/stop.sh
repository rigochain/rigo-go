#!/usr/bin/env bash

# kill only arcanus processes which is launched by start.sh
PID=`ps -e | grep 'arcanus_localnet_' | grep -v 'grep' | awk '{print $1}'`
for p in $PID
do
  echo "Kill $p..."
  kill -15 $p
done