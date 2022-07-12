{%hackmd BJrTq20hE %}
# Mock DB for testing HTTP API in Go and achieve 100% coverage
###### tags: `simplebank`

Section 3 of [Building RESTful HTTP JSON API](/Ts3fNR-oTPCvC2mnrWDHyQ)

[article](https://dev.to/techschoolguru/mock-db-for-testing-http-api-in-go-and-achieve-100-coverage-4pa9)
[youtube](https://www.youtube.com/watch?v=rL0aeMutoJ0)

If you have trouble isolating unit test data to avoid conflicts, think about mock DB!

In this article, we will learn how to use `Gomock` to generate stubs for the DB interface, which helps us write API unit tests faster, cleaner, and easily achieve 100% coverage.

# Why mocking DB
When it comes to testing these APIs, some people might choose to connect to the real database, while some others might prefer to just mocking it. So which approach should we use?

![](https://i.imgur.com/IRp2UXq.png)

But is it good enough to test our API with just a mock DB? Can we be confident that our codes will still perform well when a real DB is plugged in?

![](https://i.imgur.com/vX2C88k.png)

Because our code that talks to the real DB is already tested carefully in the previous lecture.


So all we need to do is: make sure that `the mock DB implements the same interface as the real DB`. Then everything will be working just fine when being put together.

# How to mock DB
There are 2 ways to mock DB.

## Fake db
The first one is to implement a fake DB, which stores data in memory.

![](https://i.imgur.com/ZDKBS5b.png)

- We have a fake DB MemStore struct, which implements all actions of the Store interface, but only uses a map to read and write data.
    - it requires us to write a lot more codes that only be used for testing, which is quite time-consuming for both development and maintenance later.

a better way to mock DB, which is using `stubs` instead of `fake DB`.

# gomock
The idea is to use [gomock](https://github.com/golang/mock) package to generate and build stubs that return hard-coded values for each scenario we want to test.

![](https://i.imgur.com/hYiMzHb.png)
In this example, `gomock` already generated a `MockStore` for us. So all we need to do is to call its `EXPECT()` function to build a `stub`, which tells `gomock` that: this `GetAccount()` function should be 

**called exactly 1 time with this input accountID, and return this account object as output**.

## 1. Install gomock
```
go install github.com/golang/mock/mockgen@v1.6.0
```

## 2. Define Store interface
In order to use a mock DB in the API server tests, we have to replace that store object with an interface.
```go
type Store interface {
    // TODO: add functions to this interface
}

type SQLStore struct {
    db *sql.DB
    *Queries
}
```
Then this NewStore() function should not return a pointer, but just a Store interface.
```go
func NewStore(db *sql.DB) Store {
    return &SQLStore{
        db:      db,
        Queries: New(db),
    }
}
```

change the type of the store receiver of the execTx() function and the TransferTx() function to *SQLStore
```go
func (store *SQLStore) execTx(ctx context.Context, fn func(*Queries) error) error {
    ...
}

func (store *SQLStore) TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error) {
}
```

define a list of actions that the Store interface can do.

Basically, it should have `all functions of the Queries struct`, and one more function to execute the transfer money `transaction`.

```
type Store interface {
    TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error)
}
```

For the functions of the Queries struct, of course, we can do the same, like going through all of them and copy-paste one by one. However, it will be too time-consuming because this struct can contain a lot of functions.

Lucky for us, the sqlc package that we used to generate CRUD codes also has an option to emit an interface that contains all of the function of the Queries struct.

All we have to do is to change this emit_interface setting in the sqlc.yaml file to true
```
make sqlc
```

After this, in the db/sqlc folder, we can see a new file called querier.go. It contains the generated Querier interface with all functions to insert and query data from the database:
```go
type Querier interface {
    AddAccountBalance(ctx context.Context, arg AddAccountBalanceParams) (Account, error)
    CreateAccount(ctx context.Context, arg CreateAccountParams) (Account, error)
    CreateEntry(ctx context.Context, arg CreateEntryParams) (Entry, error)
    CreateTransfer(ctx context.Context, arg CreateTransferParams) (Transfer, error)
    DeleteAccount(ctx context.Context, id int64) error
    GetAccount(ctx context.Context, id int64) (Account, error)
    GetAccountForUpdate(ctx context.Context, id int64) (Account, error)
    GetEntry(ctx context.Context, id int64) (Entry, error)
    GetTransfer(ctx context.Context, id int64) (Transfer, error)
    ListAccounts(ctx context.Context, arg ListAccountsParams) ([]Account, error)
    ListEntries(ctx context.Context, arg ListEntriesParams) ([]Entry, error)
    ListTransfers(ctx context.Context, arg ListTransfersParams) ([]Transfer, error)
    UpdateAccount(ctx context.Context, arg UpdateAccountParams) (Account, error)
}

var _ Querier = (*Queries)(nil)
```

Now what we need to do is just embed this Querier inside the Store interface. That would make Store interface to have all of its functions in addition to the TransferTx() function that we’ve added before

go back to the api/server.go file and remove this * from *db.Store
```go
func NewServer(store db.Store) *Server {
    ...
}
```

we don’t have to change anything in the main.go file because the db.NewStore() function is now also returning a Store interface with the actual implementation SQLStore that connects to the real SQL DB.

## 3. Generate mock DB
Now as we have the db.Store interface, we can use gomock to generate a mock implementation of it.

create a new mock folder inside the db package.

Run `mockgen -help`

Mockgen gives us 2 ways to generate mocks.
- The `source mode` will generate mock interfaces from a single source file.
    - Things would be more complicated if this source file imports packages from other files, which is often the case when we work on a real project.
- In this case, it’s better to use the `reflect mode`, where we only need to provide the name of the package and the interface, and let mockgen use `reflection` to automatically figure out what to do.

Let's run 
```
mockgen github.com/aidenmarshall/simplebank/db/sqlc Store
```
- The first argument is an import path to the Store interface.
- The second argument we need to pass in this command is the name of the interface, which is Store in this case.

use the `-destination` option to tell it to write the mock store codes to `db/mock/store.go` file

- Error handling
    - ![](https://i.imgur.com/uX6Hacj.png)
        ```
        mockgen -destination db/mock/store.go --build_flags=--mod=mod github.com/aidenmarshall/simplebank/db/sqlc Store
        ```
    - https://github.com/golang/mock/issues/494#issuecomment-718999803
    - ![](https://i.imgur.com/47RlzKS.png)

The generated code
![](https://i.imgur.com/IfIBtfM.png)

# Write unit test for Get Account API

we need to create a new mock store using this mockdb.NewMockStore() generated function.

- It expects a `gomock.Controller` object as input, so we have to create this controller by calling `gomock.NewController` and pass in the testing.T object.
- We should defer calling `Finish` method of this controller. This is very important because it will check to see if all methods that were expected to be called were called.

create a new store by calling mockdb.NewMockStore() with this input controller.
```go
func TestGetAccountAPI(t *testing.T) {
    account := randomAccount()

    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    store := mockdb.NewMockStore(ctrl)
}
```

## build the stubs for this mock store
So let’s build stub for this method by calling `store.EXPECT().GetAccount()`.

the GetAccount() method of our Store interface requires 2 input parameters: a context and an account ID.

we have to specify what values of these 2 parameters we expect this function to be called with.

The first context argument could be any value, so we use gomock.Any() matcher for it. The second argument should equal to the ID of the random account we created above. So we use this matcher: gomock.Eq() and pass the account.ID to it.

```go
func TestGetAccountAPI(t *testing.T) {
    account := randomAccount()

    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    store := mockdb.NewMockStore(ctrl)
    store.EXPECT().
        GetAccount(gomock.Any(), gomock.Eq(account.ID)).
        Times(1).
        Return(account, nil)
}
```
we can use the Return() function to tell gomock to return some specific values whenever the GetAccount() function is called. For example, in this case, we want it to return the account object and a nil error.

Alright, now the stub for our mock Store is built. We can use it to start the test HTTP server and send GetAccount request.

## create a server by calling NewServer() function with the mock store.
```go
func TestGetAccountAPI(t *testing.T) {
    ...

    server := NewServer(store)
    recorder := httptest.NewRecorder()
}
```
For testing an HTTP API in Go, we don’t have to start a real HTTP server. 

Instead, we can just use the recording feature of the `httptest` package to record the response of the API request. So here we call `httptest.NewRecorder()` to create a new ResponseRecorder

Next we will declare the url path of the API we want to call, which should be /accounts/{ID of the account we want to get}.
```go
func TestGetAccountAPI(t *testing.T) {
    ...

    server := NewServer(store)
    recorder := httptest.NewRecorder()

    url := fmt.Sprintf("/accounts/%d", tc.accountID)
    request, err := http.NewRequest(http.MethodGet, url, nil)
    require.NoError(t, err)
}
```

Then we call server.router.ServeHTTP() function with the created recorder and request objects

Basically, this will send our API request through the server router and record its response in the recorder. All we need to do is to check that response.

```go
func TestGetAccountAPI(t *testing.T) {
    account := randomAccount()

    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    store := mockdb.NewMockStore(ctrl)
    store.EXPECT().
        GetAccount(gomock.Any(), gomock.Eq(account.ID)).
        Times(1).
        Return(account, nil)

    server := NewServer(store)
    recorder := httptest.NewRecorder()

    url := fmt.Sprintf("/accounts/%d", account.ID)
    request, err := http.NewRequest(http.MethodGet, url, nil)
    require.NoError(t, err)

    server.router.ServeHTTP(recorder, request)
    require.Equal(t, http.StatusOK, recorder.Code)
}
```

The simplest thing we can check is the HTTP status code. In the happy case, it should be http.StatusOK. This status code is recorded in the Code field of the recorder.
![](https://i.imgur.com/NSdkJY9.png)

### Check Response Body
The response body is stored in the `recorder.Body` field, which is in fact just a `bytes.Buffer` pointer.

```go
func requireBodyMatchAccount(t *testing.T, body *bytes.Buffer, account db.Account) {
    data, err := ioutil.ReadAll(body)
    require.NoError(t, err)

    var gotAccount db.Account
    err = json.Unmarshal(data, &gotAccount)
    require.NoError(t, err)
    require.Equal(t, account, gotAccount)
}
```

1. First we call ioutil.ReadAll() to read all data from the response body and store it in a data variable. We require no errors to be returned. 
2. Then we declare a new gotAccount variable to store the account object we got from the response body data. 
3. Then we call json.Unmarshal to unmarshal the data to the gotAccount object. Require no errors, then require the gotAccount to be equal to the input account.

go back to the unit test and call requireBodyMatchAccount function with the testing.T, the recorder.Body, and the generated account as input arguments.

# Achieve 100% coverage
declare a list of test cases (Table tests)

```go
func TestGetAccountAPI(t *testing.T) {
    account := randomAccount()

    testCases := []struct {
        name          string
        accountID     int64
        buildStubs    func(store *mockdb.MockStore)
        checkResponse func(t *testing.T, recoder *httptest.ResponseRecorder)
    }{
        // TODO: add test data
    }
    ...
}
```

## add the first scenario for the happy case.
```go
func TestGetAccountAPI(t *testing.T) {
    account := randomAccount()

    testCases := []struct {
        name          string
        accountID     int64
        buildStubs    func(store *mockdb.MockStore)
        checkResponse func(t *testing.T, recoder *httptest.ResponseRecorder)
    }{
        {
            name:      "OK",
            accountID: account.ID,
            buildStubs: func(store *mockdb.MockStore) {
                store.EXPECT().
                    GetAccount(gomock.Any(), gomock.Eq(account.ID)).
                    Times(1).
                    Return(account, nil)
            },
            checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
                require.Equal(t, http.StatusOK, recorder.Code)
                requireBodyMatchAccount(t, recorder.Body, account)
            },
        },
    }
    ...
}
```
## use for loop to iterate throgh the list of test cases

run each case as a separate sub-test of this unit test, so let’s call `t.Run()` function, pass in 
1. the name of this test case
2. a function that takes testing.T object as input.

```go
func TestGetAccountAPI(t *testing.T) {
    ...

    for i := range testCases {
        tc := testCases[i]

        t.Run(tc.name, func(t *testing.T) {
            ctrl := gomock.NewController(t)
            defer ctrl.Finish()

            store := mockdb.NewMockStore(ctrl)
            tc.buildStubs(store)

            server := NewServer(store)
            recorder := httptest.NewRecorder()

            url := fmt.Sprintf("/accounts/%d", tc.accountID)
            request, err := http.NewRequest(http.MethodGet, url, nil)
            require.NoError(t, err)

            server.router.ServeHTTP(recorder, request)
            tc.checkResponse(t, recorder)
        })
    }
}
```
## Second case for "NotFound"
We can use the same accountID here because mock store is separated for each test case. But we need to change our buildStubs function a bit.
```go
        {
            name:      "NotFound",
            accountID: account.ID,
            buildStubs: func(store *mockdb.MockStore) {
                store.EXPECT().
                    GetAccount(gomock.Any(), gomock.Eq(account.ID)).
                    Times(1).
                    Return(db.Account{}, sql.ErrNoRows)
            },
            checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
                require.Equal(t, http.StatusNotFound, recorder.Code)
            },
        },
```
- Here instead of returning this specific account, we should return an empty Account{} object together with a ***sql.ErrNoRows*** error.

- change the checkResponse function, because in this case, we expect the server to return http.StatusNotFound instead.

## Gin `Test mode`
There are many duplicate debug logs written by `Gin`, which make it harder to read the test result.

call gin.SetMode to change it to gin.TestMode
```go
func TestMain(m *testing.M) {
    gin.SetMode(gin.TestMode)
    os.Exit(m.Run())
}
```

# Conclusion
OK, so today we have learned how to use gomock to generate mocks for our DB interface, and use it to write unit tests for the Get Account API to achieve 100% coverage.