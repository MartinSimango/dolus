package schema

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	dynamicstruct "github.com/ompluscator/dynamic-struct"
)

type ResponseSchema struct {
	Path       string
	Method     string
	StatusCode string
	Schema     interface{}
}

func New(path, method, statusCode string, ref *openapi3.ResponseRef, mediaType string) *ResponseSchema {
	return &ResponseSchema{
		Path:       path,
		Method:     method,
		StatusCode: statusCode,
		Schema:     getSchema(ref, mediaType),
	}

}

func (schema *ResponseSchema) MarshalSchema() (string, error) {
	bytes, err := json.Marshal(schema.Schema)
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
		exportName := strings.ToUpper(name[0:1]) + name[1:]
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

func structFromSchemaRef(schemaRef openapi3.SchemaRef) any {
	return nil
}

func getSchema(ref *openapi3.ResponseRef, mediaType string) any {
	content := ref.Value.Content.Get(mediaType)
	if content != nil {
		if content.Schema != nil {
			return structFromSchema(*content.Schema.Value)
		}
	}
	return nil
}
