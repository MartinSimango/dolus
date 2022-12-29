package dolus

import (
	"github.com/MartinSimango/dolus/pkg/example"
	"github.com/labstack/echo/v4"
)

type PathMethod struct {
	Path   string
	Method string
}

type B struct {
	Priority   int
	StatusCode string
	Request    any
	Example    *example.Example
	//Matcher for request
	// GenerateFields

}

type ResponseRepository struct {
	Expectations map[PathMethod][]B
}

func NewResponseRepository() *ResponseRepository {
	return &ResponseRepository{
		Expectations: make(map[PathMethod][]B),
	}
}

type GeneralError struct {
	Path     string
	Method   string
	ErrorMsg string
}

func (repo *ResponseRepository) GetEchoResponse(path, method string, ctx echo.Context) error {
	// need to look at request path, method
	// look at path and method - see if any expectations are there for

	// look for one with highest priority - wth path and method of which matches the request
	// -- of those look which one matches request if anything to match
	// -- look at the expectation examples and return

	for _, expectation := range repo.Expectations[PathMethod{Path: path, Method: method}] {

		if expectation.StatusCode == "200" {
			return ctx.JSON(200, expectation.Example.Get())

		}
	}
	return ctx.JSON(500, GeneralError{
		Path:     path,
		Method:   method,
		ErrorMsg: "No expectation found for path and HTTP method.",
	})

}

func (repo *ResponseRepository) Add(path, method, code string, example *example.Example) {
	if example == nil {
		return
	}
	pathMethod := PathMethod{
		Path:   path,
		Method: method,
	}

	repo.Expectations[pathMethod] = append(repo.Expectations[pathMethod],
		B{
			Priority:   0,
			StatusCode: code,
			Request:    nil,
			Example:    example})

}

// func (repo *ResponseRepository) GetResponse(operation, path string, ctx echo.Context) error {

// }
