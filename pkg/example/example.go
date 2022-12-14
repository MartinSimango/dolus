package example

import (
	"fmt"
	"reflect"
	"strings"
	"unsafe"

	"github.com/MartinSimango/dolus/pkg/generator"
	"github.com/MartinSimango/dolus/pkg/schema"
)

type Field struct {
	Type      reflect.Kind
	Address   uintptr
	SubFields []*Field
	Tags      reflect.StructTag
}

type Example struct {
	Value             schema.ResponseSchema
	fields            map[string]*Field // maps the memory location of fields
	GeneratedFields   map[string]generator.GenerationFunction
	GenerateAllFields bool
}

func New(responseSchema *schema.ResponseSchema) *Example {
	// TODO
	// if responseSchema == nil {
	// 	responseSchema = example.createSchema
	// }
	// generatedFields["status"] = generator.NewGenerationFunc(generator.GenerateStringFromRegex, "^[A-Z]{3}$")
	// gf["authorizedAmount.amount"] = generator.GenerateInteger
	// gf["authorizedAmount.code"] = generator.GenerateFixedString

	example := &Example{
		Value:           *responseSchema,
		fields:          make(map[string]*Field),
		GeneratedFields: make(map[string]generator.GenerationFunction),
	}
	example.populateFieldMap(responseSchema.Schema, "")

	return example
}

func setField[T any](f *Field, field string, value T) {
	*(*T)(unsafe.Pointer(f.Address)) = value
}

func getField[T any](f *Field, field string) T {
	return *(*T)(unsafe.Pointer(f.Address))
}

func (example *Example) GenerateFields() {
	for k, genFunc := range example.GeneratedFields {
		example.SetField(k, genFunc.Generate())
	}
}

func (example *Example) SetField(field string, value any) error {

	f := example.fields[field]
	if f == nil {
		return fmt.Errorf("No such field '%s' exists in schema", field)
	}
	switch f.Type {
	case reflect.String:
		setField(f, field, value.(string))
	case reflect.Int64:
		setField(f, field, value.(int64))
	case reflect.Float64:
		setField(f, field, value.(float64))

	default:
		panic(fmt.Sprintf("unsupported type '%s'", f.Type))
	}

	return nil
}

func (example *Example) GetField(field string) (any, error) {
	f := example.fields[field]
	if f == nil {
		return nil, fmt.Errorf("No such field '%s' exists in schema", field)
	}
	switch f.Type {
	case reflect.String:
		return getField[string](f, field), nil
	case reflect.Int64:
		return getField[int64](f, field), nil
	default:
		panic(fmt.Sprintf("unsupported type '%s'", f.Type))
	}

}

func (example *Example) populateFieldMap(config any, root string) (newFields []*Field) {
	if config == nil {
		return
	}
	inputConfig := reflect.ValueOf(config).Elem()
	for i := 0; i < inputConfig.NumField(); i++ {
		field := inputConfig.Field(i)
		fieldName := inputConfig.Type().Field(i).Name
		fieldName = strings.ToLower(fieldName[0:1]) + fieldName[1:]
		fieldTags := inputConfig.Type().Field(i).Tag
		if root != "" {
			fieldName = fmt.Sprintf("%s.%s", root, fieldName)

		}
		switch field.Kind() {
		case reflect.Struct:
			example.fields[fieldName] = &Field{
				Address:   inputConfig.Field(i).UnsafeAddr(),
				Type:      inputConfig.Field(i).Kind(),
				SubFields: example.populateFieldMap(field.Addr().Interface(), fieldName),
				Tags:      fieldTags,
			}
		default:
			example.fields[fieldName] = &Field{
				Address: inputConfig.Field(i).UnsafeAddr(),
				Type:    inputConfig.Field(i).Kind(),
				Tags:    fieldTags,
			}

			example.setDefaultValue(field, fieldTags, fieldName)
		}

		newFields = append(newFields, example.fields[fieldName])
	}
	return

}

func (example *Example) setDefaultValue(field reflect.Value, structTag reflect.StructTag, fieldName string) {
	if field.Kind() != reflect.Ptr {
		pattern := structTag.Get("pattern")
		if pattern != "" {
			example.GeneratedFields[fieldName] =
				generator.GenerateStringFromRegexFunc(pattern)
		} else if field.Kind() == reflect.String {
			example.GeneratedFields[fieldName] =
				generator.GenerateFixedStringFunc("string")

		} else if field.Kind() == reflect.Float64 {
			example.GeneratedFields[fieldName] =
				generator.GenerateFloatFunc(0, 10)
		}

	}

}
