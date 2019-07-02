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

// This is a sample vstream client. */

package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	binlogdatapb "vitess.io/vitess/go/vt/proto/binlogdata"
	logutilpb "vitess.io/vitess/go/vt/proto/logutil"
	topodatapb "vitess.io/vitess/go/vt/proto/topodata"
	_ "vitess.io/vitess/go/vt/vtctl/grpcvtctlclient"
	"vitess.io/vitess/go/vt/vtctl/vtctlclient"
	_ "vitess.io/vitess/go/vt/vtgate/grpcvtgateconn"
	"vitess.io/vitess/go/vt/vtgate/vtgateconn"
)

func main() {
	ctx := context.Background()
	vgtid, err := getPosition(ctx, "commerce", "0")
	if err != nil {
		log.Fatal(err)
	}
	filter := &binlogdatapb.Filter{
		Rules: []*binlogdatapb.Rule{{
			Match: "/.*/",
		}},
	}

	conn, err := vtgateconn.Dial(ctx, "localhost:15991")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	reader, err := conn.VStream(ctx, topodatapb.TabletType_MASTER, vgtid, filter)
	for {
		e, err := reader.Recv()
		switch err {
		case nil:
			fmt.Printf("%v\n", e)
		case io.EOF:
			fmt.Printf("stream ended\n")
		default:
			fmt.Printf("remote error: %v\n", err)
		}
	}
}

func getPosition(ctx context.Context, keyspace, shard string) (*binlogdatapb.VGtid, error) {
	results, err := execVtctl(ctx, []string{"ShardReplicationPositions", fmt.Sprintf("%s:%s", keyspace, shard)})
	if err != nil {
		return nil, err
	}
	// results contains multiple lines like this:
	// zone1-0000000100 commerce 0 master sougou-XPS:15100 sougou-XPS:17100 [] MySQL56/ee59be3a-9cf5-11e9-a6e0-9cb6d089e1c3:1-9 0
	// zone1-0000000101 commerce 0 replica sougou-XPS:15101 sougou-XPS:17101 [] MySQL56/ee59be3a-9cf5-11e9-a6e0-9cb6d089e1c3:1-9 0
	// zone1-0000000102 commerce 0 rdonly sougou-XPS:15102 sougou-XPS:17102 [] MySQL56/ee59be3a-9cf5-11e9-a6e0-9cb6d089e1c3:1-9 0
	// Just parse out one position.
	splits := strings.Split(results[0], " ")
	return &binlogdatapb.VGtid{
		ShardGtids: []*binlogdatapb.ShardGtid{{
			Keyspace: keyspace,
			Shard:    shard,
			Gtid:     splits[7],
		}},
	}, nil
}

func execVtctl(ctx context.Context, args []string) ([]string, error) {
	client, err := vtctlclient.New("localhost:15999")
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	stream, err := client.ExecuteVtctlCommand(ctx, args, 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("cannot execute remote command: %v", err)
	}

	var results []string
	for {
		e, err := stream.Recv()
		switch err {
		case nil:
			if e.Level == logutilpb.Level_CONSOLE {
				results = append(results, e.Value)
			}
		case io.EOF:
			return results, nil
		default:
			return nil, fmt.Errorf("remote error: %v", err)
		}
	}
	return results, nil
}
