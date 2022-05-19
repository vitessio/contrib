#!/bin/bash

# Copyright 2022 The Vitess Authors.

# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at

#     http://www.apache.org/licenses/LICENSE-2.0

# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.


shopt -s expand_aliases
alias vtctlclient="vtctlclient --server=localhost:15999"
alias mysql="mysql -h 127.0.0.1 -P 15306 -u user"

# checkPodStatusWithTimeout:
# $1: regex used to match pod names
# $2: number of pods to match (default: 1)
function checkPodStatusWithTimeout() {
  regex=$1
  nb=$2

  # Number of pods to match defaults to one
  if [ -z "$nb" ]; then
    nb=1
  fi

  # We use this for loop instead of `kubectl wait` because we don't have access to the full pod name
  # and `kubectl wait` does not support regex to match resource name.
  for i in {1..1200} ; do
    out=$(kubectl get pods)
    echo "$out" | grep -E "$regex" | wc -l | grep "$nb" > /dev/null 2>&1
    if [ $? -eq 0 ]; then
      echo "$regex found"
      return
    fi
    sleep 1
  done
  echo -e "ERROR: checkPodStatusWithTimeout timeout to find pod matching:\ngot:\n$out\nfor regex: $regex"
  exit 1
}
