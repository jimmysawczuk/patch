package patch

import (
	"encoding/json"
	"errors"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

type basicType struct {
	A string
	B bool
	C int
	D float64
	E int `json:"e"`
}

func TestBasicUpdate(t *testing.T) {
	var err error
	var orig basicType

	orig = getBasicOriginal()
	err = Update(&orig, getRawMessageMap(t, []byte(`{
		"A": "bar",
		"C": 2
	}`)), nil)

	assert.Nil(t, err)
	assert.Equal(t, "bar", orig.A)
	assert.Equal(t, true, orig.B)
	assert.Equal(t, 2, orig.C)
	assert.Equal(t, math.Pi, orig.D)

	orig = getBasicOriginal()
	err = Update(&orig, getRawMessageMap(t, []byte(`{
		"B": false,
		"D": 0
	}`)), nil)

	assert.Nil(t, err)
	assert.Equal(t, "foo", orig.A)
	assert.Equal(t, false, orig.B)
	assert.Equal(t, 1, orig.C)
	assert.Equal(t, float64(0), orig.D)

	orig = getBasicOriginal()
	err = Update(&orig, getRawMessageMap(t, []byte(`{}`)), nil)

	assert.Nil(t, err)
	assert.Equal(t, "foo", orig.A)
	assert.Equal(t, true, orig.B)
	assert.Equal(t, 1, orig.C)
	assert.Equal(t, math.Pi, orig.D)

	orig = getBasicOriginal()
	err = Update(&orig, getRawMessageMap(t, []byte(`{
		"E": 123
	}`)), nil)
	assert.Error(t, err)

	orig = getBasicOriginal()
	err = Update(&orig, getRawMessageMap(t, []byte(`{
		"e": 123
	}`)), nil)
	assert.Nil(t, err)
	assert.Equal(t, 123, orig.E)
}

func TestUpdateWithValidator(t *testing.T) {
	vf := ValidateFunc(func(key string, value interface{}) error {
		switch key {
		case "A":
			v := *(value.(*string))
			if v == "" {
				return errors.New("can't be empty string")
			}

		case "C":
			v := *(value.(*int))
			if v <= 0 {
				return errors.New("can't be < 0")
			}
		}

		return nil
	})

	orig := getBasicOriginal()

	err := Update(&orig, getRawMessageMap(t, []byte(`{
		"A": "bar",
		"C": 2
	}`)), vf)
	assert.Nil(t, err)

	err = Update(&orig, getRawMessageMap(t, []byte(`{
		"A": "",
		"C": 2
	}`)), vf)
	assert.Error(t, err)
	assert.IsType(t, ValidateError{}, err)
	assert.Equal(t, "validate error on key A: can't be empty string", err.Error())
}

func TestNonPointerUpdate(t *testing.T) {
	orig := basicType{}
	err := Update(orig, getRawMessageMap(t, []byte(`{
		"A": "bar",
		"C": 2
	}`)), nil)
	assert.Error(t, err)
}

func TestInvalidTypeUpdate(t *testing.T) {
	orig := basicType{}
	err := Update(&orig, getRawMessageMap(t, []byte(`{
		"A": false
	}`)), nil)
	assert.Error(t, err)

	err = Update(&orig, getRawMessageMap(t, []byte(`{
		"C": 3.1415926535
	}`)), nil)
	assert.Error(t, err)
}

func getBasicOriginal() basicType {
	return basicType{
		A: "foo",
		B: true,
		C: 1,
		D: math.Pi,
	}
}

func getRawMessageMap(t *testing.T, in []byte) map[string]json.RawMessage {
	target := map[string]json.RawMessage{}
	err := json.Unmarshal(in, &target)
	if err != nil {
		t.Fatalf("failed to unmarshal %q into map[string]json.RawMessage: %s", string(in), err)
		return nil
	}
	return target
}
