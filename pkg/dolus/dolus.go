package dolus

import (
	"context"
	"fmt"
	"net/http"

	"github.com/MartinSimango/dolus/pkg/example"
	"github.com/MartinSimango/dolus/pkg/schema"
	"github.com/fatih/color"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/labstack/echo/v4"
)

const (
	banner = `  
    ____            __                
   / __ \  ____    / / __  __   _____
  / / / / / __ \  / / / / / /  / ___/
 / /_/ / / /_/ / / / / /_/ /  (__  ) 
/_____/  \____/ /_/  \____/  /____/ %s
Go framework for creating customizable and extendable mock servers
%s

--------------------------------------------------------------------

`
	Version = "0.0.1"
	website = "https://github.com/MartinSimango/dolus"
)

func printBanner() {
	versionColor := color.New(color.FgGreen).SprintFunc()("v", Version)
	websiteColor := color.New(color.FgBlue).SprintFunc()(website)
	fmt.Printf(banner, versionColor, websiteColor)
}

type Dolus struct {
	OpenAPIspec        string
	HideBanner         bool
	HidePort           bool
	EchoServer         *echo.Echo
	ResponseRepository *ResponseRepository
}

type OperationResponse struct {
	Operation string
	Response  http.Response
}

type Paths map[string][]OperationResponse

func New() *Dolus {
	return &Dolus{
		HideBanner:         false,
		HidePort:           false,
		OpenAPIspec:        "openapi.yaml",
		ResponseRepository: NewResponseRepository(),
	}
}

func (d *Dolus) initHttpServer() {
	d.EchoServer = echo.New()
	d.EchoServer.HideBanner = true
	d.EchoServer.HidePort = d.HidePort
}

func (d *Dolus) startHttpServer(address string) error {
	d.initHttpServer()

	ctx := context.Background()
	loader := &openapi3.Loader{Context: ctx, IsExternalRefsAllowed: true}
	doc, err := loader.LoadFromFile(d.OpenAPIspec)
	if err != nil {
		return err
	}

	if err := doc.Validate(ctx); err != nil {
		return err
	}

	// for path find operations
	// for each operations get the response if no response throw error for path and operations
	// look at schema and build struct from that schema (will be recursive)
	// if no schema look for examples values if no examples throw error as can't produce response

	for path := range doc.Paths {
		for method, operation := range doc.Paths[path].Operations() {
			p := path
			m := method
			for code, ref := range operation.Responses {
				s := schema.New(path, method, code, ref, "application/json")
				d.ResponseRepository.Add(p, m, code, example.New(s))

			}
			d.EchoServer.Router().Add(m, p, func(ctx echo.Context) error {
				// c, _ := strconv.Atoi(code)
				return d.ResponseRepository.GetEchoResponse(p, m, ctx)
			})
		}

	}

	return d.EchoServer.Start(address)
}

func (d *Dolus) Start(address string) error {
	if !d.HideBanner {
		printBanner()
	}
	return d.startHttpServer(address)
}

func (d *Dolus) AddExpectation() {
}
