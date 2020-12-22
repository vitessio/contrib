A golang CLI utility to generate vtctlclient commands to add vreplication
rules:

```
Usage: vreplgen [-on_ddl (ignore|stop|exec|exec_ignore)] <tablet_id> <src_keyspace> <src_shard> <dest_keyspace> <dest_table1> 'filter1' [<dest_table2> 'filter2']...
```

E.g.:

```
./vreplgen cell-0000000001 main -80 main_copy transactionhistory 'select * from transactionhistory where in_keyrange(merchant_id, "hash", "80-")'
```

The utility also supports multiple table filters, which allows multiple tables
to be specified in a single vreplication stream (good for if you have
a lot of tables you want to process via vreplication).  E.g.:

```
./vreplgen cell-0000000001 main -80 main_copy transactionhistory 'select * from transactionhistory where in_keyrange(merchant_id, "hash", "80-")' transactionhistory2 'select * from transactionhistory2 where in_keyrange(merchant_id, "hash", "-80")'
```

An important thing to note is that a single vreplication stream cannot use 
the same source table in the same stream.  The utility will not prevent
you from doing this, however.

`vreplgen` assumes you are running vtctld on localhost port 15999.  If not,
you can set your VTCTLCLIENT environment variable to the vtctlclient command
you want `vreplgen` to generate, e.g.:

```
export VTCTLCLIENT="vtctlclient -server vtctld:15999"
```

Lastly, you can control the on_ddl flag using vreplgen. The default if you
do not specify the `-on_ddl` option is `ignore`, but you can specify:

  * `-on_ddl ignore`
  * `-on_ddl stop`
  * `-on_ddl exec`
  * `-on_ddl exec_ignore`

depending on how you want your DDL to be handled in your replication streams.
See the main vreplication documentation for more details.
