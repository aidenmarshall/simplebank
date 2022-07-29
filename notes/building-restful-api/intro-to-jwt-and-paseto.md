{%hackmd BJrTq20hE %}
# Why PASETO is better than JWT for token-based authentication?
###### tags: `simplebank`

Section 9 of [Building RESTful HTTP JSON API](/Ts3fNR-oTPCvC2mnrWDHyQ)

[article](https://dev.to/techschoolguru/how-to-write-stronger-unit-tests-with-a-custom-go-mock-matcher-55pc)
[youtube](https://www.youtube.com/watch?v=mJ8b5GcvoxQ&list=PLy_6D98if3ULEtXtNSY_2qN21VCKgoQAE&index=19)

# JWT vs PASETO
Nowadays, `token-based authentication` has become more and more popular in the development of web and mobile applications.
- JSON web token (or JWT) is one of the most widely used.
    - However, in the past few years, we’ve also discovered several security issues regarding JSON web token, mainly because of its poorly designed standard.

- So recently people have started migrating to other types of tokens such as `PASETO`, which promises to bring better security to the application.

# Token-based authentication

![](https://i.imgur.com/SQeGPWJ.png)
1. Basically, in this authentication mechanism, the client will make the first request to log in user, where it provides the username and password to the server.
2. The server will check if the username and password are correct or not. If they are, the server will create and sign a token with its secret or private key, then sends back a `200 OK` response to the client together with the `signed access token`.
    - The reason it’s called `access token` is that later the client will use this token to get access to other resources on the server.

![](https://i.imgur.com/U0PE3n6.png)
1. When client makes a request to the server, it `embeds the user’s access token` in the header of the request.
2. Upon receiving this request, the server will verify if the provided token is valid or not. 

## lifetime
Note that an access token normally has a lifetime duration before it gets expired. 

And during this time, the client can use the same token to send multiple requests to the server.

# JSON Web Token
Here’s an example of a JSON Web Token:
![](https://i.imgur.com/uT6Sa3x.png)
It is a base64 encoded string, composed of 3 main parts, separated by a dot.
1. the header of the token
    - token type: JWT
    - the algorithm: HS256
2. the payload data
    - information about the logged-in user
        - id
            - to uniquely identify the token. It will be useful in case we want to revoke access of the token in case it is leaked.
        - username
        - timestamp when the token will be expired
        - ...
    - You can customize this JSON payload to store any other information you want.
3. the digital signature
![](https://i.imgur.com/Iz87Qhm.png)

## How can server verify the authenticity of the access token?

The idea is simple, only the server has the secret/private key to sign the token. So if a hacker attempts to create a fake token without the correct key, it will be easily detected by the server in the verification process.

The JWT standards provide many different types of digital signature algorithms, but they can be classified into 2 main categories.
### Symmetric-key algorithm
the same secret key is used to both sign and verify the tokens.
![](https://i.imgur.com/1d5Gy93.png)
![](https://i.imgur.com/v90pppo.png)
Some specific algorithms which belong to this symmetric-key category are: `HS256`, `HS384`, and `HS512`.

Here `HS256` is the combination of HMAC and SHA-256. `HMAC` stands for `Hash-based Message Authentication Code`, and `SHA` is the `Secure Hashing Algorithm`. While 256/384/512 is the number of output bits.

#### weakness
However, we cannot use it in case there’s an external `third-party service that wants to verify the token`, because it would mean we must give them our secret key.


### Asymmetric-key algorithm
![](https://i.imgur.com/9Fuphp2.png)
The private key is used to sign the token, while the public key is used only to verify it.

Within this asymmetric-key category, there are several groups of algorithms, such as RS group, PS group, or ES group.
![](https://i.imgur.com/kWSHOSR.png)

# Problems of JWT
## Weak algorithms
* RSA with PKCSv1.5 is susceptible to a padding oracle attack.
* Or ECDSA can face an invalid-curve attack.
![](https://i.imgur.com/i4gNXvp.png)
For developers without deep experience in security, it would be hard for them to know which algorithm is the best to use.

## Trivial Forgery
JSON web token makes token forgery so trivial

One bad thing about JWT is that it includes the signing algorithm in the token header.
- we have seen in the past, an attacker can just set the alg header to none to bypass the signature verification process.
- Another more dangerous potential attack is to purposely set the algorithm header to a `symmetric-key` one
    - such as `HS256` while knowing that the server actually uses an asymmetric-key algorithm
    - such as `RSA` to sign and verify the token.

### hack with RSA
the hacker can just create a fake token of the admin user, where he purposely set the algorithm header to HS256, which is a symmetric-key algorithm.

 the hacker can just create a fake token of the admin user, where he purposely set the algorithm header to HS256, which is a symmetric-key algorithm.
 
![](https://i.imgur.com/n1R6kmi.png)

However, since the token’s algorithm header is saying HS256, the server will verify the signature with this symmetric algorithm HS256 instead of RSA.

***the developers didn’t check the algorithm header before verify the token signature.***
![](https://i.imgur.com/5V49bkZ.png)

# PASETO - Platform Agnostic Security Token
/puh sA to/

Platform Agnostic Security Token is one of the most successful designs that is being widely accepted by the community as the best-secured alternative to JWT.

## Strong algorithms
It solves all issues of JSON web token by first, provide strong signing algorithms out of the box.

![](https://i.imgur.com/zHFAXPL.png)
- Each `PASETO` version has already been implemented with 1 strong cipher suite.
- At any time, there will be only at most 2 latest versions of `PASETO` are active.

## Non-trivial forgery
Now with the design of PASETO, token forgery is no longer trivial.

Because the algorithm header doesn’t exist anymore, so the attacker cannot set it to none, or force the server to use the algorithm it chose in this header.
![](https://i.imgur.com/Th5bMIS.png)
Everything in the token is also authenticated with AEAD, so it’s not possible to tamper with it.

## Structure
![](https://i.imgur.com/vOO3W2P.png)
This is a PASETO version 2 token for local usage purposes. There are 4 main parts of the token, separated by a dot.

The first part is PASETO version (with red color), which is version 2.

The second part is the purpose of the token, is it used for local or public scenarios? In this case, it is local, which means using a symmetric-key authenticated encryption algorithm.

The third part (with green color) is the main content or the payload data of the token. Note that it is encrypted, so if we decrypt it using the secret key, we will get 3 smaller parts:

![](https://i.imgur.com/8cyci3z.png)
* First, the payload body. In this case, we just store a simple message and the expiration time of the token.
* Second, the nonce value that is used in both encryption and message authentication process.
* And finally the message authentication tag to authenticate the encrypted message and its associated unencrypted data.

![](https://i.imgur.com/QlXREzT.png)
You can store any public information in the footer because it won’t be encrypted like the payload body, but only base64 encoded. So anyone who has the token can decode it to read the footer data.

Note that the footer is optional, so you can have a PASETO token without a footer. For example, this is another PASETO token for the public usage scenario:
![](https://i.imgur.com/mWg6s0B.png)

While the blue part of the payload is the signature of the token, created by the digital signature algorithm using the private key.

The server will use its paired public key to verify the authenticity of this signature.
