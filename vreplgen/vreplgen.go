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

var onDDL string

func init() {
	flag.StringVar(&onDDL, "on_ddl", "ignore", "Set on_ddl value for replication stream - ignore, stop, exec, exec_ignore")
}

func main() {
	flag.Parse()

	if len(os.Args) < 9 {
		fmt.Println("Usage: vreplgen [-on_ddl (ignore|stop|exec|exec_ignore)] <tablet_id> <src_keyspace> <src_shard> <dest_keyspace> <dest_table1> 'filter1' [<dest_table2> 'filter2']...")
		os.Exit(1)
	}

	vtctl := os.Getenv("VTCTLCLIENT")
	if vtctl == "" {
		vtctl = "vtctlclient -server localhost:15999"
	}

	// First, we process fixed positional arguments
	//   such as the intended target and source
	tabletID := os.Args[3]
	sourceKeyspace := os.Args[4]
	sourceShard := os.Args[5]
	destKeyspace := os.Args[6]
	destDbName := destKeyspace
	var rules []*binlogdatapb.Rule

	// Next, we iterate over all possible rules
	//  Note this can be a variable number!
	for i := 7; i < len(os.Args); i = i + 2 {
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

	onDDLAction := binlogdatapb.OnDDLAction(binlogdatapb.OnDDLAction_value[strings.ToUpper(onDDL)])

	bls := &binlogdatapb.BinlogSource{
		Keyspace: sourceKeyspace,
		Shard:    sourceShard,
		Filter:   filter,
		OnDdl:    onDDLAction,
	}
	val := sqltypes.NewVarBinary(fmt.Sprintf("%v", bls))
	var sqlEscaped bytes.Buffer
	val.EncodeSQL(&sqlEscaped)
	query := fmt.Sprintf("insert into _vt.vreplication "+
		"(db_name, source, pos, max_tps, max_replication_lag, tablet_types, time_updated, transaction_timestamp, state) values"+
		"('%s', %s, '', 9999, 9999, 'master', 0, 0, 'Running')", destDbName, sqlEscaped.String())

	fmt.Printf("%s VReplicationExec %s '%s'\n", vtctl, tabletID, strings.Replace(query, "'", "'\"'\"'", -1))
}
