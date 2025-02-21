package main

import (
	"fmt"
	"frame"
)

func main() {
	engine := frame.New()
	g := engine.Group("user")
	g.Get("/hello", func(context *frame.Context) {
		fmt.Fprintln(context.W, "hello test test")
	})
	engine.Run()
}
