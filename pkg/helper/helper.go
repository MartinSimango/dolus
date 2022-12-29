package helper

import (
	"reflect"
	"strings"
)

func GetExportName(name string) string {
	name = strings.ReplaceAll(name, "-", "_")
	return strings.ToUpper(name[:1]) + name[1:]
}

func GetPointerToInterface(str any) any {
	return reflect.New(reflect.ValueOf(str).Type()).Interface()
}

func GetUnderlyingPointerValue(ptr any) any {
	return reflect.ValueOf(ptr).Elem().Interface()
}

func GetSliceType(value reflect.Value) reflect.Type {
	return reflect.TypeOf(value.Interface()).Elem()
}

func GetPointerToSliceType(sliceType reflect.Type) any {
	return reflect.New(sliceType).Elem().Interface()
}
