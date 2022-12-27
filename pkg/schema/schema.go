package schema

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/MartinSimango/dolus/pkg/dstruct"
	"github.com/getkin/kin-openapi/openapi3"
	dynamicstruct "github.com/ompluscator/dynamic-struct"
)

type ResponseSchema struct {
	Path        string
	Method      string
	StatusCode  string
	FieldPrefix string
	schema      any
}

func New(path, method, statusCode string, ref *openapi3.ResponseRef, mediaType string, fieldPrefix string,
) *ResponseSchema {
	return &ResponseSchema{
		Path:        path,
		Method:      method,
		StatusCode:  statusCode,
		schema:      getSchema(ref, mediaType, fieldPrefix),
		FieldPrefix: fieldPrefix,
	}

}

func (rs *ResponseSchema) GetSchema() any {
	// Make copy of schema to use as struct that is being modified to not modify original schema
	if rs.schema == nil {
		return nil
	}
	schemaValue := reflect.ValueOf(rs.schema).Elem().Interface()
	return reflect.New(reflect.ValueOf(schemaValue).Type()).Interface()
}

func (schema *ResponseSchema) MarshalSchema() (string, error) {
	bytes, err := json.Marshal(schema.schema)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func inList(key string, list []string) bool {
	for k := range list {
		if list[k] == key {
			return true
		}
	}
	return false
}

func getTags(name string, requiredField bool, schema openapi3.Schema) (string, bool) {
	nullable := "true"
	if !schema.Nullable {
		nullable = "false"
	}

	required := "false"
	if requiredField {
		required = "true"
	}
	pointer := nullable == "true" || required == "false"
	return fmt.Sprintf(`json:"%s" type:"%s" pattern:"%s" required:%s nullable:%s`, name, schema.Type, schema.Pattern, required, nullable), pointer
}

func addField[T any](name string, tags string, nullable bool, builder *dynamicstruct.Builder) {
	if nullable {
		(*builder).AddField(name, new(T), tags)
	} else {
		(*builder).AddField(name, *new(T), tags)
	}
}

func structFromSchema(schema openapi3.Schema, fieldPrefix string) any {
	dsb := dynamicstruct.NewStruct()

	for name, p := range schema.Properties {
		exportName := fieldPrefix + name
		tags, nullable := getTags(name, inList(name, schema.Required), *p.Value)
		switch p.Value.Type {
		case "object":
			internalStruct := structFromSchema(*p.Value, fieldPrefix)
			dsb.AddField(exportName, reflect.ValueOf(internalStruct).Elem().Interface(), tags)
		case "string":
			addField[string](exportName, tags, nullable, &dsb)
		case "number":
			addField[float64](exportName, tags, nullable, &dsb)
		default:
			panic(fmt.Sprintf("Unsupported for type '%s'", p.Value.Type))
		}

	}
	return dsb.Build().New()
}

func structFromExample(example openapi3.Examples) dynamicstruct.DynamicStruct {

	for _, v := range example {
		m := (v.Value.Value).(map[string]interface{})
		for _, v := range m {
			fmt.Println(v)
			g, _ := buildExample(m, "", reflect.ValueOf(m).Kind(), "")
			// a, _ := json.MarshalIndent(g, "", "\t")
			fmt.Println()
			fmt.Printf("AG: %+v\n", g)
			f := dstruct.New(reflect.New(reflect.ValueOf(g).Type()).Interface(), "Dolus_")

			// fmt.Println(reflect.New())
			fmt.Printf("%+v\n", f.Get())
			fmt.Println(reflect.TypeOf(g))
		}

	}
	fmt.Println()
	return nil
}

func buildExample(config interface{}, name string, parentType reflect.Kind, root string) (interface{}, *string) {
	dsb := dynamicstruct.NewStruct()
	fullFieldName := name
	if root != "" {
		fullFieldName = fmt.Sprintf("%s.%s", root, name)
	}

	if config == nil {
		return nil, nil
	}
	configKind := reflect.ValueOf(config).Kind()
	//  TODO remember arrays of defaults are different from arrays of maps etcs..
	switch configKind {

	case reflect.Map:
		for k, v := range config.(map[string]interface{}) {
			eName := "Dolus_" + strings.ReplaceAll(k, "-", "_")
			i, _type := buildExample(v, k, configKind, "") // 3.0 id map
			if _type != nil {
				switch *_type {
				case "string":
					dsb.AddField(eName, i.(string), fmt.Sprintf(`json:"%s" type:"%s"`, k, *_type))
				case "number":
					dsb.AddField(eName, i.(float64), fmt.Sprintf(`json:"%s" type:"%s"`, k, *_type))
				case "integer":
					dsb.AddField(eName, int64(i.(float64)), fmt.Sprintf(`json:"%s" type:"%s"`, k, *_type))

				case "boolean":
					dsb.AddField(eName, i.(bool), fmt.Sprintf(`json:"%s" type:"%s"`, k, *_type))
				case "slice":
					dsb.AddField(eName, i, fmt.Sprintf(`json:"%s"`, k))

				}
			} else {
				dsb.AddField(eName, i, fmt.Sprintf(`json:"%s"`, k))
			}
		}
	case reflect.Slice:
		// TODO remember to deal with slices within slices
		// if one of the elements is a slice and the rest aren't then array is array of strings
		// var elements []any
		// elements should be part of the same struct
		slice := config.([]interface{})
		if len(slice) == 0 {
			return nil, nil
		}
		currentElement, _ := buildExample(slice[0], name, configKind, "")
		// sliceElement := elementType.Type == "slice"
		originalElement := currentElement
		for i := 1; i < len(slice); i++ {
			newElement, _ := buildExample(slice[i], name, configKind, "")
			// TODO also check that currentElement is struct
			if reflect.ValueOf(newElement).Kind() == reflect.Struct {
				var err error
				var mergedStruct *dstruct.DynamicStructModifier
				if mergedStruct, err = dstruct.MergeStructs(currentElement, newElement, dstruct.StructProperties{
					Name:   fullFieldName,
					Prefix: "Dolus_",
				}); err != nil {
					panic(err.Error())
				}
				currentElement = mergedStruct.Get()
				// Account for different types of elements that are not struct
			} else if reflect.TypeOf(newElement) != reflect.TypeOf(originalElement) {
				currentElement = ""
				if reflect.ValueOf(originalElement).Kind() == reflect.Slice {
					currentElement = []string{}
				}
				break

			}

		}
		sliceOfElementType := reflect.SliceOf(reflect.ValueOf(currentElement).Type())
		_type := "slice"
		return reflect.MakeSlice(sliceOfElementType, 0, 1024).Interface(), &_type

	default:
		t := "unknown"
		switch configKind {
		case reflect.String:
			t = "string"
		case reflect.Float64:
			val := config.(float64)
			if val-float64(int64(val)) == 0 {
				t = "integer"
			} else {
				t = "number"
			}

		case reflect.Bool:
			t = "boolean"
		}
		return config, &t
	}
	t := "struct"
	return reflect.ValueOf(dsb.Build().New()).Elem().Interface(), &t

}

func getSchema(ref *openapi3.ResponseRef, mediaType string, fieldPrefix string) any {
	content := ref.Value.Content.Get(mediaType)
	if content != nil { // TODO if no example response maybe be empty eg /v1/cancel/charge 200 response
		if content.Schema != nil {
			return structFromSchema(*content.Schema.Value, fieldPrefix)
		} else {
			return structFromExample(content.Examples)
		}
	}
	return nil
}
