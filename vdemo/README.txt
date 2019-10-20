Fill data
kmysql <data.sql

Product Materialized view
kvtctl Materialize -create_table -is_reference product.product customer.product
kvtctl Expose --auto_route customer.product

Merchant Orders
kvtctl Materialize -create_table -primary_vindex=mname:md5 customer.orders merchant.orders
kvtctl Expose --auto_route merchant.orders

Sales schema
kvtctl ApplySchema -sql='create table sales(pid int, kount int, amount int, primary key(pid))' product

Materialize sales
kvtctl Materialize 'select pid, count(*) as kount, sum(price) as amount from customer.orders group by pid' product.sales
kvtctl Expose product.sales

Migrate orders
kvtctl MigrateReads -tablet_type=rdonly merchant.orders
kvtctl MigrateReads -tablet_type=replica merchant.orders
kvtctl MigrateWrites merchant.orders
