package generator

import (
	"reflect"
)

func setValue(val reflect.Value, generationConfig GenerationConfig) {
	switch val.Kind() {
	case reflect.Struct:
		setStructValues(val, generationConfig)
	case reflect.Slice:
		panic("Unhanled setValue case")
	default:
		val.Set(reflect.ValueOf(generationConfig.DefaultGenerationFunctions[val.Kind()].Generate()))
		// val.SetString(GenerateStringFromRegexFunc("^[a-z ,.'-]+$").Generate().(string))
	}
}

func setStructValues(config reflect.Value, generationConfig GenerationConfig) {
	for j := 0; j < config.NumField(); j++ {
		setValue(config.Field(j), generationConfig)
	}

}
