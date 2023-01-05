package dstruct

import (
	"fmt"
	"reflect"

	"github.com/MartinSimango/dolus/pkg/helper"
	dynamicstruct "github.com/ompluscator/dynamic-struct"
)

// MergeStructs merges two structs
func MergeStructs(left, right interface{}, structName string) (*DynamicStructModifier, error) {
	if reflect.ValueOf(left).Kind() != reflect.Struct || reflect.ValueOf(right).Kind() != reflect.Struct {
		return nil, fmt.Errorf("failed to merged structs: Both interface types are not structs")
	}

	mergedStruct, err := updateCurrentElementSchema(left, right, reflect.Struct, structName)
	if err != nil {
		return nil, err
	}
	return New(helper.GetPointerToInterface(mergedStruct)), nil
}

// TODO clean this function up
func updateCurrentElementSchema(currentElement interface{}, newElement interface{}, parentKind reflect.Kind, root string) (any, error) {
	currentElementDStruct := New(helper.GetPointerToInterface(currentElement))
	currentElementDynamicStruct := currentElementDStruct._struct
	currentElementDynamicStructFields := currentElementDStruct.Fields
	newElementDynamicStructFields := New(helper.GetPointerToInterface(newElement)).Fields

	// struct to be returned
	newStruct := dynamicstruct.ExtendStruct(currentElementDynamicStruct)

	for k, v := range newElementDynamicStructFields {
		elementName := newElementDynamicStructFields[k].Tags.Get("json")
		cV := currentElementDynamicStructFields[k]
		fullFieldName := k
		if root != "" {
			fullFieldName = fmt.Sprintf("%s.%s", root, k)
		}
		if cV == nil {
			newStruct.AddField(helper.GetExportName(elementName), v.Value.Interface(), string(v.Tags))
			continue
		}
		if err := validateTypes(v.Value, cV.Value, fullFieldName); err != nil {
			return nil, err
		}

		if v.Kind == reflect.Slice {
			vSliceType := helper.GetSliceType(v.Value)
			cVSliceType := helper.GetSliceType(cV.Value)
			if err := validateSliceTypes(vSliceType, cVSliceType, v.Value, cV.Value, fullFieldName); err != nil {
				return nil, err
			}
			newStruct.RemoveField(helper.GetExportName(k))
			if cVSliceType.Kind() == reflect.Struct {

				newSliceTypeStruct, err := updateCurrentElementSchema(helper.GetPointerToSliceType(cVSliceType),
					helper.GetPointerToSliceType(vSliceType), reflect.Slice, fullFieldName)

				if err != nil {
					return nil, err
				}
				newStruct.AddField(helper.GetExportName(elementName), newSliceTypeStruct, "")
			} else {
				newStruct.AddField(helper.GetExportName(elementName), v.Value.Interface(), "")
			}

		} else if v.Kind == reflect.Struct {
			updatedSchema, err := updateCurrentElementSchema(currentElementDynamicStructFields[k].Value.Interface(), v.Value.Interface(), reflect.Struct, fullFieldName)
			if err != nil {
				return nil, err
			}
			newStruct.RemoveField(helper.GetExportName(elementName))
			newStruct.AddField(helper.GetExportName(elementName), updatedSchema, string(newElementDynamicStructFields[k].Tags))
		}

	}

	if parentKind == reflect.Slice {
		sliceOfElementType := reflect.SliceOf(reflect.ValueOf(newStruct.Build().New()).Elem().Type())
		return reflect.MakeSlice(sliceOfElementType, 0, 1024).Interface(), nil
	}
	return reflect.ValueOf(newStruct.Build().New()).Elem().Interface(), nil
}

func shouldTypeMatch(kind reflect.Kind) bool {
	if kind == reflect.Array || kind == reflect.Struct || kind == reflect.Slice {
		return false
	}
	return true
}

func validateTypes(v, cV reflect.Value, fullFieldName string) error {
	currentElementType := reflect.TypeOf(cV.Interface())
	newElementType := reflect.TypeOf(v.Interface())
	if shouldTypeMatch(v.Kind()) || shouldTypeMatch(cV.Kind()) {
		if currentElementType != newElementType {
			return fmt.Errorf("mismatching types for field '%s': %s and %s", fullFieldName, currentElementType, newElementType)
		}
	} else {
		if v.Kind() != cV.Kind() {
			return fmt.Errorf("mismatching types for field '%s': %s and %s", fullFieldName, currentElementType, newElementType)
		}
	}
	return nil
}

func validateSliceTypes(vSliceType, cVSliceType reflect.Type, v, cV reflect.Value, fullFieldName string) error {
	currentElementType := reflect.TypeOf(reflect.New(cVSliceType).Interface())
	newElementType := reflect.TypeOf(reflect.New(vSliceType).Interface())

	if shouldTypeMatch(vSliceType.Kind()) || shouldTypeMatch(cVSliceType.Kind()) {
		if currentElementType != newElementType {
			return fmt.Errorf("mismatching types for field '%s': %s and %s", fullFieldName, reflect.TypeOf(v.Interface()), reflect.TypeOf(cV.Interface()))
		}
	} else {
		if v.Kind() != cV.Kind() {
			return fmt.Errorf("mismatching types for field '%s': %s and %s", fullFieldName, reflect.TypeOf(v.Interface()), reflect.TypeOf(cV.Interface()))
		}
	}
	return nil
}
