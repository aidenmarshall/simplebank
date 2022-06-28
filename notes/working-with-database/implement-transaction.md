{%hackmd BJrTq20hE %}
# A clean way to implement database transaction in Golang
###### tags: `simplebank`

## What is transaction?
![](https://i.imgur.com/BGdcfv7.png)
### example: transfer money from one to another
![](https://i.imgur.com/uSJUys3.jpg)

## Why use transaction?
![](https://i.imgur.com/BDdFBv2.png)

## ACID Property
![](https://i.imgur.com/0eyrOdd.jpg)

## How to run SQL transactions?
![](https://i.imgur.com/UTW26cp.png)

https://dev.to/techschoolguru/a-clean-way-to-implement-database-transaction-in-golang-2ba

## Use composition to extend Queries' functionality

```
type Store struct {
    *Queries
    db *sql.DB
}
```

### Begin a DB transaction
```
func (store *Store) execTx(ctx context.Context, fn func(*Queries) error) error {
    tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	q := New(tx)
}
```

```
type DBTX interface {
    ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
    PrepareContext(context.Context, string) (*sql.Stmt, error)
    QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
    QueryRowContext(context.Context, string, ...interface{}) *sql.Row
}

func New(db DBTX) *Queries {
    return &Queries{db: db}
}
```

### Execute a generic DB transaction
now we have the queries that runs within transaction, we can call the input function with that queries, and get back an error.

```
func (store *Store) execTx(ctx context.Context, fn func(*Queries) error) error {
    tx, err := store.db.BeginTx(ctx, &sql.TxOptions)
    if err != nil {
        return err
    }

    q := New(tx)
    err = fn(q)
    if err != nil {
        if rbErr := tx.Rollback(); rbErr != nil {
            return fmt.Errorf("tx err: %v, rb err: %v", err, rbErr)
        }
        return err
    }

    return tx.Commit()
}
```

Finally, if all operations in the transaction are successful, we simply commit the transaction with tx.Commit(), and return its error to the caller.

### Implement money transfer transaction
```
func (store *Store) TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error) {
    var result TransferTxResult

    err := store.execTx(ctx, func(q *Queries) error {
        ...
        return nil
    })

    return result, err
}
```

## Write test for TransferTx
...

The Panic is cause by the shadow of testDB in TestMain
```go
var testDB *sql.DB

func TestMain(m *testing.M) {
	testDB, err := sql.Open(dbDriver, dbSource)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	testQueries = New(testDB)

	os.Exit(m.Run())
}
```
It will create a new local variable testDB in the function scope.
