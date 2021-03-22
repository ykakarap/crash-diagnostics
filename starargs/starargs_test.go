package starargs

import (
	"fmt"
	"go.starlark.net/starlark"
	"testing"
)

type Simple struct {
	Val string `starargs:"args1"`
}

func TestSimple(t *testing.T) {
	args := starlark.Tuple{}
	arg1 := starlark.String("args1")
	val1 := starlark.String("value1")
	kwargs := []starlark.Tuple{
		starlark.Tuple([]starlark.Value{arg1, val1}),
	}
	fname := "simple_func"
	s := &Simple{}
	err := UnpackArgs(fname, args, kwargs, s)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("%+v\n", s)
}
