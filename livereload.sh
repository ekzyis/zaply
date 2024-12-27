#!/usr/bin/env bash

PID=$(pidof zaply)
DIRS="env/ lightning/ lnurl/ pages/ server/"

set -e

echo ":: remote port forwarding for zap-dev.ekzy.is ::"
ssh -fnNR 5555:localhost:4444 zap-dev.ekzy.is
echo

function restart_server() {
  set +e
  [[ -z "$PID" ]] || kill -15 $PID
  ENV=development make build -B
  set -e
  ./zaply 2>&1 &
  PID=$(pidof zaply)
}

function restart() {
  restart_server
  # give server time start listening for connections
  sleep 1
  date +%s.%N > public/__livereload
}

function cleanup() {
    rm -f public/__livereload
    [[ -z "$PID" ]] || kill -15 $PID
}
trap cleanup EXIT

restart

while inotifywait -r -e modify $DIRS; do
  restart
done
