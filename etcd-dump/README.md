# Etcd Dump

Get a list of all the keys stored by Vitess in Etcd for debugging and understanding how Vitess uses topo.

## Getting Started

To create a dump of all the keys in Vitess' etcd topology run ```go build``` followed by ```./etcd_dump``` or 
```go run etcd_dump.go tree.go``` 

### Options

```
-file string
    	Stores html output into this file path
-quiet
    	Do not open in browser (by default it opens the topo tree in a browser)
-server string
    	Etcd API endpoint (default "127.0.0.1:2379")
```

## Future Work

1. Currently define the mapping of the protobuf to a Vitess internal type in this tool. We should ideally use 
existing vitess libraries. However etcd V2 does not seem to have a simple way to get all keys without making recursive 
calls to etcd2. If I use etcd V3 it seems to clash with the Vitess libraries since vitess currently uses V2.

2. Test for ZooKeeper and Consul and add support for them if required.