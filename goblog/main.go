package main

import (
	"fmt"
	"frame"
)

func main() {
	engine := frame.New()
	g := engine.Group("user")
	g.Get("/hello", func(context *frame.Context) {
		fmt.Fprintln(context.W, "GET test test")
	})

	g.Post("/hello", func(context *frame.Context) {
		fmt.Fprintln(context.W, "POST test test")
	})

	g.Get("/hello/*/get", func(context *frame.Context) {
		fmt.Fprintln(context.W, "/hello/*/get test test")
	})

	engine.Run()
}
