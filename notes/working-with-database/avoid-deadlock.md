{%hackmd BJrTq20hE %}
# How to avoid deadlock in DB transaction? Queries order matter!
###### tags: `simplebank`

Chapter 8 of [Backend Master Class [Golang + PostgreSQL + Kubernetes]](/3xKIijmxQJqv0z56ifCMeQ)

https://www.youtube.com/watch?v=qn3-5wdOfoA&list=PLy_6D98if3ULEtXtNSY_2qN21VCKgoQAE&index=8

One of the hardest thing when working with database transaction is locking and handling deadlock.

From author's experience, the best way to deal with deadlock is to `avoid it`.
- fine-tune our queries in the transaction

## A potential deadlock scenario

Basically we’ve fixed the deadlock issue caused by the foreign key constraints.

However, if we look at the code carefully, we can see a potential deadlock scenario.

In this transaction, we’re updating the balance of the fromAccount and the toAccount. And we know that they both require an exclusive lock to perform the operation.

So if there are 2 concurrent transactions involving the same pair of accounts, there might be a potential deadlock.

## Replicate deadlock scenario in a test
duplicate that `TestTransferTx` function, and change its name to `TestTransferTxDeadlock`.

The idea is to have 5 transactions that send money from account 1 to account 2, and another 5 transactions that send money in reverse direction, from account 2 to account 1.

***In this scenario, we only need to check for deadlock error***, we don’t need to care about the result because it has already been checked in the other test.

```go
func TestTransferTxDeadlock(t *testing.T) {
	store := NewStore(testDB)

	account1 := createRandomAccount(t)
	account2 := createRandomAccount(t)
	fmt.Println(">> before:", account1.Balance, account2.Balance)

	n := 10
	amount := int64(10)
	errs := make(chan error)

	for i := 0; i < n; i++ {
		fromAccountID := account1.ID
		toAccountID := account2.ID

		if i%2 == 1 {
			fromAccountID = account2.ID
			toAccountID = account1.ID
		}

		go func() {
			_, err := store.TransferTx(context.Background(), TransferTxParams{
				FromAccountID: fromAccountID,
				ToAccountID:   toAccountID,
				Amount:        amount,
			})

			errs <- err
		}()
	}

	for i := 0; i < n; i++ {
		err := <-errs
		require.NoError(t, err)
	}

	// check the final updated balance
	updatedAccount1, err := store.GetAccount(context.Background(), account1.ID)
	require.NoError(t, err)

	updatedAccount2, err := store.GetAccount(context.Background(), account2.ID)
	require.NoError(t, err)

	fmt.Println(">> after:", updatedAccount1.Balance, updatedAccount2.Balance)
	require.Equal(t, account1.Balance, updatedAccount1.Balance)
	require.Equal(t, account2.Balance, updatedAccount2.Balance)
}
```


## Fix the deadlock issue
So this gives us an idea of how deadlock can be avoided by 
`making both transactions update the accounts balance in the same order`.

For example, in our case, we can easily change our code so that it always updates the account with smaller ID first.

Here we check if arg.FromAccountID is less than arg.ToAccountID then the fromAccount should be updated before the toAccount. Else, the toAccount should be updated before the fromAccount.

```go
if arg.FromAccountID < arg.ToAccountID {
    result.FromAccount, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
        ID:     arg.FromAccountID,
        Amount: -arg.Amount,
    })
    if err != nil {
        return err
    }

    result.ToAccount, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
        ID:     arg.ToAccountID,
        Amount: arg.Amount,
    })
    if err != nil {
        return err
    }
} else {
    result.ToAccount, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
        ID:     arg.ToAccountID,
        Amount: arg.Amount,
    })
    if err != nil {
        return err
    }

    result.FromAccount, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
        ID:     arg.FromAccountID,
        Amount: -arg.Amount,
    })
    if err != nil {
        return err
    }
}
```

## Refactor the code
because now it looks quite long and somewhat duplicated. To do this, I’m gonna define a new addMoney() function to add money to 2 accounts.

```go
func addMoney(
    ctx context.Context,
    q *Queries,
    accountID1 int64,
    amount1 int64,
    accountID2 int64,
    amount2 int64,
) (account1 Account, account2 Account, err error) {
    account1, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
        ID:     accountID1,
        Amount: amount1,
    })
    if err != nil {
        return
    }

    account2, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
        ID:     accountID2,
        Amount: amount2,
    })
    return
}
```

```go
func (store *Store) TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error) {
    var result TransferTxResult

    err := store.execTx(ctx, func(q *Queries) error {
        ...

        if arg.FromAccountID < arg.ToAccountID {
            result.FromAccount, result.ToAccount, err = addMoney(ctx, q, arg.FromAccountID, -arg.Amount, arg.ToAccountID, arg.Amount)
        } else {
            result.ToAccount, result.FromAccount, err = addMoney(ctx, q, arg.ToAccountID, arg.Amount, arg.FromAccountID, -arg.Amount)
        }

        return err
    })

    return result, err
}
```