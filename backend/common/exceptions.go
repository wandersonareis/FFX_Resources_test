package common

import (
	"errors"
	"fmt"
	"reflect"
)

// errArgumentNil is returned when the argument is nil
var errArgumentNil = errors.New("argument is nil")

// errEmptyArgument is returned when the argument is the zero value for its type
var errEmptyArgument = errors.New("argument is zero value")

func argumentIsNil(name string) error {
	return fmt.Errorf("%s: %w", name, errArgumentNil)
}

func argumentIsEmpty(name string) error {
	return fmt.Errorf("%s: %w", name, errEmptyArgument)
}

// isNilable checks if a type can be nil
func isNilable(kind reflect.Kind) bool {
	switch kind {
	case reflect.Ptr, reflect.Interface, reflect.Slice, reflect.Map, reflect.Chan, reflect.Func:
		return true
	default:
		return false
	}
}

// validateStruct checks if a struct is initialized (non-nil pointer)
func validateStruct(arg interface{}) error {
	val := reflect.ValueOf(arg)
	if val.Kind() != reflect.Ptr || val.IsNil() {
		return argumentIsNil("struct")
	}
	return nil
}

// validateSlice checks if a slice is nil or empty
func validateSlice(arg interface{}) error {
	val := reflect.ValueOf(arg)
	if val.Kind() != reflect.Slice || val.IsNil() || val.Len() == 0 {
		return argumentIsNil("slice")
	}
	return nil
}

// validatePrimitive checks if a primitive type is the zero value
func validatePrimitive(arg interface{}) error {
	val := reflect.ValueOf(arg)
	if val.IsZero() {
		return argumentIsEmpty("primitive type")
	}
	return nil
}

// CheckArgumentNil checks if the argument is nil or a zero value (for interfaces, pointers, slices, maps, etc.).
// If the argument is invalid, the function will panic.
//
// Usage examples:
//
// 1. Interfaces:
//    var i interface{}
//    CheckArgumentNil(i, "i") // panics if 'i' is nil
//
// 2. Pointers to structs:
//    s := &MyStruct{}
//    CheckArgumentNil(s, "s") // does not panic if 's' is valid
//
// 3. Uninitialized pointers to structs:
//    var s *MyStruct
//    CheckArgumentNil(s, "s") // panics if 's' is nil
//
// 4. Slices:
//    s := []int{}
//    CheckArgumentNil(s, "s") // panics if 's' is nil or empty
//
// 5. Uninitialized slices:
//    var s []int
//    CheckArgumentNil(s, "s") // panics if 's' is nil
//
// 6. Primitive types:
//    i := 0
//    CheckArgumentNil(i, "i") // panics if 'i' is zero value
func CheckArgumentNil(arg interface{}, name string) {
    val := reflect.ValueOf(arg)

    if !val.IsValid() || (isNilable(val.Kind()) && val.IsNil()) {
        panic(argumentIsNil(name))
    }

    switch val.Kind() {
    case reflect.Struct:
        if err := validateStruct(arg); err != nil {
            panic(err)
        }
    case reflect.Slice:
        if err := validateSlice(arg); err != nil {
            panic(err)
        }
    default:
        if err := validatePrimitive(arg); err != nil {
            panic(err)
        }
    }
}