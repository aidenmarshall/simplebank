{%hackmd BJrTq20hE %}
# Write Golang unit tests for database CRUD with random data
###### tags: `simplebank`

https://dev.to/techschoolguru/write-go-unit-tests-for-db-crud-with-random-data-53no

## open a connection in TestMain
If you get this error, you have to import a database driver.
```
cannot connect to db:sql: unknown driver "postgres" (forgotten import?)
```

`database/sql` package just provides a generic interface around SQL database. 

It needs to be used in conjunction with a database driver in order to talk to a specific database engine.


`lib/pq` package registers the Postgres database driver with `sql.Register` in `init()` function.
```go
func init() {
	sql.Register("postgres", &Driver{})
}
```

## init()
The init() function is typically used to initialize the state for the application.

`init()` runs before all the other code, even the main() function.


## the underscore in an import statement
From the [Go Specification](http://golang.org/ref/spec#Import_declarations):
> To import a package solely for its side-effects (initialization), use the blank identifier as explicit package name:
> import _ "lib/math"

We are using the initialization of the `lib/pq` package. 

So we have to import it with a black identifier.

## [Testify](https://github.com/stretchr/testify)
A package provides many tools for testifying that your code will behave as you intend.

### assert
The assert package provides some helpful methods that allow you to write better test code in Go.
* Prints friendly, easy to read failure descriptions
* Allows for very readable code
* Optionally annotate each assertion with a message

```go
import (
  "testing"
  "github.com/stretchr/testify/assert"
)

func TestSomething(t *testing.T) {

  var a string = "Hello"
  var b string = "Hello"

  assert.Equal(t, a, b, "The two words should be the same.")
}
```

### require
The `require` package provides same global functions as the `assert` package, but instead of returning a boolean result they terminate current test.

## Random utils
add random utils for creating random args
![](https://i.imgur.com/jSyqqD8.png)
