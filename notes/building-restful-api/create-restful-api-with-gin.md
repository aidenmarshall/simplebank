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
