{%hackmd BJrTq20hE %}
# DB transaction lock & How to handle deadlock
###### tags: `simplebank`

Chapter 7 of [Backend Master Class [Golang + PostgreSQL + Kubernetes]](/3xKIijmxQJqv0z56ifCMeQ)

So in this lecture we’re gonna implement this feature to learn more about database locking and how to handle a deadlock situation.

## Test Driven Development

- check account
- check balance

## Query without lock
let’s start the psql console in 2 different terminal tabs and run 2 parallel transactions.

let’s run a normal SELECT query to get the account record with ID = 1.
```sql
SELECT * FROM accounts WHERE id = 1;
```
As you can see, the same account record is returned immediately without being blocked.

## Query with lock
This time, we will add FOR UPDATE clause at the end of the SELECT statement.
```sql
SELECT * FROM accounts WHERE id = 1 FOR UPDATE;
```
Now the first transaction still gets the record immediately. But when we run this query on the second transaction.

It is blocked and has to wait for the first transaction to COMMIT or ROLLBACK.

Let’s go back to that transaction and update the account balance to 500:

```sql
UPDATE accounts SET balance = 500 WHERE id = 1;
```

After this update, the second transaction is still blocked. However, as soon as we COMMIT the first transaction:

We can see that the second transaction is unblocked right away, and it gets the newly updated account with balance of 500 EUR. That’s exactly what we want to achieve!

## Update account balance with lock
Let’s go back to the account.sql file, and add a new query to get account for update:
```sql
-- name: GetAccountForUpdate :one
SELECT * FROM accounts
WHERE id = $1 LIMIT 1
FOR UPDATE;
```
We can use it in our money transfer transaction. Here, to get the first account, we call q.GetAccountForUpdate() instead of q.GetAccount().

Unfortunately, it still fails. This time the error is deadlock detected. So what can we do?

## Debug a deadlock
In order to figure out why deadlock occured, we need to print out some logs to see which transaction is calling which query and in which order.

To add the transaction name to the context, we call `context.WithValue()`, pass in the background context as its parent, and a pair of key value, where value is the transaction name.

### As you can see here
* Transaction 2 ran its first 2 operations: create transfer and create entry 1.
* Then transaction 1 jumped in to run its create transfer operation.
* Transaction 2 came back and continued running its next 2 operations: create entry 2 and get account 1.
* Finally the transaction 1 took turn and ran its next 4 operations: create entry 1, create entry 2, get account 1, and update account 1.
* At this point, we got a deadlock.

### Replicate deadlock in psql console
It sounds strange because transaction 2 only creates a record in transfers table while we’re getting a record from accounts table. 

Why a INSERT into 1 table can block a SELECT from other table?

To confirm this, let’s open this [Postgres Wiki page](https://wiki.postgresql.org/wiki/Lock_Monitoring) about lock monitoring.

```sql
SELECT blocked_locks.pid     AS blocked_pid,
        blocked_activity.usename  AS blocked_user,
        blocking_locks.pid     AS blocking_pid,
        blocking_activity.usename AS blocking_user,
        blocked_activity.query    AS blocked_statement,
        blocking_activity.query   AS current_statement_in_blocking_process
FROM  pg_catalog.pg_locks         blocked_locks
JOIN pg_catalog.pg_stat_activity blocked_activity  ON blocked_activity.pid = blocked_locks.pid
JOIN pg_catalog.pg_locks         blocking_locks 
    ON blocking_locks.locktype = blocked_locks.locktype
    AND blocking_locks.database IS NOT DISTINCT FROM blocked_locks.database
    AND blocking_locks.relation IS NOT DISTINCT FROM blocked_locks.relation
    AND blocking_locks.page IS NOT DISTINCT FROM blocked_locks.page
    AND blocking_locks.tuple IS NOT DISTINCT FROM blocked_locks.tuple
    AND blocking_locks.virtualxid IS NOT DISTINCT FROM blocked_locks.virtualxid
    AND blocking_locks.transactionid IS NOT DISTINCT FROM blocked_locks.transactionid
    AND blocking_locks.classid IS NOT DISTINCT FROM blocked_locks.classid
    AND blocking_locks.objid IS NOT DISTINCT FROM blocked_locks.objid
    AND blocking_locks.objsubid IS NOT DISTINCT FROM blocked_locks.objsubid
    AND blocking_locks.pid != blocked_locks.pid

JOIN pg_catalog.pg_stat_activity blocking_activity ON blocking_activity.pid = blocking_locks.pid
WHERE NOT blocked_locks.granted;
```

This long and complex query allows us to `look for blocked queries` and `what is blocking them`.

As you can see, the blocked statement is
```sql
SELECT FROM accounts FOR UPDATE;
```

And the one that’s blocking it is 
```
INSERT INTO transfers;
```
So it’s true that queries on these 2 different tables can block each other.

## Let’s dig deeper to understand why the SELECT query has to wait for the INSERT query.

If we go back to the [Postgres Wiki page](https://wiki.postgresql.org/wiki/Lock_Monitoring) and scroll down a bit, we will see another query that will allow us to list all the locks in our database.

![](https://i.imgur.com/58Kl5fX.jpg)

The from_account_id and to_account_id columns of transfers table are referencing the id column of accounts table. So any UPDATE on the account ID will affect this foreign key constraint.

That’s why when we select an account for update, it needs to acquire a lock to prevent conflicts and ensure the consistency of the data.

## Fix deadlock [the bad way]
Remove foreign key

## Fix dead lock [the better way]
```sql
-- name: UpdateAccount :one
UPDATE accounts
SET balance = $2
WHERE id = $1
RETURNING *;
```

The account ID will never be changed because it’s the primary key of accounts table.

So if we can tell Postgres that I’m selecting this account for update, but its primary key won’t be touched, then Postgres will not need to acquire the transaction lock, and thus no deadlock.

Fortunately, it’s super easy to do so. In the GetAccountForUpdate query, instead of just SELECT FOR UPDATE, we just need to say more clearly:
`SELECT FOR NO KEY UPDATE`

```sql
-- name: GetAccountForUpdate :one
SELECT * FROM accounts
WHERE id = $1 LIMIT 1
FOR NO KEY UPDATE;
```

## Update account balance [the better way]
Now before we finish, I’m gonna show you a much better way to implement this update account balance operation.

```sql
-- name: AddAccountBalance :one
UPDATE accounts
SET balance = balance + $1
WHERE id = $2
RETURNING *;
```

### modify the generate params with `sqlc.arg`

```sql
-- name: AddAccountBalance :one
UPDATE accounts
SET balance = balance + sqlc.arg(amount)
WHERE id = sqlc.arg(id)
RETURNING *;
```