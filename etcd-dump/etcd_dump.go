package main

import (
	"context"
	"flag"
	"fmt"
	_ "fmt"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/browser"
	"go.etcd.io/etcd/clientv3"
	"html"
	"io/ioutil"
	"strings"
	"time"
	"unicode"
	"vitess.io/vitess/go/vt/log"
	"vitess.io/vitess/go/vt/proto/topodata"
	"vitess.io/vitess/go/vt/proto/vschema"
)

var (
	server = flag.String("server", "127.0.0.1:2379", "Etcd API endpoint")
	outputFile = flag.String("out", "", "Stores html output into this file path")
	quiet = flag.Bool("quiet", false, "Do not open in browser (by default it opens the topo tree in a browser)")
	dialTimeout = 2 * time.Second
	requestTimeout = 10 * time.Second
	protoDecodeError = "Error decoding"
	messages = map[string]proto.Message {
		"RoutingRules": &vschema.RoutingRules{},
		"CellInfo":&topodata.CellInfo{},
		"Keyspace":&topodata.Keyspace{},
		"Shard":&topodata.Shard{},
		"SrvVSchema":&vschema.SrvVSchema{},
		"SrvKeyspace":&topodata.SrvKeyspace{},
		"ShardReplication":&topodata.ShardReplication{},
		"Tablet":&topodata.Tablet{},
		"VSchema":&vschema.Keyspace{},
	}

)

var css = `
<style>
/* Remove default bullets */
ul, #root {
  list-style-type: none;
}

/* Remove margins and padding from the parent ul */
#root {
  margin: 0;
  padding: 0;
}

/* Style the caret/arrow */
.caret {
  cursor: pointer;
  user-select: none; /* Prevent text selection */
}

/* Create the caret/arrow with a unicode, and style it */
.caret::before {
  content: "\25B6";
  color: black;
  display: inline-block;
  margin-right: 6px;
}

/* Rotate the caret/arrow icon when clicked on (using JavaScript) */
.caret-down::before {
  transform: rotate(90deg);
}

/* Hide the nested list */
.nested {
  display: none;
}

/* Show the nested list when the user clicks on the caret/arrow (with JavaScript) */
.active {
  display: block;
}
</style>`

var js = `
<script>
var toggler = document.getElementsByClassName("caret");
var i;

for (i = 0; i < toggler.length; i++) {
  toggler[i].addEventListener("click", function() {
    this.parentElement.querySelector(".nested").classList.toggle("active");
    this.classList.toggle("caret-down");
  });
}

function all(open) {
	var elems = document.getElementsByClassName("nested")	
	for (i=0; i < elems.length; i++) 
			open ? elems[i].classList.add("active") : elems[i].classList.remove("active")
	elems = document.getElementsByClassName("caret")	
	for (i=0; i < elems.length; i++) 
			open ? elems[i].classList.add("caret-down") : elems[i].classList.remove("caret-down")
}
</script>
`

func decodeProtoBuf(objectType string, objectBuf []byte) string {
	msg, found := messages[objectType]
	if !found {
		log.Errorf("No mapping for %v", objectType)
		return string(objectBuf)
	}
	err := proto.Unmarshal(objectBuf, msg)
	if err != nil {
		log.Errorf("Could not unmarshal %v:%v", objectType, objectBuf)
		return string(objectBuf)
	}
	json, err := new(jsonpb.Marshaler).MarshalToString(msg)
	if err != nil {
		log.Errorf("JSON error %v %v", err, json)
		return msg.String()
	}
	return json
}

func store(tree *Tree, key string, value string ) {
	paths := strings.Split(key, "/")
	typeString := paths[len(paths)-1]
	if unicode.IsUpper(rune(typeString[0])) {
		value = decodeProtoBuf(typeString, []byte(value))
		value = html.EscapeString(value)
	}

	paths = append(paths, value)
	paths = paths[1:]
	parent := tree.root
	for _, path := range paths {
		parent = tree.addChild(parent, path)
	}
}

func loadKeys() *Tree{
	ctx, _ := context.WithTimeout(context.Background(), requestTimeout)
	client, err := clientv3.New(clientv3.Config{
		DialTimeout: dialTimeout,
		Endpoints: []string{*server},
	})

	if err != nil {
		fmt.Errorf("Cannot connect to server at %s : %v\n", *server, err)
		return nil
	}
	defer client.Close()

	tree := NewTree("Vitess Etcd")
	tree.setRoot("root")

	kv := clientv3.NewKV(client)

	gr, err := kv.Get(ctx, "", clientv3.WithPrefix(),
					clientv3.WithSort(clientv3.SortByKey, clientv3.SortAscend))
	if err != nil {
		return nil
	}
	for _, item := range gr.Kvs {
		key := string(item.Key)
		store(tree, key, string(item.Value))
	}
	return tree
}

type Item struct {
	level int
	name string
	next int
}

func flatten(tree *Tree) ([]Item, int) {
	var list []Item
	maxLevel := 0 //maxlevel not required for a tree representation, but useful for displaying in a a table for example
	callback := func (name string, level int) {
		list = append(list, Item{
			level: level,
			name:  name,
		})
		if level > maxLevel {
			maxLevel = level
		}
	}
	tree.root.traverse(0, callback)
	for idx, item := range list {
		if idx != len(list) -1 {
			list[idx].next = list[idx+1].level - item.level
			if list[idx].next < -1 {
				list[idx].next = -1
			}
		}
	}
	return list, maxLevel
}

func toHtml(list []Item) string {
	html := "<html><body><h1>etcd: dump of all keys</h1>\n"
	html += css
	html += "<a href='javascript:all(true)' style='text-decoration:none'>Expand All</a>&nbsp;&nbsp;&nbsp;&nbsp;"
	html += "<a href='javascript:all(false)' style='text-decoration:none'>Collapse All</a>\n"
	html += "<br><br><ul id='root'>\n"
	closeTags := make([]string,0)
	push := func (s string) {
		closeTags = append(closeTags, s)
	}
	pop := func () string {
		var s string
		s, closeTags = closeTags[len(closeTags)-1], closeTags[:len(closeTags)-1]
		return s
	}
	push("</ul>")
	for _, item := range list {
		switch item.next {
		case 1:
			html += "<li><span class='caret'>" + item.name + "</span>\n"
			html += "<ul class='nested'>\n"
			push("</ul></li>")
			break
		case 0:
			html += "<li>"+item.name+"</li>"
			break
		case -1:
			html += "<li>"+item.name+"</li>"
			html += pop()
			break
		default:
			panic(fmt.Sprintf("Invalid item.next %d", item.next))

		}
	}
	return html + js
}

func main() {
	flag.Parse()
	tree := loadKeys()
	if tree == nil {
		fmt.Errorf("no keys found")
		return
	}
	list, _ := flatten(tree)
	html := toHtml(list)
	if *outputFile != "" {
		_ = ioutil.WriteFile(*outputFile, []byte(html), 0644)
	}
	if !*quiet {
		browser.OpenReader(strings.NewReader(html))
	}
}
