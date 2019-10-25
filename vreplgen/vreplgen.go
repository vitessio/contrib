/*
Copyright 2019 The Vitess Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// This program generates a fully escaped command line for
// issuing a VReplicationExec command.
// Change the contents of the data structures in the main
// program to match your requirements and issue 'go run vreplgen.go'

package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"strings"

	"vitess.io/vitess/go/sqltypes"
	binlogdatapb "vitess.io/vitess/go/vt/proto/binlogdata"
)

var onDdl string

func init() {
	flag.StringVar(&onDdl, "on_ddl", "ignore", "Set on_ddl value for replication stream - ignore, stop, exec, exec_ignore")
	flag.Parse()
}

func main() {
	argOffset := 0
	if (len(os.Args) > 1 && strings.HasPrefix(os.Args[1], "-")) {
		argOffset = 2
	}

	if (len(os.Args) < (7+argOffset)) {
		fmt.Println("Usage: vreplgen [-on_ddl (ignore|stop|exec|exec_ignore)] <tablet_id> <src_keyspace> <src_shard> <dest_keyspace> <dest_table1> 'filter1' [<dest_table2> 'filter2']...")
		os.Exit(1)
	}

	vtctl := os.Getenv("VTCTLCLIENT")
	if (vtctl == "") {
	  vtctl = "vtctlclient -server localhost:15999"
	}
	tabletID := os.Args[1+argOffset]
	sourceKeyspace := os.Args[2+argOffset]
	sourceShard := os.Args[3+argOffset]
	destKeyspace := os.Args[4+argOffset]
	destDbName := "vt_" + destKeyspace
	var rules []*binlogdatapb.Rule
	for i := 5+argOffset; i < len(os.Args); i = i+2 {
		destTable := os.Args[i]
		destFilter := os.Args[i+1]
		rule := new(binlogdatapb.Rule)
		rule.Match = destTable
		rule.Filter = destFilter
		rules = append(rules, rule)
	}
	filter := &binlogdatapb.Filter{
		Rules: rules,
	}

	var onDdlAction binlogdatapb.OnDDLAction
	switch onDdl {
	case "ignore":
		onDdlAction = binlogdatapb.OnDDLAction_IGNORE
	case "stop":
		onDdlAction = binlogdatapb.OnDDLAction_STOP
	case "exec":
		onDdlAction = binlogdatapb.OnDDLAction_EXEC
	case "exec_ignore":
		onDdlAction = binlogdatapb.OnDDLAction_EXEC_IGNORE
	}

	bls := &binlogdatapb.BinlogSource{
		Keyspace: sourceKeyspace,
		Shard:    sourceShard,
		Filter:   filter,
		OnDdl:    onDdlAction,
	}
	val := sqltypes.NewVarBinary(fmt.Sprintf("%v", bls))
	var sqlEscaped bytes.Buffer
	val.EncodeSQL(&sqlEscaped)
	query := fmt.Sprintf("insert into _vt.vreplication "+
		"(db_name, source, pos, max_tps, max_replication_lag, tablet_types, time_updated, transaction_timestamp, state) values"+
		"('%s', %s, '', 9999, 9999, 'master', 0, 0, 'Running')", destDbName, sqlEscaped.String())

	fmt.Printf("%s VReplicationExec %s '%s'\n", vtctl, tabletID, strings.Replace(query, "'", "'\"'\"'", -1))
}
