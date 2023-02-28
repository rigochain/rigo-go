#!/usr/bin/env bash

kill -15 `ps -ef | grep rigo | grep -v 'grep' | awk '{print $2}'`