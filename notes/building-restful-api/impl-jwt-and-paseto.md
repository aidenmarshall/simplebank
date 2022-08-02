{%hackmd BJrTq20hE %}
# How to create and verify JWT & PASETO token in Golang
###### tags: `simplebank`

Section 10 of [Building RESTful HTTP JSON API](/Ts3fNR-oTPCvC2mnrWDHyQ)

[article](https://dev.to/techschoolguru/how-to-create-and-verify-jwt-paseto-token-in-golang-1l5j)
[youtube](https://www.youtube.com/watch?v=mJ8b5GcvoxQ&list=PLy_6D98if3ULEtXtNSY_2qN21VCKgoQAE&index=20)

Today we will learn how to implement `JWT` and `PASETO` in Golang to see why `PASETO` is also much easier and simpler to implement compared to `JWT`.

# Declare token Maker interface
1. create a new package called token
2. create a new file maker.go inside this package

The idea is to declare a general token.Maker interface to manage the creation and verification of the tokens.
- JWTMaker
- PasetoMaker

By doing so, we can easily switch between different types of token makers whenever we want.

## Interface
```go
type Maker interface {
    CreateToken(username string, duration time.Duration) (string, error)
    VerifyToken(token string) (*Payload, error)
}
```
- The `CreateToken()` method takes a username string and a valid duration as input. It returns a signed token string or an error.
    - Basically, this method will create and sign a new token for a specific username and valid duration.
- The second method is `VerifyToken()`, which takes a token string as input, and returns a Payload object or an error.
    - The role of this VerifyToken() method is to checks if the input token is valid or not. 
    - If it is valid, the method will return the payload data stored inside the body of the token.


# Declare token Payload struct
1. let’s create a new payload.go file
2. define the Payload struct inside it.

This struct will contain the payload data of the token

```go
type Payload struct {
    ID        uuid.UUID `json:"id"`
    Username  string    `json:"username"`
    IssuedAt  time.Time `json:"issued_at"`
    ExpiredAt time.Time `json:"expired_at"`
}
```

- `Username`: The most important field is `Username`, which is used to identify the token owner.
- `IssuedAt`: Then an `IssuedAt` field to know when the token is created.
- `ExpiredAt`: When using token based authentication, it’s crucial to make sure that each access token only has a short valid duration. So we need an `ExpiredAt` field to store the time at which the token will be expired.

- `ID`: However, if we want to have a mechanism to `invalidate some specific tokens in case they are leaked`, we need to add an ID field to uniquely identify each token.

3. get `google/uuid`

## define NewPayload
```go
func NewPayload(username string, duration time.Duration) (*Payload, error) {
    tokenID, err := uuid.NewRandom()
    if err != nil {
        return nil, err
    }

    payload := &Payload{
        ID:        tokenID,
        Username:  username,
        IssuedAt:  time.Now(),
        ExpiredAt: time.Now().Add(duration),
    }
    return payload, nil
}
```

1. call `uuid.NewRandom()` to generate a unique token ID. If an error occurs, we simply return a nil payload and the error itself.
2. create the `payload`, where `ID` is the generated random token `UUID`, `Username` is the input `username`, `IssuedAt` is `time.Now()`, and `ExpiredAt` is `time.Now().Add(duration)`.
3. return this payload object and a nil error. And we’re done!

# Implement JWT Maker
get a package for JWT
```
go get github.com/dgrijalva/jwt-go
```
## 1. declare `JWTMaker` struct
This struct is a JSON web token maker, which implements the token.Maker interface.

```go
type JWTMaker struct {
    secretKey string
}
```

this tutorial uses `symmetric key algorithm` to sign the tokens, so this struct will have a field to store the secret key.

## 2. add a function `NewJWTMaker()`
It takes a `secretKey` string as input, and returns a `token.Maker` interface, or an error as output.

```go
const minSecretKeySize = 32

func NewJWTMaker(secretKey string) (Maker, error) {
    if len(secretKey) < minSecretKeySize {
        return nil, fmt.Errorf("invalid key size: must be at least %d characters", minSecretKeySize)
    }
    return &JWTMaker{secretKey}, nil
}

```
Now, although the algorithm we’re gonna use doesn’t require how long the secret key should be, it’s still a good idea to ensure that the key should not be too short, for better security. So I will declare a constant minSecretKeySize = 32 characters.


Then inside this function, we check if the length of the secret key is less than minSecretKeySize or not. If it is, we just return a nil object and an error saying that the key must have at least 32 characters.

Otherwise, we return a new JWTMaker object with the input secretKey, and a nil error.

## add the `CreateToken()` and `VerifyToken()` methods
```go
func (maker *JWTMaker) CreateToken(username string, duration time.Duration) (string, error) {}

func (maker *JWTMaker) VerifyToken(token string) (*Payload, error) {}
```

### Implement the `CreateToken()`
1. Create a new token payload by calling `NewPayload()`, and pass in the input username and valid duration.
2. If error is not nil, we return an empty token string and the error.
3. Create a new jwtToken by calling the `jwt.NewWithClaims()` function of the jwt-go package.
    - This function expects 2 input arguments:
        1. the signing method (or algorithm). I’m gonna use HS256 in this case.
        2. the claims, which actually is our created payload.

4. To generate a token string, we call jwtToken.SignedString(), and pass in the secretKey after converting it to []byte slice.

#### Implement `Valid()` for `jwt.Claims` interface

Here we have an error because our Payload struct doesn’t implement the jwt.Claims interface. It’s missing one method called Valid().

The jwt-go package needs this method to check if the token payload is valid or not. So let’s open the payload.go to add this method.

```go
var ErrExpiredToken = errors.New("token has expired")

func (payload *Payload) Valid() error {
    if time.Now().After(payload.ExpiredAt) {
        return ErrExpiredToken
    }
    return nil
}
```

The simplest but also the most important thing we must check is the expiration time of the token.

If time.Now() is after the payload.ExpiredAt, then it means that the token has expired. So we just return a new error saying: token has expired.

We should declare this error as a public constant: ErrExpiredToken, so that we can check the error type from outside.

If the token is not expired, then we simply return nil. And that’s it! The Valid function is done.

#### Implement the JWT VerifyToken method
we have to parse the token by calling jwt.ParseWithClaims and pass in the input token string, an empty Payload object, and a `key function`.

##### Key function
it’s a function that receives the parsed but unverified token.

You should verify its header to make sure that the signing algorithm matches with what you normally use to sign the tokens.

Then if it matches, you return the key so that jwt-go can use it to verify the token.

`This step is very important to prevent the trivial attack mechanism`

we can get its signing algorithm via the token.Method field. 
```go=
keyFunc := func(token *jwt.Token) (interface{}, error) {
    _, ok := token.Method.(*jwt.SigningMethodHMAC)
    if !ok {
        return nil, ErrInvalidToken
    }
    return []byte(maker.secretKey), nil
}
```
- Note that its type is a SigningMethod, which is just an interface. So we have to try converting it to a specific implementation.


##### declare error
```go
var (
    ErrInvalidToken = errors.New("token is invalid")
    ErrExpiredToken = errors.New("token has expired")
)
```