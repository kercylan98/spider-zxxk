#!/bin/bash

mkdir -p ./data
rm -rf ./save.signal
rm -rf ./stop.signal
nohup /opt/xkw/spider-zxxk > xkw.log 2>&1 &

