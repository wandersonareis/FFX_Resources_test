package common

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Example structs and interfaces
type MyStruct struct {
	Name  string
	Age   int
	Date  time.Time
}

type MyInterface interface {
	Method()
}

// Tests with testify
func TestCheckArgumentNil(t *testing.T) {
	assert := assert.New(t)

	// Uninitialized interface
	assert.Panics(func() {
		var myInterface MyInterface
		CheckArgumentNil(myInterface, "myInterface")
	}, "Deveria ter panic para interface não inicializada")

	// Uninitialized pointer to struct
	assert.Panics(func() {
		var myStruct *MyStruct
		CheckArgumentNil(myStruct, "myStruct")
	}, "Deveria ter panic para ponteiro não inicializado")

	// Struct initialized with default values (não deve panic)
	myStruct2 := &MyStruct{}
	assert.NotPanics(func() {
		CheckArgumentNil(myStruct2, "myStruct2")
	}, "Não deveria ter panic para struct inicializada")

	// Uninitialized slice
	assert.Panics(func() {
		var mySlice []int
		CheckArgumentNil(mySlice, "mySlice")
	}, "Deveria ter panic para slice não inicializado")

	// Empty slice
	assert.Panics(func() {
		mySlice2 := []int{}
		CheckArgumentNil(mySlice2, "mySlice2")
	}, "Deveria ter panic para slice vazio")

	// Slice with elements (não deve panic)
	mySlice3 := []int{1, 2, 3}
	assert.NotPanics(func() {
		CheckArgumentNil(mySlice3, "mySlice3")
	}, "Não deveria ter panic para slice populado")

	// Primitive type (int) zero value
	assert.Panics(func() {
		i := 0
		CheckArgumentNil(i, "i")
	}, "Deveria ter panic para int zero value")

	// Primitive type (string) vazio
	assert.Panics(func() {
		str := ""
		CheckArgumentNil(str, "str")
	}, "Deveria ter panic para string vazia")
}