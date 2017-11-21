# patch

[![GoDoc](https://godoc.org/github.com/jimmysawczuk/patch?status.svg)](https://godoc.org/github.com/jimmysawczuk/patch)

[![Go Report Card](https://goreportcard.com/badge/github.com/jimmysawczuk/patch)](https://goreportcard.com/report/github.com/jimmysawczuk/patch)

Package patch facilitates updating strongly-typed, JSON-friendly objects with weakly-typed objects that might come from an API request. It'll only touch fields in the strongly-typed object that are set in the weakly-typed object. It also allows custom validation before any fields are set.

## An example

```go
package main

import (
	"fmt"

	"github.com/jimmysawczuk/patch"
)

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

	err := patch.Apply(&v, []byte(`{
		"A": "bar",
		"B": false
	}`), nil)

	fmt.Println("err:", err)
	fmt.Println("updated value:", v)
	// v.A == "bar"
	// v.B == false

	err = patch.Apply(&v, []byte(`{
		"A": 1
	}`), nil)

	// an err is returned because A's type doesn't match its target object.
	fmt.Println("err:", err)

	err = patch.Apply(&v, []byte(`{
		"C": -1
	}`), patch.ValidateFunc(func(key string, value interface{}) error {
		switch key {
		case "C":
			v := *(value.(*int))
			if v <= 0 {
				return fmt.Errorf("C can't be <= 0")
			}
		}

		return nil
	}))

	// an err is returned because C's value doesn't pass the validation rule.
	fmt.Println("err:", err)
}
```

## To do

- Implement common validation functions as additional struct tags:

	```go
	type Foo struct {
		A string `patch:"nonempty"`
	}
	```
