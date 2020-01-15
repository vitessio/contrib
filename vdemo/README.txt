Fill data
kmysql <data.sql

Product Materialized view
kvtctl ApplyVSchema -vschema="$(cat new_customer.json)" customer
kvtctl Materialize "$(cat product_mat.json)"

Merchant Orders
kvtctl ApplyVSchema -vschema="$(cat new_merchant.json)" merchant
kvtctl Materialize "$(cat orders_mat.json)"

Materialize sales
kvtctl ApplyVSchema -vschema="$(cat new_product.json)" product
kvtctl Materialize "$(cat sales_mat.json)"
