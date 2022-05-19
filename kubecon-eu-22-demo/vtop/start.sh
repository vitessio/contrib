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

source ./vtop/utils.sh

cd vtop
# call this file after specifying the following environment variables
# RDS_DBNAME
# RDS_HOST
# RDS_PASSWORD
# RDS_PORT
# RDS_USER

kind create cluster --wait 30s --name kind
killall kubectl
kubectl apply -f operator.yaml
checkPodStatusWithTimeout "vitess-operator(.*)1/1(.*)Running(.*)"

cat ./101_initial_cluster_config.yaml | sed "s,<RDS_USER>,$RDS_USER," | sed "s,<RDS_PASSWORD>,$RDS_PASSWORD," | sed "s,<RDS_DBNAME>,$RDS_DBNAME," | sed "s,<RDS_HOST>,$RDS_HOST," | sed "s,<RDS_PORT>,$RDS_PORT," > ./101_initial_cluster.yaml

kubectl apply -f 101_initial_cluster.yaml
checkPodStatusWithTimeout "demo-zone1-vtctld(.*)1/1(.*)Running(.*)"
checkPodStatusWithTimeout "demo-zone1-vtgate(.*)1/1(.*)Running(.*)"
checkPodStatusWithTimeout "demo-etcd(.*)1/1(.*)Running(.*)" 3
checkPodStatusWithTimeout "demo-vttablet-zone1(.*)3/3(.*)Running(.*)" 2

sleep 10
./pf.sh > /dev/null 2>&1 &
sleep 5

rdsTablet=$(vtctlclient ListAllTablets -- -keyspace="rds" | grep -o -E "zone1-[0-9]*")
vtctlclient TabletExternallyReparented "$rdsTablet"
