package dstruct

import (
	"fmt"
	"reflect"

	"github.com/MartinSimango/dolus/pkg/helper"
)

// An extension of "github.com/ompluscator/dynamic-struct" to help modify dynamic structs

type Field struct {
	Name      string
	Kind      reflect.Kind
	SubFields []*Field
	Tags      reflect.StructTag
	Value     reflect.Value
}

type FieldMap map[string]*Field

type FieldModifier func(*Field)

type DynamicStructModifier struct {
	Fields        FieldMap // holds the meta data for the fields
	allFields     FieldMap
	_struct       any // the struct that actually stores the data
	fieldModifier FieldModifier
}

func New(dstruct any) *DynamicStructModifier {
	return NewDynamicStructModifierWithFieldModifier(dstruct, nil)
}

func NewDynamicStructModifierWithFieldModifier(dstruct any, fieldModifier FieldModifier) *DynamicStructModifier {
	ds := &DynamicStructModifier{
		_struct:       dstruct,
		Fields:        make(FieldMap),
		allFields:     make(FieldMap),
		fieldModifier: fieldModifier,
	}
	ds.populateFieldMap(ds._struct, "", ds.allFields)

	return ds
}

// TODO clean this up
func (ds *DynamicStructModifier) populateFieldMap(config any, root string, allFields FieldMap) (newFields []*Field) {
	if config == nil {
		return
	}

	inputConfig := reflect.ValueOf(config).Elem()

	for i := 0; i < inputConfig.NumField(); i++ {
		field := inputConfig.Field(i)
		fieldTags := inputConfig.Type().Field(i).Tag
		fieldName := fieldTags.Get("json")
		if root != "" {
			fieldName = fmt.Sprintf("%s.%s", root, fieldName)
		}
		switch field.Kind() {
		case reflect.Struct:
			subStruct := &DynamicStructModifier{
				_struct:       field.Addr().Interface(),
				Fields:        make(FieldMap),
				fieldModifier: ds.fieldModifier,
			}
			ds.Fields[fieldName] = &Field{
				Name:      fieldName,
				Kind:      inputConfig.Field(i).Kind(),
				SubFields: subStruct.populateFieldMap(subStruct._struct, fieldName, allFields),
				Tags:      fieldTags,
				Value:     field,
			}

		default:
			ds.Fields[fieldName] = &Field{
				Name:  fieldName,
				Kind:  inputConfig.Field(i).Kind(),
				Tags:  fieldTags,
				Value: field,
			}
			if ds.fieldModifier != nil {
				ds.fieldModifier(ds.Fields[fieldName])
			}
		}

		allFields[fieldName] = ds.Fields[fieldName]
		newFields = append(newFields, ds.Fields[fieldName])

	}
	return

}

func (ds *DynamicStructModifier) Get() any {
	return helper.GetUnderlyingPointerValue(ds._struct)
}

func (ds *DynamicStructModifier) SetField(field string, value any) error {
	f := ds.allFields[field]
	if f == nil {
		return fmt.Errorf("no such field '%s' exists in schema", field)
	}
	switch f.Kind {
	case reflect.String:
		f.Value.SetString(value.(string))
	case reflect.Int64:
		f.Value.SetInt(value.(int64))
	case reflect.Float64:
		f.Value.SetFloat(value.(float64))
	case reflect.Slice:
		f.Value.Set(reflect.ValueOf(value))
	default:
		panic(fmt.Sprintf("unsupported type '%s'", f.Kind))
	}

	return nil
}

func (ds *DynamicStructModifier) GetField(field string) (any, error) {
	f := ds.allFields[field]
	if f == nil {
		return nil, fmt.Errorf("no such field '%s' exists in schema", field)
	}

	switch f.Kind {
	case reflect.String:
		return f.Value.String(), nil
	case reflect.Int64:
		return f.Value.Int(), nil
	case reflect.Slice:
		sliceLen := reflect.ValueOf(f.Value.Interface()).Len()
		return f.Value.Slice(0, sliceLen).Interface(), nil
	default:
		panic(fmt.Sprintf("unsupported type '%s'", f.Kind))
	}
}

func (ds *DynamicStructModifier) DoesFieldExist(field string) bool {
	return ds.allFields[field] != nil
}

// func setField[T any](f *Field, field string, value T) {
// 	*(*T)(unsafe.Pointer(f.Address)) = value
// }

// func getField[T any](f *Field, field string) T {
// 	return *(*T)(unsafe.Pointer(f.Address))
// }
