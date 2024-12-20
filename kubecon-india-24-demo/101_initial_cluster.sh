#!/bin/bash

# Copyright 2019 The Vitess Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# this script brings up zookeeper and all the vitess components
# required for a single shard deployment.

source "./env.sh"

# This is done here as a means to support testing the experimental
# custom sidecar database name work in a wide variety of scenarios
# as the local examples are used to test many features locally.
# This is NOT here to indicate that you should normally use a
# non-default (_vt) value or that it is somehow a best practice
# to do so. In production, you should ONLY use a non-default
# sidecar database name when it's truly needed.
SIDECAR_DB_NAME=${SIDECAR_DB_NAME:-"_vt"}

# start topo server
CELL=zone1 ./scripts/etcd-up.sh

# start vtctld
CELL=zone1 ./scripts/vtctld-up.sh

vtctldclient CreateKeyspace --sidecar-db-name="${SIDECAR_DB_NAME}" --durability-policy=semi_sync commerce || fail "Failed to create and configure the commerce keyspace"
vtctldclient CreateKeyspace --sidecar-db-name="${SIDECAR_DB_NAME}" --durability-policy=semi_sync unsharded || fail "Failed to create and configure the unsharded keyspace"

# start mysqlctls for keyspace commerce
# because MySQL takes time to start, we do this in parallel
for i in 100 101 102 200 201 202 300 301 302; do
	CELL=zone1 TABLET_UID=$i ./scripts/mysqlctl-up.sh &
done

# without a sleep, we can have below echo happen before the echo of mysqlctl-up.sh
sleep 2
echo "Waiting for mysqlctls to start..."
wait
echo "mysqlctls are running!"

# start vttablets for keyspaces
for i in 100 101 102; do
	CELL=zone1 KEYSPACE=unsharded TABLET_UID=$i ./scripts/vttablet-up.sh
done

for i in 200 201 202; do
	SHARD=80- CELL=zone1 KEYSPACE=commerce TABLET_UID=$i ./scripts/vttablet-up.sh
done

for i in 300 301 302; do
	SHARD=-80 CELL=zone1 KEYSPACE=commerce TABLET_UID=$i ./scripts/vttablet-up.sh
done

# start vtorc
./scripts/vtorc-up.sh

# Wait for all the tablets to be up and registered in the topology server
# and for a primary tablet to be elected in the shard and become healthy/serving.
wait_for_healthy_shard commerce 80- || exit 1
wait_for_healthy_shard commerce -80 || exit 1
wait_for_healthy_shard unsharded 0 || exit 1

# create the schema
vtctldclient ApplySchema --sql-file create_commerce_schema.sql commerce
vtctldclient ApplySchema --sql-file create_unsharded_schema.sql unsharded

# create the vschema
vtctldclient ApplyVSchema --vschema-file vschema_unsharded.json unsharded
vtctldclient ApplyVSchema --vschema-file vschema_commerce.json commerce

# start vtgate
CELL=zone1 ./scripts/vtgate-up.sh

# start vtadmin
./scripts/vtadmin-up.sh

