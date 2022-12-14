package dolus

import "net/http"

// type NoExpectationType

type Expectation struct {
	Priority int
	Path     string
	Request  http.Request
	Response http.Response
	Schema   interface{}
	Example  interface{}
}

// TODO have different types of expe

// TODO server should have cap of when to send 429 if certain number of requests are coming through

// If No expectation
//   -- check if operation has example
//      return example
//   -- check if operation has schema
//      	check no expectation type (GENERATED or USED Default type values)
//   --
// -- if no schema return error (internal server with message saying response could not be given)
// -- Check type of expectation (GENERATERD tpyes)

// -- return 200 for any request for any Operation if no schema

// Request Operation Path -> 200
