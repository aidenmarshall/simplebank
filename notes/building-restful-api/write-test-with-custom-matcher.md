{%hackmd BJrTq20hE %}
# How to write stronger unit tests with a custom go-mock matcher
###### tags: `simplebank`

Section 8 of [Building RESTful HTTP JSON API](/Ts3fNR-oTPCvC2mnrWDHyQ)

[article](https://dev.to/techschoolguru/how-to-write-stronger-unit-tests-with-a-custom-go-mock-matcher-55pc)
[youtube](https://www.youtube.com/watch?v=mJ8b5GcvoxQ&list=PLy_6D98if3ULEtXtNSY_2qN21VCKgoQAE&index=18)

In this lecture, we will learn how to write a custom gomock matcher to make our Golang unit tests stronger.

# The weak unit test
## basic tests
### the simple version of the tests
```go
func TestCreateUserAPI(t *testing.T) {
    user, password := randomUser(t)

    testCases := []struct {
        name          string
        body          gin.H
        buildStubs    func(store *mockdb.MockStore)
        checkResponse func(recoder *httptest.ResponseRecorder)
    }{
        {
            name: "OK",
            body: gin.H{
                "username":  user.Username,
                "password":  password,
                "full_name": user.FullName,
                "email":     user.Email,
            },
            buildStubs: func(store *mockdb.MockStore) {
                store.EXPECT().
                    CreateUser(gomock.Any(), gomock.Any()).
                    Times(1).
                    Return(user, nil)
            },
            checkResponse: func(recorder *httptest.ResponseRecorder) {
                require.Equal(t, http.StatusOK, recorder.Code)
                requireBodyMatchUser(t, recorder.Body, user)
            },
        },
    }
    ...
}
```

There are several different cases we can test, such as:
* The successful case
* Internal server error case
* Duplicate username case
* Invalid username, email, or password case
### iterate through all of test cases
We iterate through all of these cases, and run a separate sub-test for each of them.
```go
func TestCreateUserAPI(t *testing.T) {
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

            // Marshal body data to JSON
            data, err := json.Marshal(tc.body)
            require.NoError(t, err)

            url := "/users"
            request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
            require.NoError(t, err)

            server.router.ServeHTTP(recorder, request)
            tc.checkResponse(recorder)
        })
    }
}
```

1. In each sub-test, we create a new `gomock controller`, and use it to build a new `mock DB store`.
2. Then we call the `buildStubs()` function of the test case to set up the stubs for that store.
3. After that, we create a new `server` using the `mock store`, and create a new `HTTP response recorder` to record the result of the API call.
4. marshal the input request body to JSON, and make a new `POST request` to the create-user API endpoint with that JSON data.
5. We call `server.router.ServeHTTP()` function with the recorder and request object. And finally just call `tc.checkResponse()` function to check the result.

## What's the weakness of the test case?

In the build stubs function, we expect the `CreateUser()` function of the store to be called with 2 parameters. 

In this simple version, we’re using the `gomock.Any()` matcher for both of them.

However, using that same matcher for the second argument will weaken the test, because it won’t be able to `detect if the createUserParams object passed in the CreateUser() function is correct or not`

This is very bad, because the implementation of the handler is completely wrong, but the test could not detect it!
- set the argument variable to an empty CreateUserParam{} object.
- ignore the user’s input password, and just hash a constant value, such as "xyz" here.

# Try using gomock.Eq

1. declare a new arg variable of type `db.CreateUserParams`, where `username` is `user.Username`.
2. For the `HashedPassword` field, we need to hash the input naked password
3. the `FullName` should be `user.FullName`, and finally `Email` should be `user.Email`.
4. replace `gomock.Any()` matcher with `gomock.Eq(arg)`

The hash password in generated user is always different to the hash password calculated by HashPassword because the `random salt`.

So we cannot simply use the built-in `gomock.Eq()` matcher to compare the argument.

The only way to fix this properly is `to implement a new custom matcher` on our own in this case.

# Implement a custom gomock matcher
Check out the interface of `Matcher`

```go
func Eq(x interface{}) Matcher { return eqMatcher{x} }

type Matcher interface {
    // Matches returns whether x is a match.
    Matches(x interface{}) bool

    // String describes what the matcher matches.
    String() string
}
```

For our custom matcher, we will have to write a similar implementation of the `Matcher` interface, which has only 2 methods:
* The first one is `Matches()`, which should return whether the input x is a match or not.
* And the second one is `String()`, which just describes what the matcher matches for logging purpose.

## eqCreateUserParamsMatcher
* First the arg field of type `db.CreateUserParams`
* And second, the `password` field to store the naked password value.

now let’s implement the Matches() function. Since the input parameter x is an interface, we should convert it to db.CreateUserParams object.

```go
func (e eqCreateUserParamsMatcher) Matches(x interface{}) bool {
    arg, ok := x.(db.CreateUserParams)
    if !ok {
        return false
    }

    err := util.CheckPassword(e.password, arg.HashedPassword)
    if err != nil {
        return false
    }

    e.arg.HashedPassword = arg.HashedPassword
    return reflect.DeepEqual(e.arg, arg)
}
```
### String()
```go
func (e eqCreateUserParamsMatcher) String() string {
    return fmt.Sprintf("matches arg %v and password %v", e.arg, e.password)
}
```

## add a function to return an instance of this matcher

```go
func EqCreateUserParams(arg db.CreateUserParams, password string) gomock.Matcher {
    return eqCreateUserParamsMatcher{arg, password}
}
```

## Final
```go
func TestCreateUserAPI(t *testing.T) {
    user, password := randomUser(t)

    testCases := []struct {
        name          string
        body          gin.H
        buildStubs    func(store *mockdb.MockStore)
        checkResponse func(recoder *httptest.ResponseRecorder)
    }{
        {
            name: "OK",
            body: gin.H{
                "username":  user.Username,
                "password":  password,
                "full_name": user.FullName,
                "email":     user.Email,
            },
            buildStubs: func(store *mockdb.MockStore) {
                arg := db.CreateUserParams{
                    Username: user.Username,
                    FullName: user.FullName,
                    Email:    user.Email,
                }
                store.EXPECT().
                    CreateUser(gomock.Any(), EqCreateUserParams(arg, password)).
                    Times(1).
                    Return(user, nil)
            },
            checkResponse: func(recorder *httptest.ResponseRecorder) {
                require.Equal(t, http.StatusOK, recorder.Code)
                requireBodyMatchUser(t, recorder.Body, user)
            },
        },
        ...
    }

    ...
}
```