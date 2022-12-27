package generator

type GenerationFunction interface {
	Generate() any
}

type GenerationFunc struct {
	Func func(...any) any
	Args []any
}

var _ GenerationFunction = &GenerationFunc{}

func (f GenerationFunc) Generate() any {
	return f.Func(f.Args...)
}
