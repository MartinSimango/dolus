package main

import (
	"fmt"

	"github.com/MartinSimango/dolus/pkg/dolus"
)

func main() {
	d := dolus.New()
	d.OpenAPIspec = "openapi2.yaml"
	if err := d.Start(fmt.Sprintf(":%d", 1080)); err != nil {
		fmt.Println(err)
	}
}
