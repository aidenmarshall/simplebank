{%hackmd BJrTq20hE %}
# Add users table with unique & foreign key constraints in PostgreSQL
###### tags: `simplebank`

Section 5 of [Building RESTful HTTP JSON API](/Ts3fNR-oTPCvC2mnrWDHyQ)

[article](https://dev.to/techschoolguru/add-users-table-with-unique-foreign-key-constraints-in-postgresql-1i29)
[youtube](https://www.youtube.com/watch?v=D4VtNC3vQUs&list=PLy_6D98if3ULEtXtNSY_2qN21VCKgoQAE&index=15)

So today, we’re gonna take the first step to implement `authentication` and `authorization`
- adding a new users table to the database
- link it with the existing accounts table via some db constraints.

# Add table users
Now, the accounts table has an owner field to tell us whom this account belong to

-  use this field as a foreign key to link to the new users table that we’re going to create
## define `Table users`
```
Table users as U {
  username varchar [pk]
  hashed_password varchar [not null]
  full_name varchar [not null]
  email varchar [unique, not null]
  password_changed_at timestamptz [not null, default: '0001-01-01 00:00:00Z']
  created_at timestamptz [not null, default: `now()`]
}
```

- users table should have is email.
    - We will use it later to communicate with the users
        - reset password
    - unique

- `password_changed_at`
    - for security reason, it’s often a good idea to ask users to change their password frequently
    - use a `default value` which is a long time in the past
        - zero value timestamp of Go
            - `'0001-01-01 00:00:00Z'`

### Why is it hashed_password and not just password?

critical security issue

### Why each field is `not null`?
> The reason I want every field to be not null is because it makes our developer’s life much easier since we don’t have to deal with null pointers.

So, in practical, maybe there are some circumstatnces that using value can be null is better
- saving space

# Add foreign key constraint
This course designs that ***1 user to have multiple accounts with different currencies***
- link the owner field of the accounts table to the username field of users table.
![](https://i.imgur.com/kGSyuFQ.png)

# Add unique constraint
> accounts should have different currencies.

add a `composite unique index` to the accounts table.

## why use `unique index` rather than `unique`
UNIQUE could be better, the step below will change it to `unique`

under the hood, adding `unique constraint` will automatically create the same unique composite index for owner and currency

Postgres will need that index to check and enforce the unique constraint faster.

# Export to PostgreSQL
- users table

```sql=
CREATE TABLE "users" (
  "username" varchar PRIMARY KEY,
  "hashed_password" varchar NOT NULL,
  "full_name" varchar NOT NULL,
  "email" varchar UNIQUE NOT NULL,
  "password_changed_at" timestamptz NOT NULL DEFAULT '0001-01-01 00:00:00Z',
  "created_at" timestamptz NOT NULL DEFAULT (now())
);
```
- foreign key constraint to `username`
```sql=
ALTER TABLE "accounts" ADD FOREIGN KEY ("owner") REFERENCES "users" ("username");
```
- composite unique index for the owner and currency

```sql=
CREATE UNIQUE INDEX ON "accounts" ("owner", "currency");
```

# Add new schema change to our project
the right way to apply new schema change is to create a new migration version

```
migrate create -ext sql -dir db/migration -seq add_users
```
1. tell migrate to set the the output file extension to sql
2. the output directory to db/migration
3. use a sequential number as the file name prefix
4. the migration name is add_users.

# Implement the up migration
```sql=
CREATE TABLE "users" (
  "username" varchar PRIMARY KEY,
  "hashed_password" varchar NOT NULL,
  "full_name" varchar NOT NULL,
  "email" varchar UNIQUE NOT NULL,
  "password_changed_at" timestamptz NOT NULL DEFAULT('0001-01-01 00:00:00Z'),  
  "created_at" timestamptz NOT NULL DEFAULT (now())
);

ALTER TABLE "accounts" ADD FOREIGN KEY ("owner") REFERENCES "users" ("username");

CREATE UNIQUE INDEX ON "accounts" ("owner", "currency");
```

## A way to ensure that each owner has at most 1 account for a specific currency.
```
ALTER TABLE "accounts" ADD CONSTRAINT "owner_currency_key" UNIQUE ("owner", "currency");
```

under the hood, adding `unique constraint` will automatically create the same unique composite index for owner and currency

Postgres will need that index to check and enforce the unique constraint faster.

# Run the migration up
1. `make migrateup` - fail
    - Because owner field is completely random and doesn’t link to any existed user.
        - in this case, we have to clean up all the existing data before running migrate up
            -  This is possible because our existing system is not ready to deploy to production yet.

    - As the previous migrate up run was failed, it will change the current schema migration to version 2 but in a dirty state.


2. try to run `make migratedown`
So now if we run make migratedown with the purpose of cleaning up the data, we will get an error because of this dirty version.
- set dirty to false and run migratedown manually

3. clean existing data
4. run `make migrateup`

# Implement the migration down
Reverse what has done before in migrateion up
- drop the unique constraint for the owner and currency pair of the accounts table.
```sql=
ALTER TABLE IF EXISTS "accounts" DROP CONSTRAINT IF EXISTS "owner_currency_key";
```
- drop the foreign key constraint for the owner field
```sql=
ALTER TABLE IF EXISTS "accounts" DROP CONSTRAINT IF EXISTS "accounts_owner_fkey";
```
- drop the users table.
```sql=
DROP TABLE IF EXISTS "users";
```

# Test the up and down migrations
## add script to run 1 last migration
```
migratedown1:
  migrate -path db/migration -database "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable" -verbose down 1
  
migrateup1:
  migrate -path db/migration -database "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable" -verbose up 1
```