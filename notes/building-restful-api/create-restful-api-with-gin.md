{%hackmd BJrTq20hE %}
# Implement RESTful HTTP API in Go using Gin
###### tags: `simplebank`

Now it’s time to learn how to implement some RESTful HTTP APIs that will allow frontend clients to interact with our banking service backend.

# Go web frameworks and HTTP routers
Although we can use the standard net/http package to implement those APIs, It will be much easier to just take advantage of some existing web frameworks.

![](https://i.imgur.com/dPVKSj6.png)

They offer a wide range number of features such as routing, parameter binding, validation, middleware, and some of them even have a built-in ORM.

If you prefer a lightweight package with only routing feature, then here are some of the most popular HTTP routers for golang:
![](https://i.imgur.com/s6ar3PL.png)

## 1. Install Gin
```
go get -u github.com/gin-gonic/gin
```
## 2. Define server struct
### create struct
Now I’m gonna create a new folder called `api`. Then create a new file `server.go` inside it. This is where we implement our HTTP API server.

```
type Server struct {
    store  *db.Store
    router *gin.Engine
}
```

* The first one is a db.Store that we have implemented in previous lectures. It will allow us to interact with the database when processing API requests from clients.
* The second field is a router of type gin.Engine. This router will help us send each API request to the correct handler for processing.

### create `NewServer`
```
func NewServer(store *db.Store) *Server {
    server := &Server{store: store}
    router := gin.Default()

    // TODO: add routes to router

    server.router = router
    return server
}
```
First, we create a new Server object with the input store. Then we create a new router by calling gin.Default(). We will add routes to this router in a moment.

### add routes
```
func NewServer(store *db.Store) *Server {
    server := &Server{store: store}
    router := gin.Default()

    router.POST("/accounts", server.createAccount)

    server.router = router
    return server
}
```

- Now let’s add the first API route to create a new account. It’s gonna use POST method, so we call router.POST.
- We must pass in a path for the route, which is `/accounts` in this case, and then one or multiple handler functions.

### Implement create account API
I’m gonna implement server.createAccount method in a new file `account.go` inside the api folder.

```
func (server *Server) createAccount(ctx *gin.Context) {
    ...
}
```

Basically, when using Gin, everything we do inside a handler will involve this `context` object. It provides a lot of convenient methods to read input parameters and write out responses.

### validate input data from clients
Gin uses a `validator package` internally to perform data validation automatically under the hood.

For example, we can use a binding tag to tell Gin that the field is required. 

And later, we call the `ShouldBindJSON` function to parse the input data from HTTP request body, and Gin will validate the output object to make sure it satisfy the conditions we specified in the binding tag.
```
type createAccountRequest struct {
    Owner    string `json:"owner" binding:"required"`
    Currency string `json:"currency" binding:"required,oneof=USD EUR"`
}
```

- I’m gonna add a binding required tag to both the owner and the currency field. 
- use the oneof condition for limitation of supporting 2 types of currency for now: USD and EUR. 
    - We use a comma to separate multiple conditions, and a space to separate the possible values for the `oneof` condition.

### Implement create account API (2)
```
func (server *Server) createAccount(ctx *gin.Context) {
    var req createAccountRequest
    if err := ctx.ShouldBindJSON(&req); err != nil {
        ctx.JSON(http.StatusBadRequest, errorResponse(err))
        return
    }

    ...
}
```

### errRespone for returning error
```
func errorResponse(err error) gin.H {
    return gin.H{"error": err.Error()}
}
```

This function will take an error as input, and it will return a `gin.H` object, which is in fact just a shortcut for `map[string]interface{}`.

So we can store whatever key-value data that we want in it.

For now let’s just return gin.H with only 1 key: error, and its value is the error message. Later we might check the error type and convert it to a better format if we want.

### Implement create account API (3)
```
func (server *Server) createAccount(ctx *gin.Context) {
    var req createAccountRequest
    if err := ctx.ShouldBindJSON(&req); err != nil {
        ctx.JSON(http.StatusBadRequest, errorResponse(err))
        return
    }

    arg := db.CreateAccountParams{
        Owner:    req.Owner,
        Currency: req.Currency,
        Balance:  0,
    }

    account, err := server.store.CreateAccount(ctx, arg)
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, errorResponse(err))
        return
    }

    ctx.JSON(http.StatusOK, account)
}

```

## 3. Start HTTP server
```go
func (server *Server) Start(address string) error {
    return server.router.Run(address)
}
```
I’m gonna add a new `Start()` function to our Server struct. This function will take an address as input and return an error. Its role is to **run the HTTP server** on the input address to start listening for API requests.

### create entry point for server in main

```go
const (
    dbDriver      = "postgres"
    dbSource      = "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable"
)

func main() {
    conn, err := sql.Open(dbDriver, dbSource)
    if err != nil {
        log.Fatal("cannot connect to db:", err)
    }

    store := db.NewStore(conn)
    server := api.NewServer(store)

    ...
}
```

### to start the server
we just need to call server.Start() and pass in the server address
- In the future, we will refactor the code to load all of these configurations from environment variables or a setting file.

### a new make command to the Makefile
```
...
server:
    go run main.go

.PHONY: postgres createdb dropdb migrateup migratedown sqlc test server
```

## 4. Test create account API with Postman
Now I’m gonna use [Postman](https://www.postman.com/downloads/) to test the create account API.
![](https://i.imgur.com/ygZjKCx.png)

Let’s add a new request, select the POST method, fill in the URL, which is http://localhost:8080/accounts.

The parameters should be sent via a JSON body, so let’s select the Body tab, choose Raw, and select JSON format. We have to add 2 input fields: the owner’s name, I will use my name here, and a currency, let’s say USD.
```
{
    "owner": "Quang Pham",
    "currency": "USD"
}
```
![](https://i.imgur.com/vUKzeLG.png)

OK, then click Send.

![](https://i.imgur.com/XVCwPMP.png)

Yee, it’s successful. We’ve got a 200 OK status code, and the created account object. It has ID = 1, balance = 0, with the correct owner’s name and currency.

### try invalid inputs
![](https://i.imgur.com/02CASjv.png)

It’s really great how Gin has handled all the input binding and validation for us with just a few lines of code. It also prints out a nice form of request logs, which is very easy to read for human eyes.

## 5. implement other APIs
### Implement get account API
```go
func NewServer(store *db.Store) *Server {
    ...

    router.POST("/accounts", server.createAccount)
    router.GET("/accounts/:id", server.getAccount)

    ...
}
```

Here instead of POST, we will use GET method. And this path should include the id of the account we want to get /accounts/:id. Note that we have a colon before the id. It’s how we tell Gin that id is a URI parameter.

```go
type getAccountRequest struct {
    ID int64 `uri:"id" binding:"required,min=1"`
}
```
Now, since ID is a `URI` parameter, we cannot get it from the request body as before.

Instead, we use the `uri tag` to tell Gin the name of the URI parameter.

In this case, let’s set `min = 1`, because it’s the smallest possible value of account ID.

```go
func (server *Server) getAccount(ctx *gin.Context) {
    var req getAccountRequest
    if err := ctx.ShouldBindUri(&req); err != nil {
        ctx.JSON(http.StatusBadRequest, errorResponse(err))
        return
    }

    account, err := server.store.GetAccount(ctx, req.ID)
    if err != nil {
        if err == sql.ErrNoRows {
            ctx.JSON(http.StatusNotFound, errorResponse(err))
            return
        }

        ctx.JSON(http.StatusInternalServerError, errorResponse(err))
        return
    }

    ctx.JSON(http.StatusOK, account)
}
```
- we declare a new req variable of type getAccountRequest. Then here instead of ShouldBindJSON, we should call `ShouldBindUri`.

- If there’s an error, we just return a 400 Bad Request status code. Otherwise, we call `server.store.GetAccount()` to get the account with ID equals to `req.ID`. This function will return an account and an error.

If error is not nil, then there are 2 possible scenarios.

* The first scenario is some internal error when querying data from the database. In this case, we just return 500 Internal Server Error status code to the client.
* The second scenario is when the account with that specific input ID doesn’t exist. In that case, the error we got should be a sql.ErrNoRows. So we just check it here, and if it’s really the case, we simply send a 404 Not Found status code to the client, and return.

#### Test get account API with Postman
Let’s add a new request with method GET, and the URL is http://localhost:8080/accounts/1. We add a /1 at the end because we want to get the account with ID = 1. Now click send:
![](https://i.imgur.com/GoSHd6e.png)


The request is successful, and we’ve got a 200 OK status code together with the found account. This is exactly the account that we’ve created before.
![](https://i.imgur.com/Cn3npzY.png)

##### Try to get an account that doesn't exist.
![](https://i.imgur.com/RgxU4U1.png)
This time we’ve got a 404 Not Found status code, and an error: sql no rows in result set. Exactly what we expected

##### Try get account with a negative ID
![](https://i.imgur.com/egnVKJq.png)
Now we got a 400 Bad Request status code with an error message about the failed validation.

Alright, so our getAccount API is working well.

### Implement list account API
Next step, I’m gonna show you how to implement a list account API with `pagination`.

The number of accounts stored in our database can grow to a very big number over time. Therefore, we should not query and return all of them in a single API call.

The idea of `pagination` is to divide the records into multiple pages of small size, so that the client can retrieve only 1 page per API request.

This API is a bit different because we will not get input parameters from request body or URI, but we will get them from `query string` instead.

* We have a `page_id` param, which is the index number of the page we want to get, starting from page 1. 
* And a `page_size` param, which is the maximum number of records can be returned in 1 page.

As you can see, the page_id and page_size are added to the request URL after a question mark:
```
http://localhost:8080/accounts?page_id=1&page_size=5.
```
That’s why they’re called query parameters, and not URI parameter like the account ID in the get account request.

OK, now let’s go back to our code. I’m gonna add a new route with the same `GET` method. But this time, the path should be `/accounts` only, since we’re gonna get the parameters from the query. The handler’s name should be `listAccount`.

let’s open the account.go file to implement this server.listAccount function.
```go
type listAccountRequest struct {
    PageID   int32 `form:"page_id" binding:"required,min=1"`
    PageSize int32 `form:"page_size" binding:"required,min=5,max=10"`
    }
```

This struct should store 2 parameters, PageID and PageSize. 

- Now note that we’re not getting these parameters from uri, but from query string instead, so we cannot use the uri tag. 
    - Instead, we should use form tag.
- Both parameters are required and the minimum PageID should be 1. 
    - For the PageSize, let’s say we don’t want it to be too big or too small, so I set its minimum constraint to be 5 records, and maximum constraint to be 10 records.

```go
func (server *Server) listAccount(ctx *gin.Context) {
    var req listAccountRequest
    if err := ctx.ShouldBindQuery(&req); err != nil {
        ctx.JSON(http.StatusBadRequest, errorResponse(err))
        return
    }

    arg := db.ListAccountsParams{
        Limit:  req.PageSize,
        Offset: (req.PageID - 1) * req.PageSize,
    }

    accounts, err := server.store.ListAccounts(ctx, arg)
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, errorResponse(err))
        return
    }

    ctx.JSON(http.StatusOK, accounts)
}
```

The req variable’s type should be listAccountRequest. Then we use another binding function: ShouldBindQuery to tell Gin to get data from query string.

If an error occurs, we just return a 400 Bad Request status. Else, we call server.store.ListAccounts() to query a page of account records from the database. This function requires a ListAccountsParams as input, where we have to provide values for 2 fields: Limit and Offset.

Limit is simply the req.PageSize. While Offset is the number of records that the database should skip, wo we have to calculate it from the page id and page size using this formula: (req.PageID - 1) * req.PageSize

The ListAccounts function returns a list of accounts and an error. If an error occurs, then we just need to return 500 Internal Server Error to the client. Otherwise, we send a 200 OK status code with the output accounts list.

And that’s it, the ListAccount API is done.

#### Test list account API with Postman
![](https://i.imgur.com/is3i3AE.png)
- Let’s try to get the second page.
- try one more time to get a page that doesn’t exist, let’s say page 100.
![](https://i.imgur.com/pUCQeMR.png)
it would be better if the server returns an empty list in this case.

#### Return empty list instead of null
in the account.sql.go file that sqlc has generated for us
![](https://i.imgur.com/QIufDK0.png)

We can see that the Account items variable is declared without being initialized: var items []Account. That’s why it will remain null if no records are added.

##### emit_empty_slices
Lucky for us, in the latest released of sqlc, which is version 1.5.0, we have a new setting that will instruct sqlc to create an empty slice instead of null.

The setting is called emit_empty_slices, and its default value is false. If we set this value to true, then the result returned by a many query will be an empty slice.

OK, so now let’s add this new setting to our sqlc.yaml file:

```
version: "1"
packages:
  - name: "db"
    path: "./db/sqlc"
    queries: "./db/query/"
    schema: "./db/migration/"
    engine: "postgresql"
    emit_json_tags: true
    emit_prepared_queries: false
    emit_interface: false
    emit_exact_table_names: false
    emit_empty_slices: true
```

```
make sqlc
```

![](https://i.imgur.com/uCFkMxQ.png)

#### Test list account API with Postman (2)
try some invalid parameters
- change page_size to 20, which is bigger than the maximum constraint of 10.
- try one more time with page_id = 0
    - Now we still got `400` Bad Request status, but the error is because page_id validation failed on the required tag.![](https://i.imgur.com/IKRXnd4.png)

    - What happens here is, in the validator package, any zero-value will be recognized as missing. It’s acceptable in this case because we don’t want to have the 0 page, anyway.

However, if your API has a `zero value` parameter, then you need to pay attention to this. I recommend you to read the [documentation of validator package](https://godoc.org/github.com/go-playground/validator) to learn more about it.
