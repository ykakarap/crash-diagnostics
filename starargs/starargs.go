package starargs

import (
	"fmt"
	"go.starlark.net/starlark"
	"reflect"
	"strings"
)

const tagname = "starargs"

// UnpackARgs unpacks args and kwargs into v. v should be pointer to a struct with
// "starargs" fieldtags. These field tags defined the starlark argument names
func UnpackArgs(fname string, args starlark.Tuple, kwargs []starlark.Tuple, v interface{}) error {
	pairs := []interface{}{}
	argMap := map[string]starlark.Value{}
	val := reflect.ValueOf(v)
	for i := 0; i < val.Elem().NumField(); i++ {
		var sval starlark.String
		tagVal := val.Elem().Type().Field(i).Tag.Get(tagname)
		tagVal = strings.TrimSpace(tagVal)
		argMap[tagVal] = &sval
		pairs = append(pairs, tagVal, &sval)
	}
	if err := starlark.UnpackArgs(fname, args, kwargs, pairs...); err != nil {
		return err
	}
	for i := 0; i < val.Elem().NumField(); i++ {
		tagVal := strings.TrimSpace(val.Elem().Type().Field(i).Tag.Get(tagname))
		starVal := argMap[tagVal]
		starString := starVal.(*starlark.String)
		realString := string(*starString)
		fmt.Println(realString)
		val.Elem().Field(i).Set(reflect.ValueOf(realString))
	}
	return nil
}
