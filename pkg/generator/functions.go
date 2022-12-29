package generator

import (
	"math/rand"
	"reflect"

	"github.com/takahiromiyamoto/go-xeger"
)

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
			return min + (int64(rand.Float64() * float64(max+1-min)))
		},
	}

	generateFloat GenerationFunc = GenerationFunc{
		Func: func(parameters ...any) any {
			min := parameters[0].(float64)
			max := parameters[1].(float64)
			return min + (rand.Float64() * (max + 1 - min))
		},
	}

	generateFixedString GenerationFunc = GenerationFunc{
		Func: func(parameters ...any) any {
			return parameters[0].(string)
		},
	}

	generateObject GenerationFunc = GenerationFunc{
		Func: func(parameters ...any) any {
			return nil
		},
	}

	generateSlice GenerationFunc = GenerationFunc{
		Func: func(parameters ...any) any {
			sliceType := parameters[0].(reflect.Type)
			min := parameters[1].(int)
			max := parameters[2].(int)

			len := min + (int(rand.Float64() * float64(max+1-min)))
			sliceOfElementType := reflect.SliceOf(sliceType)
			slice := reflect.MakeSlice(sliceOfElementType, 0, 1024)
			switch sliceType.Kind() {

			case reflect.Struct:
				for i := 0; i < len; i++ {
					sliceElement := reflect.New(sliceType)
					inputConfig := reflect.ValueOf(sliceElement.Interface()).Elem()
					for j := 0; j < inputConfig.NumField(); j++ {
						generateValue(inputConfig.Field(j))
					}
					slice = reflect.Append(slice, sliceElement.Elem())

				}

			}
			return slice.Interface()

		},
	}
)

func generateValue(val reflect.Value) {
	switch val.Kind() {
	case reflect.String:
		val.SetString(GenerateStringFromRegexFunc("^[a-z ,.'-]+$").Generate().(string))
	}
}

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

func GenerateSlice(_typ reflect.Type, min, max int) *GenerationFunc {
	f := generateSlice
	f.Args = []any{_typ, min, max}
	return &f
}

// func GenerateFixedSlice(value any) *GenerationFunc {
// 	return
// }

// example.generatedFields[field.Name] = generator.GenerateSlice(sliceType, 0, 10)
// // TODO fixed
// example.generatedFields[field.Name] = generator.GenerateFixedSlice(example.fixedValues[field.Name])
