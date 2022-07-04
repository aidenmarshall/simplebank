{%hackmd BJrTq20hE %}
# Load config from file & environment variables in Golang with Viper
###### tags: `simplebank`

Section 2 of [Building RESTful HTTP JSON API](/Ts3fNR-oTPCvC2mnrWDHyQ)

[article](https://dev.to/techschoolguru/load-config-from-file-environment-variables-in-golang-with-viper-2j2d)
[youtube](https://www.youtube.com/watch?v=n5p8HkO6bnE)

When developing and deploying a backend web application, we usually have to use different configurations for different environments such as development, testing, staging, and production.

Today we will learn how to use [Viper](https://github.com/spf13/viper) to load configurations from file or environment variables.

# Why file and environment variables
![](https://i.imgur.com/GEM2CIe.png)
- Reading values from file allows us to easily specify default configuration for local development and testing.
- While reading values from environment variables will help us override the default settings when deploying our application to staging or production using docker containers.

# Why Viper
![](https://i.imgur.com/XO4JfNw.png)
- It can find, load, and unmarshal values from a config file.
- It supports many types of files, such as JSON, TOML, YAML, ENV, or INI.
- It can also read values from environment variables or command-line flags.
- It gives us the ability to set or override default values.
- Moreover, if you prefer to store your settings in a remote system such as `Etcd` or `Consul`, then you can use viper to read data from them directly.
- It works for both unencrypted and encrypted values.
- One more interesting thing about Viper is, it can watch for changes in the config file, and notify the application about it.
- We can also use viper to save any config modification we made to the file.

# What will we do
In the current code of our simple bank project, we’re hard-coding some constants for the `dbDriver`, `dbSource` in the `main_test.go` file, and also one more constant for the `serverAddress` in the `main.go` file.
```
const (
    dbDriver      = "postgres"
    dbSource      = "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable"
    serverAddress = "0.0.0.0:8080"
)
```
## 1. Install Viper
```
go get github.com/spf13/viper

```

## 2. Create config file
create a new file `app.env` to store our config values for development.
* Each variable should be declared on a separate line.
* The variable's name is uppercase and its words are separated by an underscore.
* The variable value is followed by the name after an equal symbol.

```
DB_DRIVER=postgres
DB_SOURCE=postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable
SERVER_ADDRESS=0.0.0.0:8080
```

## 3. Load config file
```
type Config struct {
    DBDriver      string `mapstructure:"DB_DRIVER"`
    DBSource      string `mapstructure:"DB_SOURCE"`
    ServerAddress string `mapstructure:"SERVER_ADDRESS"`
}
```
* First, the `DBDriver` of type string,
* Second, the `DBSource`, also of type string.
* And finally the `ServerAddress` of type string as well.

Viper uses the `mapstructure` package under the hood for `unmarshaling` values, so we use the `mapstructure tags` to specify the name of each config field.

### define a new function LoadConfig()
It takes a path as input, and returns a config object or an error. 

This function will read configurations from a config file inside the path if it exists, or override their values with environment variables if they’re provided.

```go
func LoadConfig(path string) (config Config, err error) {
    viper.AddConfigPath(path)
    viper.SetConfigName("app")
    viper.SetConfigType("env")

    ...
}
```

#### read values from environment variables automatically
we call `viper.AutomaticEnv()` to tell viper to automatically override values that it has read from config file with the values of the corresponding environment variables if they exist.

call `viper.ReadInConfig()` to start reading config values. If error is not nil, then we simply return it.

call `viper.Unmarshal()` to unmarshals the values into the target config object.

## 4. Use LoadConfig in the main function
```go
func main() {
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal("cannot load config:", err)
	}

	conn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	store := db.NewStore(conn)
	server := api.NewServer(store)

	err = server.Start(config.ServerAddress)
	if err != nil {
		log.Fatal("cannot start server:", err)
	}
}
```

## 5. try to override configuration with environment variables.
```
SERVER_ADDRESS=0.0.0.0:8081 make server
```

## 6. Use LoadConfig in the test

let’s update the `main_test.go` file to use the new `LoadConfig()` function.

1. remove all of the hard-coded constants
2. Then in the TestMain() function, we call util.LoadConfig().
3. But this time, the main_test.go file is inside the db/sqlc folder, while the app.env config file is at the root of the repository, so we must pass in this relative path: "../.."

```go
func TestMain(m *testing.M) {
    config, err := util.LoadConfig("../..")
    if err != nil {
        log.Fatal("cannot load config:", err)
    }

    testDB, err = sql.Open(config.DBDriver, config.DBSource)
    if err != nil {
        log.Fatal("cannot connect to db:", err)
    }

    testQueries = New(testDB)

    os.Exit(m.Run())
}

```
