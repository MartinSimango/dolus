package dstruct

import (
	"fmt"
	"reflect"

	dynamicstruct "github.com/ompluscator/dynamic-struct"
)

type StructProperties struct {
	Name   string
	Prefix string
}

// MergeStructs merges two structs
func MergeStructs(left, right interface{}, properties StructProperties) (*DynamicStructModifier, error) {
	if reflect.ValueOf(left).Kind() != reflect.Struct || reflect.ValueOf(right).Kind() != reflect.Struct {
		return nil, fmt.Errorf("Failed to merged structs: Both interface types are not structs")
	}
	mergedStruct, err := updateCurrentElementSchema(left, right, reflect.Struct, properties.Name)
	if err != nil {
		return nil, err
	}
	return New(getPointerToInterface(mergedStruct), properties.Prefix), nil
}

func updateCurrentElementSchema(currentElement interface{}, newElement interface{}, parentKind reflect.Kind, root string) (any, error) {
	currentElementDStruct := New(getPointerToInterface(currentElement), "Dolus_")
	currentElementDynamicStruct := currentElementDStruct._struct
	currentElementDynamicStructFields := currentElementDStruct.Fields
	newElementDynamicStructFields := New(getPointerToInterface(newElement), "Dolus_").Fields

	// struct to be returned
	newStruct := dynamicstruct.ExtendStruct(currentElementDynamicStruct)

	for k, v := range newElementDynamicStructFields {
		cV := currentElementDynamicStructFields[k]
		fullFieldName := k
		if root != "" {
			fullFieldName = fmt.Sprintf("%s.%s", root, k)
		}
		if currentElementDynamicStructFields[k] == nil {
			newStruct.AddField("Dolus_"+k, v.Value.Interface(), string(v.Tags))
			continue
		}
		if err := validateTypes(v.Value, cV.Value, fullFieldName); err != nil {
			return nil, err
		}

		if v.Kind == reflect.Slice {
			vSliceType := getSliceType(v.Value)
			cVSliceType := getSliceType(cV.Value)
			if err := validateSliceTypes(vSliceType, cVSliceType, v.Value, cV.Value, fullFieldName); err != nil {
				return nil, err
			}
			newStruct.RemoveField("Dolus_" + k)
			if cVSliceType.Kind() == reflect.Struct {

				newSliceTypeStruct, err := updateCurrentElementSchema(getPointerToSliceType(cVSliceType), getPointerToSliceType(vSliceType), reflect.Slice, fullFieldName)

				if err != nil {
					return nil, err
				}
				newStruct.AddField("Dolus_"+k, newSliceTypeStruct, "")
			} else {
				newStruct.AddField("Dolus_"+k, v.Value.Interface(), "")

			}

		}

		if v.Kind == reflect.Struct {
			n, err := updateCurrentElementSchema(currentElementDynamicStructFields[k].Value.Interface(), v.Value.Interface(), reflect.Struct, fullFieldName)
			if err != nil {
				return nil, err
			}
			dS := New(n, "Dolus_")
			newStruct.RemoveField("Dolus_" + k)
			newStruct.AddField("Dolus_"+k, n, string(dS.Fields[k].Tags))
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
			return fmt.Errorf("Mismatching types for field '%s': %s and %s", fullFieldName, currentElementType, newElementType)
		}
	} else {
		if v.Kind() != cV.Kind() {
			return fmt.Errorf("Mismatching types for field '%s': %s and %s", fullFieldName, currentElementType, newElementType)
		}
	}
	return nil
}

func validateSliceTypes(vSliceType, cVSliceType reflect.Type, v, cV reflect.Value, fullFieldName string) error {
	currentElementType := reflect.TypeOf(reflect.New(cVSliceType).Interface())
	newElementType := reflect.TypeOf(reflect.New(vSliceType).Interface())

	if shouldTypeMatch(vSliceType.Kind()) || shouldTypeMatch(cVSliceType.Kind()) {
		if currentElementType != newElementType {
			return fmt.Errorf("Mismatching types for field '%s': %s and %s", fullFieldName, reflect.TypeOf(v.Interface()), reflect.TypeOf(cV.Interface()))
		}
	} else {
		if v.Kind() != cV.Kind() {
			return fmt.Errorf("Mismatching types for field '%s': %s and %s", fullFieldName, reflect.TypeOf(v.Interface()), reflect.TypeOf(cV.Interface()))
		}
	}
	return nil
}
