package dolus

import (
	"context"
	"fmt"

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
	OpenAPIspec string
	HideBanner  bool
	HidePort    bool
	echoServer  *echo.Echo
}

func New() *Dolus {
	return &Dolus{
		HideBanner: false,
		HidePort:   false,
	}
}

func (d *Dolus) initHttpServer() {
	d.echoServer = echo.New()
	d.echoServer.HideBanner = true
	d.echoServer.HidePort = d.HidePort
}

func (d *Dolus) startHttpServer(port int) {
	d.initHttpServer()

	ctx := context.Background()
	loader := &openapi3.Loader{Context: ctx, IsExternalRefsAllowed: true}
	doc, err := loader.LoadFromFile("openapi.yaml")
	if err != nil {
		fmt.Println(err)
		return
	}
	// Validate document
	_ = doc.Validate(ctx)

	for path := range doc.Paths {
		for operation, v := range doc.Paths[path].Operations() {
			a, err := v.Responses.Get(200).Value.MarshalJSON()
			if err == nil {
				fmt.Println(string(a))
			}

			d.echoServer.Router().Add(operation, path, func(ctx echo.Context) error {
				return ctx.JSON(200, v.Responses.Get(200).Value)
			})
		}

	}

	d.echoServer.Start(fmt.Sprintf(":%d", port))
}

func (d *Dolus) Start() {
	if !d.HideBanner {
		printBanner()
	}
	d.startHttpServer(1080)

}

func (d *Dolus) AddExpectation() {
}
