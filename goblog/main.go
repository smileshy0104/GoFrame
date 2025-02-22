package main

import (
	"fmt"
	"frame"
)

func main() {
	engine := frame.New()
	g := engine.Group("user")

	// 使用 Use 方法添加一个中间件，该中间件会在处理请求之前和之后分别执行一些操作。
	g.Use(func(next frame.HandlerFunc) frame.HandlerFunc {
		return func(ctx *frame.Context) {
			fmt.Println("pre handler")
			next(ctx)
			fmt.Println("post handler")
		}
	})

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
