package generator

import "reflect"

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

type FunctionValueConfig struct {
	ValueGenerationType ValueGenerationType
	ValueGeneratedCount int          // TODO should not be public
	ValueType           reflect.Kind // TODO should not be public
}

type GenerationDefaults map[reflect.Kind]GenerationFunction
