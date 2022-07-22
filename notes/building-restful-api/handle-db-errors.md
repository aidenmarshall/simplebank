{%hackmd BJrTq20hE %}
# How to handle DB errors in Golang correctly
###### tags: `simplebank`

Section 6 of [Building RESTful HTTP JSON API](/Ts3fNR-oTPCvC2mnrWDHyQ)

[article](https://dev.to/techschoolguru/how-to-handle-db-errors-in-golang-correctly-11ek)
[youtube](https://www.youtube.com/watch?v=mJ8b5GcvoxQ&list=PLy_6D98if3ULEtXtNSY_2qN21VCKgoQAE&index=16)

In the last lecture, we’ve added a new `users` table to the database schema. 

Today, let’s update our golang code to work with this table.

And while doing so, we’re also gonna learn `how to correctly handle some specific errors returned by Postgres`.

# Generate code to create and get user
1. create a new file user.sql inside the db/query folder

```sql
-- name: CreateUser :one
INSERT INTO users (
  username,
  hashed_password,
  full_name,
  email
) VALUES (
  $1, $2, $3, $4
) RETURNING *;

-- name: GetUser :one
SELECT * FROM users
WHERE username = $1;
```
2. run `make sqlc` to generate golang codes

# Write tests for the generated functions
1.  create a new file user_test.go file in this db/sqlc folder.
2.  copy the tests that we wrote for the create and get account function and paste them to this file
3.  adapt the copied lines for the user queries

## adapt createRandomUser
The argument variable will be of type CreateUserParams.

- The second field is `hashed_password`. Normally we will have to generate a random password and hash it using `bcrypt`, but that would be done in another lecture.

Add `RandomEmail()` to `util` package
```go
// RandomEmail generates a random email
func RandomEmail() string {
    return fmt.Sprintf("%s@email.com", RandomString(6))
}
```

the output result should be a user object 

### modify test code.
#### createRandomUser
the `user.CreatedAt` field should be not zero, since we expect the database to fill it with the current timestamp.

The last field we have to check is `user.PasswordChangedAt`. When the user is first created, we expect this field to be filled with a default value of a `zero timestamp`. The `IsZero()` function is used for checking this condition.

```go
func createRandomUser(t *testing.T) User {
    ...

    user, err := testQueries.CreateUser(context.Background(), arg)
    require.NoError(t, err)
    require.NotEmpty(t, user)

    require.Equal(t, arg.Username, user.Username)
    require.Equal(t, arg.HashedPassword, user.HashedPassword)
    require.Equal(t, arg.FullName, user.FullName)
    require.Equal(t, arg.Email, user.Email)
    require.NotZero(t, user.CreatedAt)
    require.True(t, user.PasswordChangedAt.IsZero())

    return user
}
```
#### TestGetUser
```go
func TestGetUser(t *testing.T) {
    user1 := createRandomUser(t)
    user2, err := testQueries.GetUser(context.Background(), user1.Username)
    require.NoError(t, err)
    require.NotEmpty(t, user2)

    require.Equal(t, user1.Username, user2.Username)
    require.Equal(t, user1.HashedPassword, user2.HashedPassword)
    require.Equal(t, user1.FullName, user2.FullName)
    require.Equal(t, user1.Email, user2.Email)
    require.WithinDuration(t, user1.PasswordChangedAt, user2.PasswordChangedAt, time.Second)
    require.WithinDuration(t, user1.CreatedAt, user2.CreatedAt, time.Second)
}
```

- The output user2 of this query should match the input user1.

- For a timestamp field like created_at and password_changed_at, I often use `require.WithinDuration` to compare the values because sometimes there might be a very small difference

# Test Whole Package
```
=== RUN   TestCreateAccount
    /Users/aiden/learn/backend/simplebank/db/sqlc/account_test.go:21:
                Error Trace:    account_test.go:21
                                                        account_test.go:35
                Error:          Received unexpected error:
                                pq: insert or update on table "accounts" violates foreign key constraint "accounts_owner_fkey"
                Test:           TestCreateAccount
--- FAIL: TestCreateAccount (0.03s)
```

got `pq: insert or update on table "accounts" violates foreign key constraint "accounts_owner_fkey"`

the reason is because of the `foreign key constraint violation`.

## Fix the failed tests
we’re just generating a random owner, and it doesn’t link to any existed users

In order to fix this, we have to 
1. create a user in the database first. 
2. use the created user’s username as the account owner


# Test `api` package
```
make test
```

got error

```
Running tool: /usr/local/go/bin/go test -timeout 30s -run ^TestGetAccountAPI$ github.com/aidenmarshall/simplebank/api

# github.com/aidenmarshall/simplebank/api [github.com/aidenmarshall/simplebank/api.test]
/Users/aiden/learn/backend/simplebank/api/account_test.go:78:24: cannot use store (variable of type *mockdb.MockStore) as type db.Store in argument to NewServer:
        *mockdb.MockStore does not implement db.Store (missing CreateUser method)
FAIL    github.com/aidenmarshall/simplebank/api [build failed]
```

`*mockdb.MockStore does not implement db.Store (missing CreateUser method)`

To fix this, we have to regenerate the code for the MockStore:
```
make mock
```

# Handle different types of DB error

## try creating an account for an owner that doesn’t exist
got an error because the foreign key constraint for the account owner is violated. This is expected.

However, the HTTP response status code is 500 Internal Server Error. 

>This status is not very suitable in this case since the fault is on the client’s side because it’s trying to create a new account for an inexisted user.

### check `pqErr`
after calling store.CreateAccount, if an error is returned, we will try to convert it to pq.Error type, and assign the result to pqErr variable

we can see the error’s code name is `foreign_key_violation`

## try to create a new account for an existed user a second time
got another error: `duplicate key value violates unique constraints owner_currency_key`

> In this case, we also want to return status 403 Forbidden instead of 500 Internal Server Error.

The error's code name is `unique_violation`

## response with http.`StatusForbidden` status code.
