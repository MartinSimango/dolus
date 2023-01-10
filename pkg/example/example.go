package example

import (
	"fmt"

	"github.com/MartinSimango/dolus/pkg/dstruct"
	"github.com/MartinSimango/dolus/pkg/generator"
	"github.com/MartinSimango/dolus/pkg/schema"
)

type ExampleFieldConfig struct {
	generationUnit generator.GenerationUnit
}

type GenerationFields map[string]*ExampleFieldConfig

type Example struct {
	Value            *dstruct.DynamicStructModifier
	GenerationConfig generator.GenerationConfig
	generatedFields  GenerationFields
}

func NewExampleWithGenerationFields(responseSchema *schema.ResponseSchema,
	generatedFields GenerationFields,
	generationConfig generator.GenerationConfig,

) *Example {

	schemaCopy := responseSchema.GetSchema()
	if schemaCopy == nil {
		return nil // no schema means we can't create an example
	}

	example := &Example{
		GenerationConfig: generationConfig,
		generatedFields:  generatedFields,
	}

	example.Value =
		dstruct.NewDynamicStructModifierWithFieldModifier(schemaCopy, example.initGenerationFunc)

	return example
}

func New(responseSchema *schema.ResponseSchema, config generator.GenerationConfig) *Example {
	return NewExampleWithGenerationFields(responseSchema, make(GenerationFields), config)
}

func (example *Example) Get() interface{} {
	example.generateFields()
	return example.Value.Get()
}

func (example *Example) generateFields() {
	for k, genFunc := range example.generatedFields {
		if err := example.Value.SetFieldValue(k, genFunc.generationUnit.Generate()); err != nil {
			fmt.Println(err)
		}
	}
}

func (example *Example) setFieldGenerationConfig(fieldName string, functionValueConfig generator.GenerationConfig) {
	field := example.Value.GetField(fieldName)

	example.generatedFields[fieldName].generationUnit.GenerationConfig = functionValueConfig
	example.generatedFields[fieldName].generationUnit.CurrentFunction = generator.GetGenerationFunction(field, functionValueConfig)
}

func (example *Example) GetFieldGenerationConfig(fieldName string) (*generator.GenerationConfig, error) {
	if example.generatedFields[fieldName] == nil {
		return nil, fmt.Errorf("no such field '%s' exists in schema", fieldName)
	}
	return &example.generatedFields[fieldName].generationUnit.GenerationConfig, nil
}

func (example *Example) Update(fieldName string) error {
	field := example.Value.GetField(fieldName)
	if field == nil {
		return fmt.Errorf("no such field '%s' exists in schema", fieldName)
	}
	example.setFieldGenerationConfig(fieldName, example.generatedFields[fieldName].generationUnit.GenerationConfig)
	return nil
}

func (example *Example) initGenerationFunc(field *dstruct.Field) {
	example.generatedFields[field.Name] = &ExampleFieldConfig{
		generationUnit: generator.GenerationUnit{
			CurrentFunction: generator.GetGenerationFunction(field,
				example.GenerationConfig),
			GenerationConfig: example.GenerationConfig,
		},
	}

}
