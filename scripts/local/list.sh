#!/usr/bin/env bash

# list only rigo processes which is launched by start.sh
ps -e | grep 'rigo start --home' | grep -v 'grep'