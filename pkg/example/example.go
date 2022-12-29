package example

import (
	"fmt"
	"reflect"

	"github.com/MartinSimango/dolus/pkg/dstruct"
	"github.com/MartinSimango/dolus/pkg/generator"
	"github.com/MartinSimango/dolus/pkg/schema"
)

type ValueGenerationType uint8

const (
	Generate     ValueGenerationType = iota // will generate all field
	GenerateOnce                            // will generate all the fields
	UseDefaults
)

type GenerationFields map[string]generator.GenerationFunction

type FixedFieldValues map[string]any

type Example struct {
	Value               *dstruct.DynamicStructModifier
	valueGenerationType ValueGenerationType
	generatedFields     GenerationFields
	fixedValues         FixedFieldValues
}

func NewExampleWithGenerationFields(responseSchema *schema.ResponseSchema,
	valueGenerationType ValueGenerationType,
	generationFields GenerationFields,
) *Example {

	schemaCopy := responseSchema.GetSchema()
	if schemaCopy == nil {
		return nil // no schema means we can't create an example
	}

	example := &Example{
		generatedFields:     generationFields,
		valueGenerationType: valueGenerationType,
	}

	example.Value =
		dstruct.NewDynamicStructModifierWithFieldModifier(schemaCopy,
			getFieldModifierFunction(valueGenerationType, example.initGenerationFunc))

	if valueGenerationType == GenerateOnce {
		example.generateFields()
	}
	return example
}

func New(responseSchema *schema.ResponseSchema, valueGenerationType ValueGenerationType) *Example {
	return NewExampleWithGenerationFields(responseSchema, valueGenerationType, make(GenerationFields))
}

func (example *Example) Get() interface{} {
	if example.valueGenerationType == Generate {
		example.generateFields()
	}
	return example.Value.Get()
}

func (example *Example) generateFields() {
	for k, genFunc := range example.generatedFields {
		// TODO LOG ERROR
		if err := example.Value.SetField(k, genFunc.Generate()); err != nil {
			fmt.Println(err)
		}
	}
}

func (example *Example) initGenerationFunc(field *dstruct.Field) {
	if example.generatedFields[field.Name] != nil { //field already has generated function
		return
	}
	if field.Kind != reflect.Ptr {
		pattern := field.Tags.Get("pattern")
		if pattern != "" {
			example.generatedFields[field.Name] =
				generator.GenerateStringFromRegexFunc(pattern)
		} else if field.Kind == reflect.String {
			example.generatedFields[field.Name] =
				generator.GenerateFixedStringFunc("string")

		} else if field.Kind == reflect.Float64 {
			example.generatedFields[field.Name] =
				generator.GenerateFloatFunc(0, 10)
		} else if field.Kind == reflect.Slice {
			sliceType := reflect.TypeOf(field.Value.Interface()).Elem()
			example.generatedFields[field.Name] = generator.GenerateSlice(sliceType, 1, 2)
			// // TODO fixed
			// example.generatedFields[field.Name] = generator.GenerateFixedSlice(example.fixedValues[field.Name])
		}
	}

}

func getFieldModifierFunction(valueGenerationType ValueGenerationType,
	modifierFunction dstruct.FieldModifier) dstruct.FieldModifier {
	if valueGenerationType != UseDefaults {
		return modifierFunction
	}
	return nil
}
