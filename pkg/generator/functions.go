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

	generateBool GenerationFunc = GenerationFunc{
		_func: func(parameters ...any) any {
			return GenerateIntegerFunc(0, 1).Generate().(int64) == 0
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

	generateNilValue GenerationFunc = GenerationFunc{
		_func: func(parameters ...any) any {
			return nil
		},
	}

	generateStruct GenerationFunc = GenerationFunc{

		_func: func(parameters ...any) any {
			generationConfig := parameters[0].(GenerationConfig)
			val := parameters[1].(reflect.Value)
			setStructValues(val, generationConfig)
			return val
		},
	}

	generateSlice GenerationFunc = GenerationFunc{
		_func: func(parameters ...any) any {
			generationConfig := parameters[0].(GenerationConfig)

			val := parameters[1].(reflect.Value)
			sliceType := reflect.TypeOf(val.Interface()).Elem()
			min := generationConfig.SliceMinLength
			max := generationConfig.SliceMaxLength

			len := min + (int(rand.Float64() * float64(max+1-min)))
			sliceOfElementType := reflect.SliceOf(sliceType)
			slice := reflect.MakeSlice(sliceOfElementType, 0, 1024)
			switch sliceType.Kind() {
			case reflect.Struct:
				sliceElement := reflect.New(sliceType)
				for i := 0; i < len; i++ {
					setStructValues(reflect.ValueOf(sliceElement.Interface()).Elem(), generationConfig)
					slice = reflect.Append(slice, sliceElement.Elem())

				}
			}
			return slice.Interface()

		},
	}

	generatePointerValue GenerationFunc = GenerationFunc{

		_func: func(parameters ...any) any {
			generationConfig := parameters[0].(GenerationConfig)
			val := parameters[1].(reflect.Value)
			ptr := reflect.New(val.Type().Elem())
			setValue(ptr.Elem(), generationConfig)
			return ptr.Interface()

		},
	}
)

func GenerateStringFromRegexFunc(regex string) GenerationFunction {
	f := generateStringFromRegex
	f.args = []any{regex}
	return f
}

func GenerateIntegerFunc(min, max int64) GenerationFunction {
	f := generateInteger
	f.args = []any{min, max}
	return f
}

func GenerateFloatFunc(min, max float64) GenerationFunction {
	f := generateFloat
	f.args = []any{min, max}
	return f
}

func GenerateBoolFunc() GenerationFunction {
	f := generateBool
	return f
}

func GenerateFixedStringFunc(regex string) GenerationFunction {
	f := generateFixedString
	f.args = []any{regex}
	return f
}

func GenerateNilValue() GenerationFunction {
	f := generateNilValue
	return f
}

func GenerateSlice(generationConfig GenerationConfig, val reflect.Value) GenerationFunction {
	f := generateSlice
	f.args = []any{generationConfig, val}
	return f
}

func GenerateStruct(generationConfig GenerationConfig, val reflect.Value) GenerationFunction {
	f := generateStruct
	f.args = []any{generationConfig, val}
	return f
}

func GeneratePointerValue(generationConfig GenerationConfig, val reflect.Value) GenerationFunction {
	f := generatePointerValue
	f.args = []any{generationConfig, val}
	return f
}

func GetGenerationFunction(field *dstruct.Field,
	functionValueConfig GenerationConfig, // TODO add example config contains slice size
) GenerationFunction {

	pattern := field.Tags.Get("pattern")
	if pattern != "" {
		return GenerateStringFromRegexFunc(pattern)
	}

	switch field.Kind {
	case reflect.Slice:
		return GenerateSlice(functionValueConfig, field.Value)
	case reflect.Struct:
		return GenerateStruct(functionValueConfig, field.Value)
	case reflect.Ptr:
		if functionValueConfig.SetNonRequiredFields {
			return GeneratePointerValue(functionValueConfig, field.Value)
		}
	}
	return functionValueConfig.DefaultGenerationFunctions[field.Kind]

}
