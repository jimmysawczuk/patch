package patch

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/pkg/errors"
)

type ValidateError struct {
	key string
	err error
}

func (u ValidateError) Error() string {
	return fmt.Sprintf("validate error on key %s: %s", u.key, u.err.Error())
}

type Validator interface {
	Validate(key string, value interface{}) error
}

type ValidateFunc func(string, interface{}) error

func (vf ValidateFunc) Validate(key string, value interface{}) error {
	return vf(key, value)
}

// Update takes takes a map[string]json.RawMessage that represents a partial target object in JSON. It then applies the
// values set in the map to the current object, only touching what's changed.
func Update(dest interface{}, src map[string]json.RawMessage, validator Validator) error {
	// dest should be a pointer here, because when we're done we'll overwrite zero or more values on it.
	if reflect.ValueOf(dest).Kind() != reflect.Ptr {
		return errors.New("destination must be a pointer")
	}

	// Get the indirect of dest to determine its concrete type.
	indirect := reflect.Indirect(reflect.ValueOf(dest))

	// We only want to touch dest if the entire function finishes successfully, rather than erroring in the middle of setting
	// some fields but not others. So we'll create a copy of dest and work with the copy until the end.
	destVal := reflect.New(indirect.Type())
	reflect.Indirect(destVal).Set(indirect)

	// Iterate through all of dest's fields, taking note of what they marshal to in JSON via the struct tags. (If there is no json tag, we assume
	// they map to the same name as the field.)
	fieldMap := map[string]int{}
	for i := 0; i < indirect.Type().NumField(); i++ {
		field := indirect.Type().Field(i)
		tag, ok := field.Tag.Lookup("json")
		if ok {
			v := strings.SplitN(tag, ",", 2)
			if v[0] != "-" {
				fieldMap[v[0]] = i
			}
		} else {
			fieldMap[field.Name] = i
		}
	}

	// We now have a map of all fields representation in JSON and where they map to on the struct. All that's left to do is iterate through
	// the incoming values and attempt to set them on our target.
	for key, val := range src {

		// Find the field on the target struct; if it's not in the map, something fishy is going on and we better abort.
		fieldIndex, ok := fieldMap[key]
		if !ok {
			return errors.Errorf("key %s wasn't found in field map", key)
		}

		// We found the field, so make a target of the same type that we can unmarshal into, then try to unmarshal it. If we can't unmarshal it,
		// abort.
		target := reflect.New(indirect.Type().Field(fieldIndex).Type).Interface()
		err := json.Unmarshal(val, target)
		if err != nil {
			return errors.Wrapf(err, "error unmarshaling %s", key)
		}

		if validator != nil {
			err = validator.Validate(key, target)
			if err != nil {
				return ValidateError{err: err, key: key}
			}
		}

		// We have our field and we have our new value, so we can go ahead and set it. Broken up into a couple lines for readability.
		targetField := reflect.Indirect(destVal).Field(fieldIndex)
		targetValue := reflect.Indirect(reflect.ValueOf(target))
		targetField.Set(targetValue)
	}

	// We're done! Now we can update our original target (dest) and return.
	reflect.Indirect(reflect.ValueOf(dest)).Set(reflect.Indirect(destVal))

	return nil
}
