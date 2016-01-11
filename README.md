# goro

Goro is a routing toolkit for Go web applications. It is designed to be fast yet flexible.

## The parts

- **Router**: Designed for speed but meant to be flexible
- **Domain Mapper**: For domain / sub-domain mapping
- **Handler Chainer**: Flexible `http.Handler` chaining
- **Context**: Get/Set values as part of the request/response flow


## Features

**http.Handler compatible**

Goro is fully compatible with standard Go `http.Handler` inferfaces (if you're not using the `Chainer`). No wrapping, no mess, no fuss, no BS. 

**Flexible route matching**


**(Sub)domain routing**

Built in mapping for sub-domain routing provided as a wrapper so as to not complicate the core routing experience. If you don't want/need to use it, then you can focus on pure routing speed.


**Route Filtering**

Often you need to deal with a request in a special way. Filtering allows you to modify requests **before** they hit the router logic.


**Handler Chaining**

We love the chainers that exist, but we felt like it should be something that

**Request-based Contexts**

## How does it work?

**Simple example**

You can set up a simple routing configuration such as the following:

```go

package main

import (
    "net/http"
    "github.com/goposse/goro"
)

var context goro.Context

// Users.Find
func UsersFindHandler(w http.ResponseWriter, req *http.Request) {
    userID := context.Get("id")
    // find user with this ID
}

func main() {
    context := goro.NewContext()
    router := goro.NewRouter()
    router.Context = context
    router.GET("/users/{id}", UsersFindHandler)

    log.Fatal(http.ListenAndServe(":8080", router))
}
```

Straightforward as hell.

**Reusing things**

Sometimes you have a set of requirements that you use often. Goro makes re-using wildcards or substituting values easy.

```go

```


