package dstruct

import (
	"fmt"
	"reflect"
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
	_struct       any      // the struct that actually stores the data
	prefix        string
	fieldModifier FieldModifier
}

func New(dstruct any, prefix string) *DynamicStructModifier {
	return NewDynamicStructModifierWithFieldModifier(dstruct, prefix, nil)
}

func NewDynamicStructModifierWithFieldModifier(dstruct any, prefix string, fieldModifier FieldModifier) *DynamicStructModifier {
	ds := &DynamicStructModifier{
		_struct:       dstruct,
		Fields:        make(FieldMap),
		prefix:        prefix,
		fieldModifier: fieldModifier,
	}

	ds.populateFieldMap(dstruct, "")
	return ds
}

func (ds *DynamicStructModifier) populateFieldMap(config any, root string) (newFields []*Field) {
	if config == nil {
		return
	}

	inputConfig := reflect.ValueOf(config).Elem()
	for i := 0; i < inputConfig.NumField(); i++ {
		field := inputConfig.Field(i)
		fieldName := inputConfig.Type().Field(i).Name[len(ds.prefix):] // remove prefix from the fieldName
		fieldTags := inputConfig.Type().Field(i).Tag
		if root != "" {
			fieldName = fmt.Sprintf("%s.%s", root, fieldName)
		}
		switch field.Kind() {
		case reflect.Struct:
			ds.Fields[fieldName] = &Field{
				Name:      fieldName,
				Kind:      inputConfig.Field(i).Kind(),
				SubFields: ds.populateFieldMap(field.Addr().Interface(), fieldName),
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

		newFields = append(newFields, ds.Fields[fieldName])
	}
	return

}

func (ds *DynamicStructModifier) Get() any {
	return getUnderlyingPointerValue(ds._struct)
}

func (ds *DynamicStructModifier) SetField(field string, value any) error {

	f := ds.Fields[field]
	if f == nil {
		return fmt.Errorf("No such field '%s' exists in schema", field)
	}
	switch f.Kind {
	case reflect.String:
		f.Value.SetString(value.(string))
	case reflect.Int64:
		f.Value.SetInt(value.(int64))
	case reflect.Float64:
		f.Value.SetFloat(value.(float64))
	default:
		panic(fmt.Sprintf("unsupported type '%s'", f.Kind))
	}

	return nil
}

func (ds *DynamicStructModifier) GetField(field string) (any, error) {
	f := ds.Fields[field]
	if f == nil {
		return nil, fmt.Errorf("No such field '%s' exists in schema", field)
	}
	switch f.Kind {
	case reflect.String:
		return f.Value.String(), nil
	case reflect.Int64:
		return f.Value.Int(), nil
	case reflect.Slice:
		return f.Value.Slice(0, reflect.ValueOf(f.Value.Interface()).Len()).Interface(), nil
	default:
		panic(fmt.Sprintf("unsupported type '%s'", f.Kind))
	}
}

// func setField[T any](f *Field, field string, value T) {
// 	*(*T)(unsafe.Pointer(f.Address)) = value
// }

// func getField[T any](f *Field, field string) T {
// 	return *(*T)(unsafe.Pointer(f.Address))
// }
