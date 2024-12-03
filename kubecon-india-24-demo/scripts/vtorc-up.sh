#!/bin/bash

source "./env.sh"

log_dir="${VTDATAROOT}/tmp"
port=16000

echo "Starting vtorc..."
vtorc \
  $TOPOLOGY_FLAGS \
  --logtostderr \
  --alsologtostderr \
  --port $port \
  --instance-poll-time '1s' \
  --topo-information-refresh-duration '3s' \
  > "${log_dir}/vtorc.out" 2>&1 &

vtorc_pid=$!
echo ${vtorc_pid} > "${log_dir}/vtorc.pid"

echo "\
vtorc is running!
  - UI: http://localhost:${port}
  - Logs: ${log_dir}/vtorc.out
  - PID: ${vtorc_pid}
"
