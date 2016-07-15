
<img src="https://raw.githubusercontent.com/goposse/goro/assets/goro_logo.png" align="center" width="172">

# goro

Goro is a mighty fine routing toolkit for Go web applications. It is designed to
be fast, yet flexible.

# Installing

**Goro requires Go 1.7+.** Goro requires the new `Context` functionality of the Go `1.7`
standard library. Its an inconvenience but its a great base for the future.

To install, run:

```
go get -u github.com/goposse/goro
```

You can then import goro using:

```
import github.com/goposse/goro
```

# Getting started

The basic code breakdown looks something like this:

```go
router := goro.NewRouter()
router.Add("GET", "/").HandleFunc(rootHandler)
http.ListenAndServe(":8080", router)
```

Pretty standard for most routers in Go.

# FAQ

## Why should I use this and not ____?

I'm not going to make any claims that Goro is the fastest router on the market or that it'll make you a million bucks. The likelihood is that even if those were true for you, they might not be for others.

What we *will* say is that here at Posse we have tried A LOT of web frameworks over many languages and that we invested in making Goro out of unhappiness with what we saw generally. If you're here, then maybe you have also.

Goro was designed from the ground up as a Router that we wanted to use and not to copy anyone else. It has the features we think are important, and is architected in a way that we think makes managing all this stuff super simple.

Give it a try and if you like it, let us know! Either way, we love feedback.

## Has it been tested in production? Can I use it in production?

The code here has been written based on Posse's experiences with clients of all sizes. It has been production tested. That said, code is always evolving. We plan to keep on using it in production but we also plan to keep on improving it. If you find a bug, let us know!

## Who the f*ck is Posse?

We're the best friggin mobile shop in NYC that's who. Hey, but we're biased. Our stuff is at [http://goposse.com](http://goposse.com). Go check it out.

# Outro

## Credits

Haitch is sponsored, owned and maintained by [Posse Productions LLC](http://goposse.com). Follow us on Twitter [@goposse](https://twitter.com/goposse). Feel free to reach out with suggestions, ideas or to say hey.

### Security

If you believe you have identified a serious security vulnerability or issue with Goro, please report it as soon as possible to apps@goposse.com. Please refrain from posting it to the public issue tracker so that we have a chance to address it and notify everyone accordingly.

## License

Goro is released under a modified MIT license. See LICENSE for details.
