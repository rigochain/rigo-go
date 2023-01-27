#!/usr/bin/env bash

# list only arcanus processes which is launched by start.sh
ps -e | grep 'arcanus start --home' | grep -v 'grep'