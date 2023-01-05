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
	Path       string
	Method     string
	StatusCode string
	schema     any
}

func New(path, method, statusCode string, ref *openapi3.ResponseRef, mediaType string,
) *ResponseSchema {
	return &ResponseSchema{
		Path:       path,
		Method:     method,
		StatusCode: statusCode,
		schema:     getSchema(ref, mediaType),
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

func structFromSchema(schema openapi3.Schema) any {
	dsb := dynamicstruct.NewStruct()

	for name, p := range schema.Properties {
		exportName := strings.ToUpper(name[:1]) + name[1:]
		tags, nullable := getTags(name, inList(name, schema.Required), *p.Value)
		switch p.Value.Type {
		case "object":
			internalStruct := structFromSchema(*p.Value)
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

func structFromExample(example openapi3.Examples) any {
	for _, v := range example {
		m := (v.Value.Value).(map[string]interface{})
		g, _ := buildExample(m, "", "")
		return reflect.New(reflect.ValueOf(g).Type()).Interface()
	}
	fmt.Println()
	return nil
}

func buildStructFromMap(_map any) (any, string) {
	dsb := dynamicstruct.NewStruct()
	for k, v := range _map.(map[string]interface{}) {
		exportName := getExportName(k)
		i, _type := buildExample(v, k, "")
		switch _type {
		case "string":
			dsb.AddField(exportName, i.(string), fmt.Sprintf(`json:"%s" type:"%s"`, k, _type))
		case "number":
			dsb.AddField(exportName, i.(float64), fmt.Sprintf(`json:"%s" type:"%s"`, k, _type))
		case "integer":
			dsb.AddField(exportName, int64(i.(float64)), fmt.Sprintf(`json:"%s" type:"%s"`, k, _type))
		case "boolean":
			dsb.AddField(exportName, i.(bool), fmt.Sprintf(`json:"%s" type:"%s"`, k, _type))
		case "slice":
			dsb.AddField(exportName, i, fmt.Sprintf(`json:"%s" type:"%s"`, k, _type))
		case "struct":
			dsb.AddField(exportName, i, fmt.Sprintf(`json:"%s" type:"%s"`, k, _type))
		}
	}
	return reflect.ValueOf(dsb.Build().New()).Elem().Interface(), "struct"
}

func buildSliceOfSliceElementType(config any, name string, root string) (any, string) {
	fullFieldName := name
	if root != "" {
		fullFieldName = fmt.Sprintf("%s.%s", root, name)
	}
	slice := config.([]interface{})

	var firstElement any
	if len(slice) == 0 {
		firstElement = "" //emtpy slice assume array of strings
	} else {
		firstElement, _ = buildExample(slice[0], name, "")
	}

	currentElement := firstElement
	for i := 1; i < len(slice); i++ {
		nextElement, _ := buildExample(slice[i], name, "")
		if reflect.ValueOf(nextElement).Kind() == reflect.Struct {
			var err error
			var mergedStruct *dstruct.DynamicStructModifier
			if mergedStruct, err = dstruct.MergeStructs(currentElement, nextElement, fullFieldName); err != nil {
				panic(err.Error())
			}
			currentElement = mergedStruct.Get()
			// Account for different types of elements that are not struct
		} else if reflect.TypeOf(nextElement) != reflect.TypeOf(firstElement) {
			currentElement = ""
			if reflect.ValueOf(firstElement).Kind() == reflect.Slice {
				currentElement = []string{}
			}
			break
		}
	}
	sliceOfElementType := reflect.SliceOf(reflect.ValueOf(currentElement).Type())
	return reflect.MakeSlice(sliceOfElementType, 0, 1024).Interface(), "slice"
}

func getType(element any, kind reflect.Kind) string {
	switch kind {
	case reflect.String:
		return "string"
	case reflect.Float64:
		val := element.(float64)
		if val-float64(int64(val)) == 0 {
			return "integer"
		} else {
			return "number"
		}

	case reflect.Bool:
		return "boolean"
	}
	return "unknown"
}

func buildExample(config interface{}, name string, root string) (interface{}, string) {

	if config == nil {
		return nil, "nil"
	}
	configKind := reflect.ValueOf(config).Kind()
	switch configKind {
	case reflect.Map:
		return buildStructFromMap(config)
	case reflect.Slice:
		return buildSliceOfSliceElementType(config, name, root)
	default:
		return config, getType(config, configKind)
	}

}

func getSchema(ref *openapi3.ResponseRef, mediaType string) any {
	content := ref.Value.Content.Get(mediaType)
	if content != nil { // TODO if no example response maybe be empty eg /v1/cancel/charge 200 response
		if content.Schema != nil {
			return structFromSchema(*content.Schema.Value)
		} else {
			return structFromExample(content.Examples)
		}
	}
	return structFromExample(content.Examples)
}

func getExportName(name string) string {
	// TODO replace special characters in name
	name = strings.ReplaceAll(name, "-", "_")
	return strings.ToUpper(name[:1]) + name[1:]
}
