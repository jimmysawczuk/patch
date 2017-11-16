# patch

[![GoDoc](https://godoc.org/github.com/jimmysawczuk/patch?status.svg)](https://godoc.org/github.com/jimmysawczuk/patch)

Package patch facilitates updating strongly-typed, JSON-friendly objects with weakly-typed objects that might come from an API request. It'll only touch fields in the strongly-typed object that are set in the weakly-typed object. It also allows custom validation before any fields are set.

## An example

```go
package main

import "github.com/jimmysawczuk/patch"

type basicType struct {
	A string
	B bool
	C int
	D float64
	E int `json:"e"`
}

func main() {
    v := basicType{
        A: "foo",
        B: true,
        C: 2,
        D: 3.14127,
        E: 10,
    }

    patch.Apply(&v, `{
        "A": "bar",
        "B": false
    }`)

    // v.A == "bar"
    // v.B == false
}
```
