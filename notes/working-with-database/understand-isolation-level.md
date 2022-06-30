{%hackmd BJrTq20hE %}
# Understand isolation levels & read phenomena in MySQL & PostgreSQL via examples
###### tags: `simplebank`

Today we will explore deeply how each level of isolation work in MySQL and Postgres by running some concrete SQL queries.

We will also learn how each `isolation level` prevents `read phenomena` such as
- dirty read
- non-repeatable read
- phantom read
- serialization anomaly.

# Transaction isolation and read phenomena
## ACID property
![](https://i.imgur.com/cD66Bnb.png)

Isolation is one of the four property of a database transaction, where at its highest level, a perfect isolation ensures that all concurrent transactions will not affect each other.

There are several ways that a transaction can be ***interfered*** by other transactions that runs simultaneously with it. This interference will cause something we called `read phenomenon`.

## 4 read phenomena
![](https://i.imgur.com/GP1B1WX.png)

### dirty read
It happens when a transaction reads data written by other concurrent transaction that has not been committed yet.

We don’t know if that other transaction will eventually be committed or rolled back. So we might end up using incorrect data in case rollback occurs.

### non-repeatable read
When a transaction reads the same record twice and see different values, because the row has been modified by other transaction that was committed after the first read.

### Phantom read
In this case, the same query is re-executed, but a different set of rows is returned, due to some changes made by other recently-committed transactions, such as inserting new rows or deleting existing rows which happen to satisfy the search condition of current transaction’s query.

### serialization anomaly
It’s when the result of a group of concurrent committed transactions could not be achieved if we try to run them sequentially in any order without overlapping each other.

## 4 isolation levels
Now in order to deal with these phenomena, 4 standard isolation levels were defined by the American National Standard Institute or ANSI.

![](https://i.imgur.com/BT35Ohh.png)
### read uncommitted
Transactions in this level can see data written by other uncommitted transactions, thus allowing dirty read phenomenon to happen.

### read committed
where transactions can only see data that has been committed by other transactions. Because of this, dirty read is no longer possible.

### repeatable read
It ensures that the same select query will always return the same result, no matter how many times it is executed, even if some other concurrent transactions have committed new changes that satisfy the query.

### serializable
Concurrent transactions running in this level are guaranteed to be able to yield the same result as if they’re executed sequentially in some order, one after another without overlapping.

# Relationship between isolation levels and read phenomena

## Isolation levels in Postgres

![](https://i.imgur.com/B1GHmsl.png)
## Summary about relationship betwen isolation levels and read phenomena
![](https://i.imgur.com/Nb5dAAR.png)
The isolation levels in Postgres produces quite similar result. However, there are still some major differences.

First, the read uncommitted isolation level behaves the same as read committed. So basically Postgres only has 3 isolation levels instead of 4 as in MySQL.
![](https://i.imgur.com/0DKafpl.png)

## Keep in mind
![](https://i.imgur.com/S09vr5Q.png)
The most important thing you should keep in mind when using high isolation level is that there might be some errors, timeout, or even deadlock. Thus, we should carefully implement a retry mechanism for our transactions.

Also, each database engine might implement isolation level differently. So make sure that you have read its documentation carefully, and tried it on your own first before jumping into coding.