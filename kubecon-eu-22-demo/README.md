# Kubecon EU 2022 Demo

These are the steps to follow to run the demo presented at KubeCon EU 2022 at Scaling Databases with Vitess.
The aim of the demo is to migrate a running rails app with RDS to use Vitess instead.

Prerequisite

1. Have a RDS database running and export the following environment variables - `RDS_PASSWORD`, `RDS_DBNAME`, `RDS_HOST`, `RDS_PORT`, `RDS_USER`.
2. Have Kind installed.


Instructions of the Demo
1. Connect to RDS using `mysql -u $RDS_USER -p$RDS_PASSWORD -h$RDS_HOST -P$RDS_PORT` and run the following commands -
   1. ``grant SYSTEM_VARIABLES_ADMIN on *.* to `admin`@`%` ``
   2. `call mysql.rds_set_configuration('binlog retention hours', 24)`
2. You can now start the rails app. Run the following commands in the `rails` directory - 
   1. `spring stop`
   2. `rails db:migrate`
   3. `rails server`
   4. You can now go ahead and spawn the rails client in the web browser at `localhost:3000`
      1. The client will try to insert 4 rows in the to the rds database every second and record the latency averaged over 5 seconds.
3. You can now start the Vitess operator in kind - `./vtop/start.sh`
4. The next step is to switch the traffic to go to rds, but through vitess. To do that, go ahead and change the rails configuration in `rails/config/database.yml` connect to `127.0.0.1` on port `15306`, `rds` database, using the user `user`, with no password.
5. We can now start the MoveTables workflow to move the data from RDS to the local Vitess cluster. The commands for this live in `./vtop/moveToVitess.sh`.
6. Before we complete the workflow, we need to change the rds config again to connect to the Vitess keyspace directly.
7. Complete and the workflow and voila, all the data has been moved from RDS to Vitess!
 
