package generator

import (
	"math/rand"
	"reflect"

	"github.com/MartinSimango/dolus/pkg/dstruct"
	"github.com/takahiromiyamoto/go-xeger"
)

var (
	generateStringFromRegex GenerationFunc = GenerationFunc{
		_func: func(parameters ...any) any {
			regex := parameters[0].(string)
			x, err := xeger.NewXeger(regex)
			if err != nil {
				panic(err)
			}
			return x.Generate()
		},
	}

	generateInteger GenerationFunc = GenerationFunc{
		_func: func(parameters ...any) any {
			min := parameters[0].(int64)
			max := parameters[1].(int64)
			return min + (int64(rand.Float64() * float64(max+1-min)))
		},
	}

	generateFloat GenerationFunc = GenerationFunc{
		_func: func(parameters ...any) any {
			min := parameters[0].(float64)
			max := parameters[1].(float64)
			return min + (rand.Float64() * (max + 1 - min))
		},
	}

	generateFixedString GenerationFunc = GenerationFunc{
		_func: func(parameters ...any) any {
			return parameters[0].(string)
		},
	}

	generateObject GenerationFunc = GenerationFunc{
		_func: func(parameters ...any) any {
			return nil
		},
	}

	generateSlice GenerationFunc = GenerationFunc{
		_func: func(parameters ...any) any {
			sliceType := parameters[0].(reflect.Type)
			min := parameters[1].(int)
			max := parameters[2].(int)

			len := min + (int(rand.Float64() * float64(max+1-min)))
			sliceOfElementType := reflect.SliceOf(sliceType)
			slice := reflect.MakeSlice(sliceOfElementType, 0, 1024)
			switch sliceType.Kind() {
			case reflect.Struct:
				sliceElement := reflect.New(sliceType)
				for i := 0; i < len; i++ {
					generateStructValues(reflect.ValueOf(sliceElement.Interface()).Elem())
					slice = reflect.Append(slice, sliceElement.Elem())

				}
			}
			return slice.Interface()

		},
	}

	generateStruct GenerationFunc = GenerationFunc{

		_func: func(parameters ...any) any {
			val := parameters[0].(reflect.Value)
			generateStructValues(val)
			return val
		},
	}
)

func generateValue(val reflect.Value) {
	switch val.Kind() {
	case reflect.String:
		val.SetString(GenerateStringFromRegexFunc("^[a-z ,.'-]+$").Generate().(string))
	case reflect.Struct:
		generateStructValues(val)
	case reflect.Bool:
		generateBoolValue(val)
	}
}

func generateStructValues(config reflect.Value) {

	for j := 0; j < config.NumField(); j++ {
		generateValue(config.Field(j))
	}

}

func generateBoolValue(val reflect.Value) {
	val.SetBool(true)
}

func GenerateStringFromRegexFunc(regex string) *GenerationFunc {
	f := generateStringFromRegex
	f.args = []any{regex}
	return &f
}

func GenerateIntegerFunc(min, max int64) *GenerationFunc {
	f := generateInteger
	f.args = []any{min, max}
	return &f
}

func GenerateFloatFunc(min, max float64) *GenerationFunc {
	f := generateFloat
	f.args = []any{min, max}
	return &f
}

func GenerateFixedStringFunc(regex string) *GenerationFunc {
	f := generateFixedString
	f.args = []any{regex}
	return &f
}

func GenerateSlice(_typ reflect.Type, min, max int) *GenerationFunc {
	f := generateSlice
	f.args = []any{_typ, min, max}
	return &f
}

func GenerateStruct(val reflect.Value) *GenerationFunc {
	f := generateStruct
	f.args = []any{val}
	return &f
}

func GetGenerationFunction(field *dstruct.Field,
	functionValueConfig FunctionValueConfig) *GenerationFunc {
	pattern := field.Tags.Get("pattern")
	if pattern != "" {
		return GenerateStringFromRegexFunc(pattern)
	}
	switch field.Kind {
	case reflect.Float64:
		return GenerateFloatFunc(0, 10)
	case reflect.Slice:
		sliceType := reflect.TypeOf(field.Value.Interface()).Elem()
		return GenerateSlice(sliceType, 1, 2)
	case reflect.Struct:
		return GenerateStruct(field.Value)
	}

	return nil
}

// func GenerateFixedSlice(value any) *GenerationFunc {
// 	return
// }

// example.generatedFields[field.Name] = generator.GenerateSlice(sliceType, 0, 10)
// // TODO fixed
// example.generatedFields[field.Name] = generator.GenerateFixedSlice(example.fixedValues[field.Name])
