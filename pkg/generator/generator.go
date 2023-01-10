package generator

import (
	"reflect"
)

type GenerationFunction interface {
	Generate() any
}

type GenerationFunc struct {
	_func func(...any) any
	args  []any
}

var _ GenerationFunction = &GenerationFunc{}

func (f GenerationFunc) Generate() any {
	return f._func(f.args...)
}

type ValueGenerationType uint8

const (
	Generate     ValueGenerationType = iota // will generate all field
	GenerateOnce                            // will generate all the fields
	UseDefaults
)

type GenerationDefaults map[reflect.Kind]GenerationFunction

type GenerationConfig struct {
	DefaultGenerationFunctions GenerationDefaults
	ValueGenerationType        ValueGenerationType
	SetNonRequiredFields       bool
	SliceMinLength             int
	SliceMaxLength             int
	FloatMin                   float64
	FloatMax                   float64
	IntMin                     int64
	IntMax                     int64
}

func NewGenerationConfig() (generationConfig *GenerationConfig) {
	generationConfig = &GenerationConfig{
		ValueGenerationType:  Generate,
		SetNonRequiredFields: false,
		SliceMinLength:       0,
		SliceMaxLength:       10,
	}
	generationConfig.initGenerationFunctionDefaults()
	return
}

func (gc *GenerationConfig) initGenerationFunctionDefaults() {
	gc.DefaultGenerationFunctions = make(GenerationDefaults)
	gc.DefaultGenerationFunctions[reflect.String] = GenerateFixedStringFunc("string") //generator.GenerateStringFromRegexFunc("^[a-z ,.'-]+$")
	gc.DefaultGenerationFunctions[reflect.Ptr] = GenerateNilValue()
	gc.DefaultGenerationFunctions[reflect.Int64] = GenerateIntegerFunc(gc.IntMin, gc.IntMax)
	gc.DefaultGenerationFunctions[reflect.Float64] = GenerateFloatFunc(gc.FloatMin, gc.FloatMin)
	gc.DefaultGenerationFunctions[reflect.Bool] = GenerateBoolFunc()

}

type GenerationUnit struct {
	CurrentFunction  GenerationFunction
	GenerationConfig GenerationConfig
	count            int
	latestValue      any
}

func (g *GenerationUnit) Generate() any {
	if g.GenerationConfig.ValueGenerationType == GenerateOnce && g.count > 0 {
		return g.latestValue
	}
	g.latestValue = g.CurrentFunction.Generate()
	g.count++
	return g.latestValue
}
