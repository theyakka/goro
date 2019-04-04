
<img src="https://storage.googleapis.com/product-logos/logo_goro.png" align="center" width="160">

Goro is a mighty fine routing toolkit for Go web applications. It is designed to
be fast, yet flexible.

[![CircleCI](https://circleci.com/gh/theyakka/goro.svg?style=svg)](https://circleci.com/gh/theyakka/goro)
[![Go Version](https://img.shields.io/badge/Go-1.7+-lightgrey.svg)](https://golang.org/)
[![codecov](https://codecov.io/gh/theyakka/goro/branch/master/graph/badge.svg)](https://codecov.io/gh/theyakka/goro)


# Features

Goro is LOADED with features, but no bloat.

- Built in `Context` support while maintaining standard Go `http.Handler` interfaces
- Flexible routing options with *wildcards* and *variables*
- Prioritized route definitions with caching
- Handler `Filter`s that allow for pre-execution `Request` or `Context` modification
- Static asset mapping
- Support for subdomains
- Handler chaining built-in

# Installing

**Goro requires Go 1.7+.** We recommend using the latest version of Go if you can.

To install, run:

```
go get -u github.com/theyakka/goro
```

You can then import goro using:

```
import github.com/theyakka/goro
```

# Getting started

The basic code breakdown looks something like this:

```go
router := goro.NewRouter()
router.Add("GET", "/").HandleFunc(rootHandler)
http.ListenAndServe(":8080", router)
```

Pretty standard for most routers in Go. But Goro has so much more going for it. Rather
than try to describe it all here, you should check out [The Goro Guide](https://github.com/theyakka/goro/wiki).

# FAQ

## Why should I use this and not ____?

I'm not going to make any claims that Goro is the fastest router on the market or that it'll make you a million bucks. The likelihood is that even if those were true for you, they might not be for others.

What we *will* say is that we have tried A LOT of web frameworks over many languages and that we invested in making Goro out of unhappiness with what we saw generally. If you're here, then maybe you have also.

Goro was designed from the ground up as a Router that we wanted to use and not to copy anyone else. It has the features we think are important, and is architected in a way that we think makes managing all this stuff super simple.

Give it a try and if you like it, let us know! Either way, we love feedback.

## Has it been tested in production? Can I use it in production?

The code here has been written based on experiences with clients of all sizes. It has been production tested. That said, code is always evolving. We plan to keep on using it in production but we also plan to keep on improving it. If you find a bug, let us know!

## Who the f*ck is Yakka?

Yakka is the premier Flutter agency and a kick-ass product company. We focus on the work. Our stuff is at [http://theyakka.com](http://theyakka.com). Go check it out.

# Outro

## Credits

Goro is sponsored, owned and maintained by [Yakka LLC](http://theyakka.com). Feel free to reach out with suggestions, ideas or to say hey.

### Security

If you believe you have identified a serious security vulnerability or issue with Goro, please report it as soon as possible to apps@theyakka.com. Please refrain from posting it to the public issue tracker so that we have a chance to address it and notify everyone accordingly.

## License

Goro is released under a modified MIT license. See LICENSE for details.
