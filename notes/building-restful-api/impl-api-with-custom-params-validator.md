{%hackmd BJrTq20hE %}
# Implement transfer money API with a custom params validator in Go
###### tags: `simplebank`

Section 4 of [Building RESTful HTTP JSON API](/Ts3fNR-oTPCvC2mnrWDHyQ)

[article](https://dev.to/techschoolguru/implement-transfer-money-api-with-a-custom-params-validator-in-go-2op2)
[youtube](https://www.youtube.com/watch?v=5q_wsashJZA&list=PLy_6D98if3ULEtXtNSY_2qN21VCKgoQAE&index=14)

Today, we will do some more practice by implementing the most important API of our application: transfer money API.

# Implement the transfer money API handler
## 1. Create struct to store input parameters of the API
The struct to store input parameters of this API should be transferRequest. It will have several fields:

```go=
type transferRequest struct {
    FromAccountID int64  `json:"from_account_id" binding:"required,min=1"`
    ToAccountID   int64  `json:"to_account_id" binding:"required,min=1"`
    Amount        int64  `json:"amount" binding:"required,gt=0"`
    Currency      string `json:"currency" binding:"required,oneof=USD EUR CAD"`
}
```

- The last field is the Currency of the money we want to transfer. For now, we only allow it to be either USD, EUR or CAD. And note that this currency should match the currency of both 2 accounts. We will verify that in the API handler function.

## 2. Create the handler function
```go=
func (server *Server) createTransfer(ctx *gin.Context) {
    var req transferRequest
    if err := ctx.ShouldBindJSON(&req); err != nil {
        ctx.JSON(http.StatusBadRequest, errorResponse(err))
        return
    }

    arg := db.TransferTxParams{
        FromAccountID: req.FromAccountID,
        ToAccountID:   req.ToAccountID,
        Amount:        req.Amount,
    }

    result, err := server.store.TransferTx(ctx, arg)
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, errorResponse(err))
        return
    }

    ctx.JSON(http.StatusOK, result)
}
```

### Create a function for checking the validity of an input
this create transfer handler is almost finished except that we haven’t taken into account the last input parameter: request.Currency.

```go=
func (server *Server) validAccount(ctx *gin.Context, accountID int64, currency string) bool {
    account, err := server.store.GetAccount(ctx, accountID)
    if err != nil {
        if err == sql.ErrNoRows {
            ctx.JSON(http.StatusNotFound, errorResponse(err))
            return false
        }

        ctx.JSON(http.StatusInternalServerError, errorResponse(err))
        return false
    }

    if account.Currency != currency {
        err := fmt.Errorf("account [%d] currency mismatch: %s vs %s", account.ID, account.Currency, currency)
        ctx.JSON(http.StatusBadRequest, errorResponse(err))
        return false
    }

    return true
}
```
* The first scenario is when the account doesn’t exist, then we send `http.StatusNotFound` to the client and return false.
* The second scenario is when some unexpected errors occur, so we just send `http.StatusInternalServerError` and return false.

if there’s no error, we will check if the account’s currency matches the input currency or not.

### call previously created function to check validity
```go=
func (server *Server) createTransfer(ctx *gin.Context) {
    var req transferRequest
    if err := ctx.ShouldBindJSON(&req); err != nil {
        ctx.JSON(http.StatusBadRequest, errorResponse(err))
        return
    }

    if !server.validAccount(ctx, req.FromAccountID, req.Currency) {
        return
    }

    if !server.validAccount(ctx, req.ToAccountID, req.Currency) {
        return
    }

    arg := db.TransferTxParams{
        FromAccountID: req.FromAccountID,
        ToAccountID:   req.ToAccountID,
        Amount:        req.Amount,
    }

    result, err := server.store.TransferTx(ctx, arg)
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, errorResponse(err))
        return
    }

    ctx.JSON(http.StatusOK, result)
}
```

We call server.validAccount() to check the validity of the request.fromAccountID and currency. If it’s not valid, then we just return immediately. We do the same thing for the request.toAccountID.

# Register the transfer money API route
```go=
func NewServer(store db.Store) *Server {
    server := &Server{store: store}
    router := gin.Default()

    router.POST("/accounts", server.createAccount)
    router.GET("/accounts/:id", server.getAccount)
    router.GET("/accounts", server.listAccounts)
    router.POST("/transfers", server.createTransfer)

    server.router = router
    return server
}
```
## Test the transfer money API
use Postman to test the new transfer money API.

1. Let's create a new request with method POST, and the URL is http://localhost:8080/transfers.
2. add request body
    ```json=
    {
        "from_account_id": 186,
        "to_account_id": 192,
        "amount": 10,
        "currency": "USD"
    }
    ```
3. Let’s open TablePlus to see the current data of these 2 accounts.
4. set API request
```json=
{
    "transfer": {
        "id": 96,
        "from_account_id": 186,
        "to_account_id": 192,
        "amount": 10,
        "created_at": "2022-07-12T15:47:19.104313Z"
    },
    "from_account": {
        "id": 186,
        "owner": "zqtavp",
        "balance": 271,
        "currency": "USD",
        "created_at": "2022-07-04T15:29:21.769776Z"
    },
    "to_account": {
        "id": 192,
        "owner": "yvtkss",
        "balance": 604,
        "currency": "USD",
        "created_at": "2022-07-04T15:29:21.827847Z"
    },
    "from_entry": {
        "id": 175,
        "account_id": 186,
        "amount": -10,
        "created_at": "2022-07-12T15:47:19.104313Z"
    },
    "to_entry": {
        "id": 176,
        "account_id": 192,
        "amount": 10,
        "created_at": "2022-07-12T15:47:19.104313Z"
    }
}
```