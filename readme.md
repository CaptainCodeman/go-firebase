# go-firebase

AppEngine friendly Firebase for Go (Golang)

Currently just the auth pieces to verify and mint custom tokens.

## Why another package?

There are a few existing firebase packages for Go but none of them seemed to work 
quite right and / or didn't work at all with AppEngine (standard) so this is a
hacked together version that works for me which I suggest you use with caution, if
at all.

This package borrows heavily from prior art, mostly [Firebase Server SDK for Golang
](https://github.com/wuman/firebase-server-sdk-go)

## Why custom tokens?

The firebase auth system is convenient and (currently) free to use and if you're
using the firebase database it's very simple and easy.

But if you have any legacy REST API that you want to use things are not quite so
obvious. Sure, you could just lookup the firebase user on each request but that is
really losing what makes bearer tokens so valuable - having a JWT that authorizes
the request without having to keep track of server-side sessions, so you can scale
your API.

You might also want some custom claims to be available in the JWT so that you can
[decode it on the client](https://github.com/auth0/jwt-decode) and adapt the UI to
match the user's roles for example.

OK, so you need custom tokens.

Now you need to jump through a few hoops and will need a server to both verify the 
firebase issued auth tokens passed to it (for, you know, security) before correctly
producing your own signed custom tokens that firebase will accept for authentication.

This is what this library does.

## What do I do on the client?

You need to do a few extra steps in order to use custom tokens on the client and
also get the correct JWT to pass to the backend (non-firebase) REST API.

The steps are:

* Sign in user with `signInWithEmailAndPassword` or one of the 3rd party providers
* Get the user token via `user.getToken(true)` (use false if *just* signed in)
* Pass the token to the auth server which issues a custom token with extra claims
* Sign the user in with that token (`auth.signInWithCustomToken`) 
* Get the user token via `user.getToken(false)` (yes, it's another token)

The last token is the one that you can send to your REST API to authorize requests.
If you only need to add extra claims for use with firebase rules, the last step can
be skipped.

### Example tokens

Here's an example of the auth tokens showing the different versions at each step
(tip: the [JWT Debugger](https://jwt.io/) helps when working with tokens):

Token received from firebase after `signInWithEmailAndPassword`:

```
{
  "iss": "https://securetoken.google.com/captain-codeman",
  "aud": "project-name",
  "auth_time": 1479745491,
  "user_id": "RE8hG0RX4YVMHHjferfb8tu4jRr2",
  "sub": "RE8hG0RX4YVMHHjferfb8tu4jRr2",
  "iat": 1479745491,
  "exp": 1479749091,
  "email": "email@address",
  "email_verified": false,
  "firebase": {
    "identities": {
      "email": [
        "email@address"
      ]
    },
    "sign_in_provider": "password"
  }
}
```

Token we get back from our custom token service:
```
{
  "aud": "https://identitytoolkit.googleapis.com/google.identity.identitytoolkit.v1.IdentityToolkit",
  "claims": {
    "roles": [
      "admin",
      "operator"
    ],
    "uid": 1
  },
  "exp": 1479749434,
  "iat": 1479745834,
  "iss": "firebase-adminsdk-0rpgf@project-name.iam.gserviceaccount.com",
  "sub": "firebase-adminsdk-0rpgf@project-name.iam.gserviceaccount.com",
  "uid": "RE8hG0RX4YVMHHjferfb8tu4jRr2"
}
```

Token we get after signing in with the custom token and using `user.getToken()`:
```
{
  "iss": "https://securetoken.google.com/project-name",
  "roles": [
    "admin",
    "operator"
  ],
  "uid": 1,
  "aud": "project-name",
  "auth_time": 1479745834,
  "user_id": "RE8hG0RX4YVMHHjferfb8tu4jRr2",
  "sub": "RE8hG0RX4YVMHHjferfb8tu4jRr2",
  "iat": 1479745834,
  "exp": 1479749434,
  "email": "email@address",
  "email_verified": false,
  "firebase": {
    "identities": {
      "email": [
        "email@address"
      ]
    },
    "sign_in_provider": "custom"
  }
}
```

Note this now includes the firebase user id (as `sub` and `user_id`), our apps
internal user id (as `uid`) and the `roles` we set - everything we might need to
authorize a REST API call on our server (just extract and verify the JWT claims).

## Server example

A very simple example server is included, note that the `app/firebase-credentials.json`
file is not included and you should instead include one created from your own project.

## Client example

I'm using [Polymer](https://www.polymer-project.org/) for my front-end and have created
an [`<auth-ajax>`](https://github.com/CaptainCodeman/auth-ajax) element to make auth-token
handling easier.

See the [demo](http://www.captaincodeman.com/auth-ajax/components/auth-ajax/demo/) which
uses an instance of this package for the server-side custom token issuing.