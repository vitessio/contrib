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

alias vtctlclient="vtctlclient --server=localhost:15999"
alias mysql="mysql -h 127.0.0.1 -P 15306 -u user"

vtctlclient MoveTables -- --source="rds" --all Create "vitess.railsApp"
vtctlclient MoveTables -- Progress "vitess.railsApp"
vtctlclient VDiff -- "vitess.railsApp"
vtctlclient MoveTables -- SwitchTraffic "vitess.railsApp"
vtctlclient GetRoutingRules rds
vtctlclient MoveTables -- -keep_data Complete "vitess.railsApp"

