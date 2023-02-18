#!/bin/bash
trap "rm cacheserver;kill 0" EXIT

go build -o cacheserver
./cacheserver -port=8001 &
./cacheserver -port=8002 &
./cacheserver -port=8003 &
./cacheserver -port=9999 -api=1 &

sleep 2 
echo ">>> start test <<<"
curl "http://localhost:9999/api?key=Tom" &
curl "http://localhost:9999/api?key=Tom" &
curl "http://localhost:9999/api?key=Tom" &

wait