package generator

import (
	"math/rand"

	"github.com/takahiromiyamoto/go-xeger"
)

type GenerationFunctionKind uint8

var (
	generateStringFromRegex GenerationFunc = GenerationFunc{
		Func: func(parameters ...any) any {
			regex := parameters[0].(string)
			x, err := xeger.NewXeger(regex)
			if err != nil {
				panic(err)
			}
			return x.Generate()
		},
	}

	generateInteger GenerationFunc = GenerationFunc{
		Func: func(parameters ...any) any {
			min := parameters[0].(int64)
			max := parameters[1].(int64)
			return min + (int64(rand.Float64() * float64(max-min)))
		},
	}

	generateFloat GenerationFunc = GenerationFunc{
		Func: func(parameters ...any) any {
			min := parameters[0].(float64)
			max := parameters[1].(float64)
			return min + (rand.Float64() * (max - min))
		},
	}

	generateFixedString GenerationFunc = GenerationFunc{
		Func: func(parameters ...any) any {
			return parameters[0].(string)
		},
	}
)

func GenerateStringFromRegexFunc(regex string) *GenerationFunc {
	f := generateStringFromRegex
	f.Args = []any{regex}
	return &f
}

func GenerateIntegerFunc(min, max int64) *GenerationFunc {
	f := generateInteger
	f.Args = []any{min, max}
	return &f
}

func GenerateFloatFunc(min, max float64) *GenerationFunc {
	f := generateFloat
	f.Args = []any{min, max}
	return &f
}

func GenerateFixedStringFunc(regex string) *GenerationFunc {
	f := generateFixedString
	f.Args = []any{regex}
	return &f
}
