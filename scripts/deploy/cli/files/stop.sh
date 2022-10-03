#!/usr/bin/env bash

kill -15 `ps -ef | grep arcanus | grep -v 'grep' | awk '{print $2}'`