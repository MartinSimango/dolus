package example

import (
	"reflect"

	"github.com/MartinSimango/dolus/pkg/dstruct"
	"github.com/MartinSimango/dolus/pkg/generator"
	"github.com/MartinSimango/dolus/pkg/schema"
)

type ValueGenerationType uint8

const (
	GenerateAll     ValueGenerationType = iota // will generate all field
	GenerateAllOnce                            // will generate all the fields
	UseDefaults
)

type GenerationFields map[string]generator.GenerationFunction

type Example struct {
	Value               *dstruct.DynamicStructModifier
	valueGenerationType ValueGenerationType
	generatedFields     GenerationFields
}

func NewExampleWithGenerationFields(responseSchema *schema.ResponseSchema,
	valueGenerationType ValueGenerationType,
	generationFields GenerationFields,
) *Example {
	// // TODO
	schemaCopy := responseSchema.GetSchema()
	if schemaCopy == nil {
		return nil
	}

	example := &Example{
		generatedFields:     generationFields,
		valueGenerationType: valueGenerationType,
	}
	var modifyFieldFunction dstruct.FieldModifier = nil
	if valueGenerationType != UseDefaults {
		modifyFieldFunction = example.initGenerationFunc
	}
	example.Value =
		dstruct.NewDynamicStructModifierWithFieldModifier(schemaCopy, responseSchema.FieldPrefix, modifyFieldFunction)

	if valueGenerationType == GenerateAllOnce {
		example.generateFields()
	}
	return example
}

func New(responseSchema *schema.ResponseSchema, valueGenerationType ValueGenerationType) *Example {
	return NewExampleWithGenerationFields(responseSchema, valueGenerationType, make(GenerationFields))
}

func (example *Example) Get() interface{} {
	if example.valueGenerationType == GenerateAll {
		// go through each field and generate a value of it
		example.generateFields()
	}
	return example.Value.Get()
}

func (example *Example) generateFields() {
	for k, genFunc := range example.generatedFields {
		// TODO LOG ERROR
		example.Value.SetField(k, genFunc.Generate())
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
		}

	}

}
