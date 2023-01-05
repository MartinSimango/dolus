package example

import (
	"fmt"
	"reflect"

	"github.com/MartinSimango/dolus/pkg/dstruct"
	"github.com/MartinSimango/dolus/pkg/generator"
	"github.com/MartinSimango/dolus/pkg/schema"
)

type ExampleConfig struct {
	function            generator.GenerationFunction
	FunctionValueConfig generator.FunctionValueConfig
}

type GenerationFields map[string]*ExampleConfig

type Example struct {
	Value                      *dstruct.DynamicStructModifier
	valueGenerationType        generator.ValueGenerationType
	generatedFields            GenerationFields
	defaultGenerationFunctions generator.GenerationDefaults
}

func (example *Example) initExampleDefaults() {
	example.defaultGenerationFunctions = make(generator.GenerationDefaults)
	example.defaultGenerationFunctions[reflect.String] = generator.GenerateStringFromRegexFunc("^[a-z ,.'-]+$") // generator.GenerateFixedStringFunc("testing")
	// example.exampleDefaults[reflect.Bool] = generator.ge

}

func NewExampleWithGenerationFields(responseSchema *schema.ResponseSchema,
	valueGenerationType generator.ValueGenerationType,
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

	example.initExampleDefaults()

	example.Value =
		dstruct.NewDynamicStructModifierWithFieldModifier(schemaCopy, example.initGenerationFunc)

	return example
}

func New(responseSchema *schema.ResponseSchema, valueGenerationType generator.ValueGenerationType) *Example {
	return NewExampleWithGenerationFields(responseSchema, valueGenerationType, make(GenerationFields))
}

func (example *Example) Get() interface{} {
	example.generateFields()
	return example.Value.Get()
}

func (example *Example) generateFields() {
	for k, genFunc := range example.generatedFields {
		switch genFunc.FunctionValueConfig.ValueGenerationType {
		case generator.GenerateOnce:
			if genFunc.FunctionValueConfig.ValueGeneratedCount > 0 {
				// NO need to regenerate value
				continue
			}

		}

		if err := example.Value.SetField(k, genFunc.function.Generate()); err != nil {
			fmt.Println(err)
		} else {
			genFunc.FunctionValueConfig.ValueGeneratedCount++
		}

	}
}

func (example *Example) SetFieldConfig(fieldName string, functionValueConfig generator.FunctionValueConfig) error {
	if !example.Value.DoesFieldExist(fieldName) {
		return fmt.Errorf("no such field '%s' exists in schema", fieldName)
	}
	if example.generatedFields[fieldName] == nil {
		example.generatedFields[fieldName] = &ExampleConfig{}
	}
	example.generatedFields[fieldName].FunctionValueConfig = functionValueConfig
	return nil
}

func (example *Example) initGenerationFunc(field *dstruct.Field) {
	if example.generatedFields[field.Name] != nil && example.generatedFields[field.Name].function != nil || field.Kind == reflect.Ptr { //field already has generated function
		return
	}
	if example.generatedFields[field.Name] == nil {
		example.generatedFields[field.Name] = &ExampleConfig{
			FunctionValueConfig: generator.FunctionValueConfig{
				ValueGenerationType: example.valueGenerationType,
			},
		}
	}
	example.generatedFields[field.Name].function = example.getGenerationFunction(field, example.generatedFields[field.Name].FunctionValueConfig)
}

func (example *Example) getGenerationFunction(field *dstruct.Field,
	functionValueConfig generator.FunctionValueConfig) generator.GenerationFunction {

	if generationFunction := generator.GetGenerationFunction(field, functionValueConfig); generationFunction != nil {
		return generationFunction
	}

	return example.defaultGenerationFunctions[field.Kind]
}
